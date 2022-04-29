package testbed

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/process"
	"go.uber.org/atomic"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

var (
	errNoAgentRunning = fmt.Errorf("No Infra Agent running")
)

type ResourceConsumption struct {
	CPUPercentAvg float64
	CPUPercentMax float64
	// Average Memory in Bytes
	RAMAvg uint32
	RAMMax uint32
	// IO Counters in Bytes
	ReadIOTotal  uint64
	WriteIOTotal uint64
}

type InfraRunner interface {
	Start() error
	Stop() error
	WatchResourceConsumption() error
	GetResourceConsumption() *ResourceConsumption
}

// childInfraRunner implements the InfraRunner interface as a child process on the same machine
type childInfraRunner struct {
	// Path to agent executable
	BinPath string

	// Configuration file
	ConfigFile string

	// LogFile captures stdout and stderr
	LogFile string

	// Command to execute
	cmd *exec.Cmd

	isStarted  bool
	stopOnce   sync.Once
	isStopped  bool
	doneSignal chan struct{}

	// Gopsutil process representation
	pMetrics *process.Process

	// Start time and elapsed time since last metrics fetch
	startTime       time.Time
	lastElapsedTime time.Time

	// Process CPU times that were fetched on last monitoring tick.
	lastProcessTimes *cpu.TimesStat

	// Current resident set size RAM in Bytes
	ramCur atomic.Uint32

	// Current IO Read Bytes
	ioReadCur atomic.Uint64

	// Current IO Write Bytes
	ioWriteCur atomic.Uint64

	// Current CPU percentage times 1000
	cpuPercentCur atomic.Uint32

	// Maximum CPU seen
	cpuPercentMax float64

	// Number of memory measurements
	memProbeCount int

	// Cumulative RAM RSS in Bytes
	ramTotal uint64

	// Maximum RAM seen
	RAMMax uint32
}

func NewChildInfraRunner(binPath, configPath, logPath string) *childInfraRunner {
	return &childInfraRunner{
		BinPath:    binPath,
		ConfigFile: configPath,
		LogFile:    logPath,
		doneSignal: make(chan struct{}),
	}
}

func (cr *childInfraRunner) Start() error {

	// Prepare log file
	logFile, err := os.Open(cr.LogFile)
	if err != nil {
		return fmt.Errorf("cannot create %s: %s", cr.LogFile, err.Error())
	}

	log.Printf("Starting Infra Agent (%s)", cr.BinPath)

	// prepare command
	args := []string{"--config", cr.ConfigFile}
	cr.cmd = exec.Command(cr.BinPath, args...)

	// Capture standard output and standard error.
	cr.cmd.Stdout = logFile
	cr.cmd.Stderr = logFile

	// Start the process.
	if err = cr.cmd.Start(); err != nil {
		return fmt.Errorf("cannot start executable at %s: %s", cr.BinPath, err.Error())
	}

	cr.startTime = time.Now()
	cr.isStarted = true

	log.Printf("Infra Agent running, pid=%d, logfile=%s\n", cr.cmd.Process.Pid, cr.LogFile)

	return nil
}

func (cr *childInfraRunner) Stop() (err error) {
	if !cr.isStarted || cr.isStopped {
		return errNoAgentRunning
	}
	cr.stopOnce.Do(func() {
		cr.isStopped = true
		close(cr.doneSignal)

		log.Printf("Gracefully terminating Infra Agent pid=%d, sending SIGTEM...", cr.cmd.Process.Pid)

		// Gracefully signal process to stop.
		if err := cr.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Printf("Cannot send SIGTEM: %s", err.Error())
		}

		finished := make(chan struct{})

		// Setup a goroutine to wait a while for process to finish and send kill signal
		// to the process if it doesn't finish.
		go func() {
			// Wait 15 seconds.
			t := time.After(15 * time.Second)
			select {
			case <-t:
				log.Printf("Infra Agent pid=%d is not responding to SIGTERM. Sending SIGKILL to kill forcedly.",
					cr.cmd.Process.Pid)
				if err = cr.cmd.Process.Signal(syscall.SIGKILL); err != nil {
					log.Printf("Cannot send SIGKILL: %s", err.Error())
				}
			case <-finished:
			}
		}()

		// Wait for process to terminate
		err = cr.cmd.Wait()

		// Let goroutine know process is finished.
		close(finished)

		// Set resource consumption stats to 0
		cr.ramCur.Store(0)
		cr.cpuPercentCur.Store(0)
		cr.ioReadCur.Store(0)
		cr.ioWriteCur.Store(0)
	})
	return
}

