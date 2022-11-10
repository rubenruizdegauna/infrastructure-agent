// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
//go:build darwin
// +build darwin

package darwin

import (
	"github.com/newrelic/infrastructure-agent/internal/plugins/common"
	testing2 "github.com/newrelic/infrastructure-agent/internal/plugins/testing"
	"github.com/newrelic/infrastructure-agent/pkg/sysinfo/cloud"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	. "gopkg.in/check.v1"
)

func TestGetKernelRelease_Success(t *testing.T) {
	hip := HostinfoPlugin{
		readDataFromCmd: func(cmd string, args ...string) (string, error) {
			return "Darwin Kernel Version 20.3.0: Thu Jan 21 00:07:06 PST 2021; root:xnu-7195.81.3~1/RELEASE_X86_64", nil
		},
	}
	actual, err := hip.getKernelRelease()

	assert.NoError(t, err)
	assert.Equal(t, "20.3.0", actual)
}

func TestGetKernelRelease_Error(t *testing.T) {
	hip := HostinfoPlugin{
		readDataFromCmd: func(cmd string, args ...string) (string, error) {
			return "", errors.New("error")
		},
	}
	actual, err := hip.getKernelRelease()

	assert.Error(t, err)
	assert.Equal(t, "", actual)
}

func TestGetKernelRelease_BadFormat(t *testing.T) {
	hip := HostinfoPlugin{
		readDataFromCmd: func(cmd string, args ...string) (string, error) {
			return "root:xnu-7195.81.3~1/RELEASE_X86_64", nil
		},
	}
	actual, err := hip.getKernelRelease()

	assert.Error(t, err)
	assert.Equal(t, "", actual)
}

func TestMemoryToKb(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "uppercase_correct_format",
			input:    "16 GB",
			expected: "16777216 kB",
			hasError: false,
		},
		{
			name:     "lowercase_correct_format",
			input:    "16 gB",
			expected: "16777216 kB",
			hasError: false,
		},
		{
			name:     "wrong_format",
			input:    "1aaas",
			expected: "",
			hasError: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			actual, err := memoryToKb(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestValidateHardwareUUID(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "uppercase_correct_format",
			input:    "C3805006-DFCF-11EB-BA80-0242AC130004",
			expected: true,
		},
		{
			name:     "lowercase_wrong_format",
			input:    "c3805006-dfcf-11eb-ba80-0242ac130004",
			expected: false,
		},
		{
			name:     "wrong_format",
			input:    "err",
			expected: false,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, validateHardwareUUID(test.input))
		})
	}
}

type HostinfoSuite struct {
	agent *testing2.MockAgent
}

var _ = Suite(&HostinfoSuite{})

var osRelease string

func (s *HostinfoSuite) SetUpSuite(c *C) {
}

func (s *HostinfoSuite) TearDownSuite(c *C) {
}

func (s *HostinfoSuite) SetUpTest(c *C) {
	s.agent = testing2.NewMockAgent()
}

func (s *HostinfoSuite) TestGatherHostInfo(c *C) {
	cloudHarvester := common.FakeCloudHarvester{}
	cloudHarvester.On("GetCloudType").Return(cloud.TypeAWS)
	cloudHarvester.On("GetRegion").Return("us-east-1", nil)
	cloudHarvester.On("GetZone").Return("us-east-1a", nil)
	cloudHarvester.On("GetInstanceImageID").Return("ami-12345", nil)
	cloudHarvester.On("GetAccountID").Return("x123", nil)

	hip := HostinfoPlugin{
		readDataFromCmd: func(cmd string, args ...string) (string, error) {
			return "", errors.New("error")
		},
		cloudHarvester: &common.FakeCloudHarvester{},
	}

	actual := hip.gatherHostinfo(s.agent)

	c.Assert(reflect.ValueOf(actual.AwsCloudData).IsZero(), Equals, false)
	c.Assert(actual.RegionAWS, Equals, "us-east-1")
}
