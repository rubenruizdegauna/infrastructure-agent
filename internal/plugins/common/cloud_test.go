// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/newrelic/infrastructure-agent/pkg/sysinfo/cloud"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCloudData(t *testing.T) {
	testCases := []struct {
		name       string
		assertions func(data *HostInfoData)
		setMock    func(*FakeCloudHarvester)
	}{
		{
			name: "no cloud",
			assertions: func(d *HostInfoData) {
				assert.Equal(t, "", d.RegionAWS)
				assert.Equal(t, "", d.RegionAzure)
				assert.Equal(t, "", d.RegionGCP)
				assert.Equal(t, "", d.RegionAlibaba)
			},
			setMock: func(h *FakeCloudHarvester) {
				h.On("GetCloudType").Return(cloud.TypeNoCloud)
			},
		},
		{
			name: "cloud aws",
			assertions: func(d *HostInfoData) {
				assert.Equal(t, "us-east-1", d.RegionAWS)
				assert.Equal(t, "us-east-1a", d.AWSAvailabilityZone)
				assert.Equal(t, "ami-12345", d.AWSImageID)
				assert.Equal(t, "x123", d.AWSAccountID)
				assert.Equal(t, "", d.RegionAzure)
				assert.Equal(t, "", d.RegionGCP)
				assert.Equal(t, "", d.RegionAlibaba)
			},
			setMock: func(h *FakeCloudHarvester) {
				h.On("GetCloudType").Return(cloud.TypeAWS)
				h.On("GetRegion").Return("us-east-1", nil)
				h.On("GetZone").Return("us-east-1a", nil)
				h.On("GetInstanceImageID").Return("ami-12345", nil)
				h.On("GetAccountID").Return("x123", nil)
			},
		},
		{
			name: "cloud azure",
			assertions: func(d *HostInfoData) {
				assert.Equal(t, "", d.RegionAWS)
				assert.Equal(t, "northeurope", d.RegionAzure)
				assert.Equal(t, "", d.RegionGCP)
				assert.Equal(t, "", d.RegionAlibaba)
				assert.Equal(t, "1", d.AzureAvailabilityZone)
				assert.Equal(t, "x123", d.AzureSubscriptionID)
			},
			setMock: func(h *FakeCloudHarvester) {
				h.On("GetAccountID").Return("x123", nil)
				h.On("GetCloudType").Return(cloud.TypeAzure)
				h.On("GetRegion").Return("northeurope", nil)
				h.On("GetZone").Return("1", nil)
			},
		},
		{
			name: "cloud gcp",
			assertions: func(d *HostInfoData) {
				assert.Equal(t, "", d.RegionAWS)
				assert.Equal(t, "", d.RegionAzure)
				assert.Equal(t, "us-east-1", d.RegionGCP)
				assert.Equal(t, "", d.RegionAlibaba)
			},
			setMock: func(h *FakeCloudHarvester) {
				h.On("GetCloudType").Return(cloud.TypeGCP)
				h.On("GetRegion").Return("us-east-1", nil)
			},
		},
		{
			name: "cloud alibaba",
			assertions: func(d *HostInfoData) {
				assert.Equal(t, "", d.RegionAWS)
				assert.Equal(t, "", d.RegionAzure)
				assert.Equal(t, "", d.RegionGCP)
				assert.Equal(t, "us-east-1", d.RegionAlibaba)
			},
			setMock: func(h *FakeCloudHarvester) {
				h.On("GetCloudType").Return(cloud.TypeAlibaba)
				h.On("GetRegion").Return("us-east-1", nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			h := new(FakeCloudHarvester)
			testCase.setMock(h)
			data := &HostInfoData{}
			cloudData, err := GetCloudData(h)
			assert.NoError(t, err)
			data.CloudData = cloudData
			testCase.assertions(data)
			h.AssertExpectations(t)
		})
	}
}