func (cr *childInfraRunner) WatchResourceConsumption() (err error) {
	cr.pMetrics, err = process.NewProcess(int32(cr.cmd.Process.Pid))
	if err != nil {
		return
	}

	cr.lastElapsedTime = time.Now()
	cr.lastProcessTimes, err = cr.pMetrics.Times()
	if err != nil {
		return
	}

	// when the agent start a lot of CPU is consumed to initialize all the processes
	log.Println("Sleeping 20 seconds to prevent star up extensive resource consumption")
	time.Sleep(20 * time.Second)

	// Measure every 10 seconds.
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cr.fetchRAMUsage()
			cr.fetchIOUsage()
			cr.fetchCPUUsage()

		case <-cr.doneSignal:
			log.Println("Stopping watching resource consumption")
			return nil
		}
	}
}

func (cr *childInfraRunner) GetResourceConsumption() *ResourceConsumption {
	rc := &ResourceConsumption{}

	if cr.pMetrics != nil {
		// Get total elapsed time since process start
		elapsedDuration := cr.lastElapsedTime.Sub(cr.startTime).Seconds()

		if elapsedDuration > 0 {
			// Calculate average CPU usage since start of process
			rc.CPUPercentAvg = cr.lastProcessTimes.Total() / elapsedDuration * 100.0
		}
		rc.CPUPercentMax = cr.cpuPercentMax

		if cr.memProbeCount > 0 {
			// Calculate average RAM usage by averaging all RAM measurements
			rc.RAMAvg = uint32(cr.ramTotal / uint64(cr.memProbeCount))
		}
		rc.RAMMax = cr.RAMMax

		rc.ReadIOTotal = cr.ioReadCur.Load()
		rc.WriteIOTotal = cr.ioWriteCur.Load()
	}

	return rc
}

func (cr *childInfraRunner) fetchRAMUsage() {
	// Get process memory and crU times
	mi, err := cr.pMetrics.MemoryInfo()
	if err != nil {
		log.Printf("cannot get process memory for %d: %s",
			cr.cmd.Process.Pid, err.Error())
		return
	}

	// Calculate RSS in Bytes.
	ramCur := uint32(mi.RSS)

	// Calculate aggregates.
	cr.memProbeCount++
	cr.ramTotal += uint64(ramCur)
	if ramCur > cr.RAMMax {
		cr.RAMMax = ramCur
	}

	// Store current usage.
	cr.ramCur.Store(ramCur)
}

func (cr *childInfraRunner) fetchIOUsage() {
	io, err := cr.pMetrics.IOCounters()
	if err != nil {
		log.Printf("cannot get process io counters for %d: %s",
			cr.cmd.Process.Pid, err.Error())
		return
	}

	// Store current usage.
	cr.ioReadCur.Store(io.ReadBytes)
	cr.ioWriteCur.Store(io.WriteBytes)
}

func (cr *childInfraRunner) fetchCPUUsage() {
	times, err := cr.pMetrics.Times()
	if err != nil {
		log.Printf("cannot get process times for %d: %s",
			cr.cmd.Process.Pid, err.Error())
		return
	}

	now := time.Now()

	// Calculate elapsed and process CPU time deltas in seconds
	deltaElapsedTime := now.Sub(cr.lastElapsedTime).Seconds()
	deltaCPUTime := times.Total() - cr.lastProcessTimes.Total()
	if deltaCPUTime < 0 {
		// We sometimes get negative difference when the process is terminated.
		deltaCPUTime = 0
	}

	cr.lastProcessTimes = times
	cr.lastElapsedTime = now

	// Calculate CPU usage percentage in elapsed period.
	cpuPercent := deltaCPUTime * 100 / deltaElapsedTime
	if cpuPercent > cr.cpuPercentMax {
		cr.cpuPercentMax = cpuPercent
	}

	curCPUPercentageX1000 := uint32(cpuPercent * 1000)

	// Store current usage.
	cr.cpuPercentCur.Store(curCPUPercentageX1000)
}
