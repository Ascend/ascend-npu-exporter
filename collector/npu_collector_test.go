//  Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// package collector for Prometheus
package collector

import (
	"github.com/patrickmn/go-cache"
	"github.com/prashantv/gostub"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"huawei.com/npu-exporter/dsmi"
	"os"
	"testing"
	"time"
)

const (
	cacheTime = 60 * time.Second
	timestamp = 1606402
	waitTime  = 2 * time.Second
)

// TestNewNpuCollector test method of NewNpuCollector
func TestNewNpuCollector(t *testing.T) {
	tests := []struct {
		mockFunc func(n *npuCollector, stop <-chan os.Signal)
		name     string
		path     string
	}{
		{
			name: "should return full list metrics when npuInfo not empty",
			path: "testdata/prometheus_metrics",
			mockFunc: func(n *npuCollector, stop <-chan os.Signal) {
				npuInfo := mockGetNPUInfo(nil)
				n.cache.Set(key, npuInfo, n.cacheTime)
			},
		},
		{
			name: "should return full list metrics when npuInfo is empty",
			path: "testdata/prometheus_metrics2",
			mockFunc: func(n *npuCollector, stop <-chan os.Signal) {
				var npuInfo []HuaWeiNPUCard
				n.cache.Set(key, npuInfo, n.cacheTime)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			excuteTestCollector(t, tt)
		})
	}
}

func excuteTestCollector(t *testing.T, tt struct {
	mockFunc func(n *npuCollector, stop <-chan os.Signal)
	name     string
	path     string
}) {
	startStub := gostub.Stub(&start, tt.mockFunc)
	defer startStub.Reset()
	var stopChan chan os.Signal
	c := NewNpuCollector(cacheTime, time.Second, stopChan)
	time.Sleep(1 * time.Second)
	r := prometheus.NewRegistry()
	r.MustRegister(c)
	defer r.Unregister(c)
	exp, err := os.Open(tt.path)
	defer exp.Close()
	if err != nil {
		t.Fatalf("test failes")
	}
	if err := testutil.CollectAndCompare(c, exp); err != nil {
		t.Fatal("Unexpected metrics returned:", err)
	}
}

// TestGetChipInfo test  method getChipInfo
func TestGetChipInfo(t *testing.T) {
	tests := []testCase{
		newTestCase("should return chip info successfully when dsmi works normally", false,
			dsmi.NewDeviceManagerMock()),
		newTestCase("should return nil when dsmi works abnormally", true, dsmi.NewDeviceManagerMockErr()),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chipInfo := packChipInfo(0, tt.mockPart.(dsmi.DeviceMgrInterface))
			t.Logf("%v", chipInfo)
			assert.NotNil(t, chipInfo)
			if tt.wantErr {
				assert.Equal(t, "", chipInfo.ChipIfo.ChipName)
			} else {
				assert.NotNil(t, chipInfo.ChipIfo)
			}
		})
	}
}

// TestGetHealthCode test getHealthCode
func TestGetHealthCode(t *testing.T) {
	tests := []struct {
		name   string
		health HealthEnum
		want   int
	}{
		{
			name:   "should return 1 when given Healthy",
			health: Healthy,
			want:   1,
		},
		{
			name:   "should return 0 when given UnHealthy",
			health: UnHealthy,
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHealthCode(tt.health); got != tt.want {
				t.Errorf("getHealthCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetNPUInfo test method of getNPUInfo
func TestGetNPUInfo(t *testing.T) {
	tests := []struct {
		name string
		args dsmi.DeviceMgrInterface
		want []HuaWeiNPUCard
	}{
		{
			name: "should return at lease one NPUInfo",
			args: dsmi.NewDeviceManagerMock(),
			want: []HuaWeiNPUCard{{
				DeviceList: nil,
				Timestamp:  time.Time{},
				CardID:     0,
			}},
		},
		{
			name: "should return zero NPU",
			args: dsmi.NewDeviceManagerMockErr(),
			want: []HuaWeiNPUCard{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNPUInfo(tt.args); len(got) != len(tt.want) {
				t.Errorf("getNPUInfo() = %v,want %v", got, tt.want)
			}
			if got := assembleNPUInfoV1(tt.args); len(got) != len(tt.want) {
				t.Errorf("assembleNPUInfoV1() = %v,want %v", got, tt.want)
			}
		})
	}
}

type testCase struct {
	name        string
	wantErr     bool
	mockPart    interface{}
	expectValue interface{}
	expectCount interface{}
}

func newTestCase(name string, wantErr bool, mockPart interface{}) testCase {
	return testCase{
		name:     name,
		wantErr:  wantErr,
		mockPart: mockPart,
	}
}

func mockGetNPUInfo(dmgr dsmi.DeviceMgrInterface) []HuaWeiNPUCard {
	var npuList []HuaWeiNPUCard
	for phyID := int32(0); phyID < 8; phyID++ {
		chipInfo := &HuaWeiAIChip{
			HealthStatus: Healthy,
			ErrorCode:    0,
			Utilization:  0,
			Temperature:  0,
			Power:        0,
			Voltage:      0,
			Frequency:    0,
			Meminf: &dsmi.MemoryInfo{
				MemorySize:  0,
				Frequency:   0,
				Utilization: 0,
			},
			ChipIfo: &dsmi.ChipInfo{
				ChipType: "Ascend",
				ChipName: "910Awn",
				ChipVer:  "V1",
			},
			HbmInfo: &dsmi.HbmInfo{
				MemorySize:              0,
				MemoryFrequency:         0,
				MemoryUsage:             0,
				MemoryTemp:              0,
				MemoryBandWidthUtilRate: 0,
			},
		}
		chipInfo.DeviceID = int(phyID)
		npuCard := HuaWeiNPUCard{
			CardID:     int(phyID),
			DeviceList: []*HuaWeiAIChip{chipInfo},
			Timestamp:  time.Unix(timestamp, 0),
		}
		npuList = append(npuList, npuCard)
	}
	return npuList
}

// TestStart test start method
func TestStart(t *testing.T) {
	ch := make(chan os.Signal)
	tests := []struct {
		collector *npuCollector
		name      string
	}{
		{
			name: "should set cache successfully",
			collector: &npuCollector{
				cache:      cache.New(cacheTime, five*time.Minute),
				cacheTime:  cacheTime,
				updateTime: time.Second,
			},
		},
	}
	gostub.Stub(&getNPUInfo, mockGetNPUInfo)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go start(tt.collector, ch)
			time.Sleep(waitTime)
			objm, ok := tt.collector.cache.Get(key)
			assert.NotNil(t, objm)
			assert.Equal(t, true, ok)
			go func() {
				ch <- os.Interrupt
				close(ch)
			}()
		})
	}
}
