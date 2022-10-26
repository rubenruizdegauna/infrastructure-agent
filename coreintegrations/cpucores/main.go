package main

import (
	"log"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v4/integration"
	"github.com/shirou/gopsutil/cpu"
)

const (
	integrationVersion = "0.0.1"
)

func main() {
	cpuTimes, err := cpu.Times(true)
	if err != nil {
		log.Fatal(err)
	}

	i, err := integration.New("cpucores", integrationVersion)
	if err != nil {
		log.Fatal(err)
	}

	// attach entity to host
	i.HostEntity.SetIgnoreEntity(false)

	for _, cpu := range cpuTimes {
		m, err := integration.Gauge(time.Now(), "host.cpucore.usage.user", cpu.User)
		if err != nil {
			log.Fatal(err)
		}
		m.AddDimension("cpu", cpu.CPU)

		// add dimensions
		i.HostEntity.AddMetric(m)
	}

	i.Publish()
}
