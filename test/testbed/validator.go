package testbed

import (
	"fmt"
	"github.com/fatih/color"
)

var (
	errNoCPURecorded  = fmt.Errorf("No average CPU recorded")
	errNoMemRecorded  = fmt.Errorf("No average Memory RAM recorded")
	errCriticalAvgCPU = fmt.Errorf("Average CPU surpassed critical threshold")
	errCriticalAvgMem = fmt.Errorf("Average Memory surpassed critical threshold")
)

type ResourceValidator interface {
	// Validate a given ResouceConsumption
	Validate(consumption *ResourceConsumption) error
	// Reporty a ResouceConsumption
	Report(consumption *ResourceConsumption) string
}

type PerfValidator struct {
	// Accepted, warning or critical average CPU thresholds
	AcceptedCPU, WarnCPU, CriticalCPU float64
	// Accepted, warning or critical average Memory thresholds
	AcceptedMem, WarnMem, CriticalMem uint32
}

func NewDefaultPerfValidator() *PerfValidator {
	return &PerfValidator{
		AcceptedCPU: 2,
		WarnCPU:     3,
		CriticalCPU: 5,
		AcceptedMem: 50000000,
		WarnMem:     500000000,
		CriticalMem: 700000000,
	}
}

// validates average cpu and memory
func (p *PerfValidator) Validate(c *ResourceConsumption) error {
	if c.CPUPercentAvg <= 0 {
		return errNoCPURecorded
	} else if c.CPUPercentAvg >= p.CriticalCPU {
		return fmt.Errorf("Error: %w CPU: %d", errCriticalAvgCPU, c.CPUPercentAvg)
	}

	if c.RAMAvg <= 0 {
		return errCriticalAvgMem
	} else if c.RAMAvg >= p.CriticalMem {
		return fmt.Errorf("Error: %w Mem: %d", errCriticalAvgMem, c.RAMAvg)
	}
	return nil
}

func (p *PerfValidator) Report(c *ResourceConsumption) string {
	// TODO: add colors based on thresholds
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
	red := color.New(color.FgRed, color.Bold).SprintFunc()

	maxCPUString := yellow(fmt.Sprintf("%4.1f%%", c.CPUPercentMax))
	avgCPUString := green(fmt.Sprintf("%4.1f%%", c.CPUPercentAvg))
	maxMemString := red(fmt.Sprintf("%4d B", c.RAMMax))
	avgMemString := red(fmt.Sprintf("%4d B", c.RAMAvg))

	totalReadIOString := red(fmt.Sprintf("%4d B", c.ReadIOTotal))
	totalWriteIOString := red(fmt.Sprintf("%4d B", c.WriteIOTotal))

	output := "\n======= CPU \n" + fmt.Sprintf("Max CPU: %s\nAvg CPU: %s\n", maxCPUString, avgCPUString)
	output += "======== MEM \n" + fmt.Sprintf("Max RAM: %s\nAvg RAM: %s\n", maxMemString, avgMemString)
	output += "======== IO  \n" + fmt.Sprintf("Total Read: %s\nTotal Writes: %s", totalReadIOString, totalWriteIOString)
	return output
}
