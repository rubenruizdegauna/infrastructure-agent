package scenaios

import (
	"github.com/newrelic/infrastructure-agent/test/testbed"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func tmpFile(data string) (file *os.File, err error) {
	file, err = ioutil.TempFile("", "")
	if err != nil {
		return
	}
	_, err = file.Write([]byte(data))
	file.Close()
	return
}

func TestSimpleMode(t *testing.T) {
	simpleConfig := `
license_key:  
staging: true
enable_process_metrics: false
`
	tmpConfig, err := tmpFile(simpleConfig)
	if err != nil {
		t.Error(err)
	}
	tmpLog, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		os.Remove(tmpConfig.Name())
		os.Remove(tmpLog.Name())
	}()

	tc := testbed.NewTestCase(t, testbed.NewChildInfraRunner("/usr/bin/newrelic-infra", tmpConfig.Name(), tmpLog.Name()), testbed.NewDefaultPerfValidator())
	tc.StartInfraAgent()
	tc.Sleep(240 * time.Second)
	tc.StopInfraAgent()
}

func TestProcessSampleMode(t *testing.T) {
	simpleConfig := `
license_key:  
staging: true
enable_process_metrics: true
metrics_process_sample_rate: 10
`
	tmpConfig, err := tmpFile(simpleConfig)
	if err != nil {
		t.Error(err)
	}
	tmpLog, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		os.Remove(tmpConfig.Name())
		os.Remove(tmpLog.Name())
	}()

	tc := testbed.NewTestCase(t, testbed.NewChildInfraRunner("/usr/bin/newrelic-infra", tmpConfig.Name(), tmpLog.Name()), testbed.NewDefaultPerfValidator())
	tc.StartInfraAgent()
	tc.Sleep(240 * time.Second)
	tc.StopInfraAgent()
}
