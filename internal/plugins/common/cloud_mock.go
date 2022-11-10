// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/newrelic/infrastructure-agent/pkg/sysinfo/cloud"
	"github.com/stretchr/testify/mock"
)

type FakeCloudHarvester struct {
	mock.Mock
}

// GetInstanceID will return the id of the cloud instance.
func (f *FakeCloudHarvester) GetInstanceID() (string, error) {
	args := f.Called()
	return args.String(0), args.Error(1)
}

// GetHostType will return the cloud instance type.
func (f *FakeCloudHarvester) GetHostType() (string, error) {
	args := f.Called()
	return args.String(0), args.Error(1)
}

// GetCloudType will return the cloud type on which the instance is running.
func (f *FakeCloudHarvester) GetCloudType() cloud.Type {
	args := f.Called()
	return args.Get(0).(cloud.Type)
}

// Returns a string key which will be used as a HostSource (see host_aliases plugin).
func (f *FakeCloudHarvester) GetCloudSource() string {
	args := f.Called()
	return args.String(0)
}

// GetRegion returns the cloud region
func (f *FakeCloudHarvester) GetRegion() (string, error) {
	args := f.Called()
	return args.String(0), args.Error(1)
}

// GetZone returns the cloud zone (availability zone)
func (f *FakeCloudHarvester) GetZone() (string, error) {
	args := f.Called()
	return args.String(0), args.Error(1)
}

// GetAccount returns the cloud account ID
func (f *FakeCloudHarvester) GetAccountID() (string, error) {
	args := f.Called()
	return args.String(0), args.Error(1)
}

// GetImageID returns the cloud instance ID
func (f *FakeCloudHarvester) GetInstanceImageID() (string, error) {
	args := f.Called()
	return args.String(0), args.Error(1)
}

// GetHarvester returns instance of the Harvester detected (or instance of themselves)
func (f *FakeCloudHarvester) GetHarvester() (cloud.Harvester, error) {
	return f, nil
}
