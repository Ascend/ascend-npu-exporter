/* Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package collector for Prometheus
package collector

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	"huawei.com/npu-exporter/v5/collector/container"
	"huawei.com/npu-exporter/v5/collector/container/v1"
	"huawei.com/npu-exporter/v5/common-utils/cache"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/devmanager"
	"huawei.com/npu-exporter/v5/devmanager/common"
)

const (
	cacheTime = 60 * time.Second
	timestamp = 1606402
	waitTime  = 2 * time.Second
	npuCount  = 8
)

type mockContainerRuntimeOperator struct{}

// Init implements ContainerRuntimeOperator
func (operator *mockContainerRuntimeOperator) Init() error {
	return nil
}

// Close implements ContainerRuntimeOperator
func (operator *mockContainerRuntimeOperator) Close() error {
	return nil
}

// ContainerIDs implements ContainerRuntimeOperator
func (operator *mockContainerRuntimeOperator) GetContainers(ctx context.Context) ([]*v1alpha2.Container, error) {
	return []*v1alpha2.Container{}, nil
}

// GetContainerInfoByID implements ContainerRuntimeOperator
func (operator *mockContainerRuntimeOperator) GetContainerInfoByID(ctx context.Context, id string) (v1.Spec, error) {
	return v1.Spec{}, nil
}

func mockScan4AscendDevices(_ string) ([]int, bool, error) {
	return []int{1}, true, nil
}

func mockGetCgroupPath(controller, specCgroupsPath string) (string, error) {
	return "", nil
}

func makeMockDevicesParser() *container.DevicesParser {
	return &container.DevicesParser{
		RuntimeOperator: new(mockContainerRuntimeOperator),
	}
}

// TestNewNpuCollector test method of NewNpuCollector
func TestNewNpuCollector(t *testing.T) {
	tests := []struct {
		mockFunc func(ctx context.Context, n *npuCollector, dmgr devmanager.DeviceInterface)
		name     string
		path     string
	}{
		{
			name: "should return full list metrics when npuInfo not empty",
			path: "testdata/prometheus_metrics",
			mockFunc: func(ctx context.Context, n *npuCollector, dmgr devmanager.DeviceInterface) {
				_ = n.devicesParser.Init()
				npuInfo := mockGetNPUInfo(nil)
				if err := n.cache.Set(key, npuInfo, n.cacheTime); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "should return full list metrics when npuInfo is empty",
			path: "testdata/prometheus_metrics2",
			mockFunc: func(ctx context.Context, n *npuCollector, dmgr devmanager.DeviceInterface) {
				_ = n.devicesParser.Init()
				var npuInfo []HuaWeiNPUCard
				if err := n.cache.Set(key, npuInfo, n.cacheTime); err != nil {
					t.Fatal(err)
				}
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
	mockFunc func(ctx context.Context, n *npuCollector, dmgr devmanager.DeviceInterface)
	name     string
	path     string
}) {
	startStub := gomonkey.ApplyFunc(start, tt.mockFunc)
	defer startStub.Reset()
	patch := gomonkey.ApplyFunc(devmanager.AutoInit, func(s string) (*devmanager.DeviceManager, error) {
		return &devmanager.DeviceManager{}, nil
	})
	defer patch.Reset()
	c, err := NewNpuCollector(context.Background(), cacheTime, time.Second, makeMockDevicesParser())
	if err != nil {
		t.Fatalf("test failes")
	}
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
			&devmanager.DeviceManagerMock{}),
		newTestCase("should return nil when dsmi works abnormally", true, &devmanager.DeviceManagerMockErr{}),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chipInfo := packChipInfo(0, tt.mockPart.(devmanager.DeviceInterface))
			t.Logf("%#v", chipInfo)
			assert.NotNil(t, chipInfo)
			if tt.wantErr {
				assert.Equal(t, "", chipInfo.ChipIfo.Name)
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
				t.Errorf("getHealthCode() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

// TestGetNPUInfo test method of getNPUInfo
func TestGetNPUInfo(t *testing.T) {
	tests := []struct {
		name string
		args devmanager.DeviceInterface
		want []HuaWeiNPUCard
	}{
		{
			name: "should return at lease one NPUInfo",
			args: &devmanager.DeviceManagerMock{},
			want: []HuaWeiNPUCard{{
				DeviceList: nil,
				Timestamp:  time.Time{},
				CardID:     0,
			}},
		},
		{
			name: "should return zero NPU",
			args: &devmanager.DeviceManagerMockErr{},
			want: []HuaWeiNPUCard{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNPUInfo(tt.args); len(got) != len(tt.want) {
				t.Errorf("getNPUInfo() = %#v,want %#v", got, tt.want)
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

func mockGetNPUInfo(dmgr devmanager.DeviceInterface) []HuaWeiNPUCard {
	var npuList []HuaWeiNPUCard
	for devicePhysicID := int32(0); devicePhysicID < npuCount; devicePhysicID++ {
		chipInfo := &HuaWeiAIChip{
			HealthStatus: Healthy,
			ErrorCode:    0,
			Utilization:  0,
			Temperature:  0,
			Power:        0,
			Voltage:      0,
			Frequency:    0,
			Meminf: &common.MemoryInfo{
				MemorySize:  0,
				Frequency:   0,
				Utilization: 0,
			},
			ChipIfo: &common.ChipInfo{
				Type:    "Ascend",
				Name:    "910Awn",
				Version: "V1",
			},
			HbmInfo: &common.HbmInfo{
				MemorySize:        0,
				Frequency:         0,
				Usage:             0,
				Temp:              0,
				BandWidthUtilRate: 0,
			},
		}
		chipInfo.DeviceID = int(devicePhysicID)
		npuCard := HuaWeiNPUCard{
			CardID:     int(devicePhysicID),
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
				cache:         cache.New(cacheSize),
				cacheTime:     cacheTime,
				updateTime:    time.Second,
				devicesParser: makeMockDevicesParser(),
			},
		},
	}
	mk := gomonkey.ApplyFunc(getNPUInfo, mockGetNPUInfo)
	defer mk.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go start(context.Background(), tt.collector, &devmanager.DeviceManagerMock{})
			time.Sleep(waitTime)
			objm, err := tt.collector.cache.Get(key)
			assert.NotNil(t, objm)
			assert.Nil(t, err)
			go func() {
				ch <- os.Interrupt
				close(ch)
			}()
		})
	}
}

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&config, nil)
	gomonkey.ApplyFunc(container.ScanForAscendDevices, mockScan4AscendDevices)
	gomonkey.ApplyFunc(container.GetCgroupPath, mockGetCgroupPath)
}
