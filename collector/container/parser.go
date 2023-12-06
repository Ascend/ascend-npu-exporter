/* Copyright(C) 2021-2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package container for monitoring containers' npu allocation
package container

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"huawei.com/npu-exporter/v5/collector/container/isula"
	"huawei.com/npu-exporter/v5/collector/container/v1"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/common-utils/utils"
)

const (
	namespaceMoby     = "moby"   // Docker
	namespaceK8s      = "k8s.io" // CRI + Containerd
	sliceLen8         = 8
	ascendDeviceInfo  = "ASCEND_VISIBLE_DEVICES"
	ascendEnvPart     = 2
	charDevice        = "c"
	devicePathPattern = `^/dev/davinci\d+$`
)

const (
	// EndpointTypeContainerd K8S + Containerd
	EndpointTypeContainerd = iota
	// EndpointTypeDockerd Docker with or without K8S
	EndpointTypeDockerd
	EndpointTypeIsula = 2
)

var (
	// ErrFromContext error is from the context
	ErrFromContext = errors.New("error from context")

	npuMajorID               []string
	npuMajorFetchCtrl        sync.Once
	parsingNpuDefaultTimeout = 3 * time.Second
)

// CntNpuMonitorOpts contains setting options for monitoring containers
type CntNpuMonitorOpts struct {
	CriEndpoint  string // CRI server address
	EndpointType int    // containerd or docker
	OciEndpoint  string // OCI server, now is containerd address
	UserBackUp   bool   // whether try to use backup address
}

// MakeDevicesParser evaluates option settings and make an instance according to it
func MakeDevicesParser(opts CntNpuMonitorOpts) *DevicesParser {
	runtimeOperator := &RuntimeOperatorTool{UseBackup: opts.UserBackUp}
	parser := &DevicesParser{}

	switch opts.EndpointType {
	case EndpointTypeContainerd:
		runtimeOperator.Namespace = namespaceK8s
		runtimeOperator.CriEndpoint = opts.CriEndpoint
		runtimeOperator.OciEndpoint = opts.OciEndpoint
		parser.RuntimeOperator = runtimeOperator
	case EndpointTypeDockerd:
		runtimeOperator.Namespace = namespaceMoby
		parser.RuntimeOperator = runtimeOperator
		runtimeOperator.CriEndpoint = opts.CriEndpoint
		runtimeOperator.OciEndpoint = opts.OciEndpoint
	case EndpointTypeIsula:
		runtimeOperator.Namespace = namespaceK8s
		parser.RuntimeOperator = runtimeOperator
		runtimeOperator.CriEndpoint = opts.CriEndpoint
		runtimeOperator.OciEndpoint = opts.OciEndpoint
	default:
		hwlog.RunLog.Errorf("Invalid type value %d", opts.EndpointType)
	}

	return parser
}

// DevicesInfo the container device information struct
type DevicesInfo struct {
	// container id
	ID string
	// container name, the format is: PodNameSpace_PodName_ContainerName
	Name    string
	Devices []int
}

// DevicesInfos the device information storage map
type DevicesInfos = map[string]DevicesInfo

// DevicesParser the parser which parse device info
type DevicesParser struct {
	// instances
	result chan DevicesInfos
	err    chan error
	// configuration
	RuntimeOperator RuntimeOperator
	Timeout         time.Duration
}

// Init initializes connection to containerd daemon and to CRI server or dockerd daemon based on name fetcher setting
func (dp *DevicesParser) Init() error {
	if err := dp.RuntimeOperator.Init(); err != nil {
		return contactError(err, "connecting to container runtime failed")
	}
	dp.result = make(chan DevicesInfos, 1)
	dp.err = make(chan error, 1)
	return nil
}

// RecvResult exposes the channel used for receiving devices info analyzing result
func (dp *DevicesParser) RecvResult() <-chan DevicesInfos {
	return dp.result
}

// RecvErr exposes the channel used for receiving errors occurred during analyzing
func (dp *DevicesParser) RecvErr() <-chan error {
	return dp.err
}

// Close closes all connections and channels established during initializing
func (dp *DevicesParser) Close() {
	_ = dp.RuntimeOperator.Close()
}

func (dp *DevicesParser) parseDevices(ctx context.Context, c *CommonContainer, rs chan<- DevicesInfo) error {
	if dp.RuntimeOperator.GetContainerType() == IsulaContainer {
		return dp.parseDeviceInIsula(ctx, c, rs)
	}

	return dp.parseDevicesInContainerd(ctx, c, rs)
}

func (dp *DevicesParser) parseDevicesInContainerd(ctx context.Context, c *CommonContainer, rs chan<- DevicesInfo) error {
	if rs == nil {
		return errors.New("empty result channel")
	}
	deviceInfo := DevicesInfo{}
	defer func(di *DevicesInfo) {
		rs <- *di
	}(&deviceInfo)

	spec, err := dp.RuntimeOperator.GetContainerInfoByID(ctx, c.Id)
	if err != nil {
		return contactError(err, fmt.Sprintf("cannot get container devices by container id (%s)", c.Id))
	}
	if spec.Linux == nil || spec.Linux.Resources == nil || len(spec.Linux.Resources.Devices) > maxDevicesNum {
		return contactError(errors.New("device error"),
			fmt.Sprintf("devices in container is too much (%v) or empty", maxDevicesNum))
	}
	if spec.Process == nil || len(spec.Process.Env) > maxEnvNum {
		return contactError(errors.New("env error"), fmt.Sprintf("env in container is too much (%v) or empty",
			maxEnvNum))
	}

	envs := spec.Process.Env
	sort.Strings(envs)
	for _, e := range envs {
		if strings.Contains(e, ascendDeviceInfo) {
			deviceInfo, err = dp.getDevicesWithAscendRuntime(e, c)
			return err
		}
	}

	deviceInfo, err = dp.getDevicesWithoutAscendRuntime(spec, c)
	return err
}

func (dp *DevicesParser) getDevicesWithoutAscendRuntime(spec v1.Spec, c *CommonContainer) (DevicesInfo, error) {
	deviceInfo := DevicesInfo{}
	devicesIDs, err := filterNPUDevices(spec)
	if err != nil {
		hwlog.RunLog.Debugf("filter npu devices failed by container id (%s), err is %v", c.Id, err)
		return DevicesInfo{}, nil
	}
	hwlog.RunLog.Debugf("filter npu devices %v in container (%s)", devicesIDs, c.Id)

	if len(devicesIDs) != 0 {
		if deviceInfo, err = makeUpDeviceInfo(c); err == nil {
			deviceInfo.Devices = devicesIDs
			return deviceInfo, nil
		}
		hwlog.RunLog.Error(err)
		return DevicesInfo{}, err
	}

	return DevicesInfo{}, nil
}

func (dp *DevicesParser) getDevicesWithAscendRuntime(ascendDevEnv string, c *CommonContainer) (DevicesInfo, error) {
	hwlog.RunLog.Debugf("get device info by env (%s) in %s", ascendDevEnv, c.Id)
	devInfo := strings.Split(ascendDevEnv, "=")
	if len(devInfo) != ascendEnvPart {
		return DevicesInfo{}, fmt.Errorf("an invalid %s env(%s)", ascendDeviceInfo, ascendDevEnv)
	}
	devList := strings.Split(devInfo[1], ",")

	devicesIDs := make([]int, 0, len(devList))
	for _, devID := range devList {
		id, err := strconv.Atoi(devID)
		if err != nil {
			hwlog.RunLog.Errorf("container (%s) has an invalid device ID (%v) in %s, error is %s", c.Id, devID,
				ascendDeviceInfo, err)
			continue
		}
		devicesIDs = append(devicesIDs, id)
	}

	if len(devicesIDs) != 0 {
		var err error
		if deviceInfo, err := makeUpDeviceInfo(c); err == nil {
			deviceInfo.Devices = devicesIDs
			return deviceInfo, nil
		}
		hwlog.RunLog.Error(err)
		return DevicesInfo{}, err
	}

	return DevicesInfo{}, nil
}

func (dp *DevicesParser) getDevWithoutAscendRuntimeInIsula(containerInfo isula.ContainerJson,
	c *CommonContainer) (DevicesInfo, error) {
	deviceInfo := DevicesInfo{}
	devicesIDs, err := filterNPUDevicesInIsula(containerInfo)
	if err != nil {
		hwlog.RunLog.Debugf("filter npu devices failed by container id (%s), err is %v", c.Id, err)
		return DevicesInfo{}, nil
	}
	hwlog.RunLog.Debugf("filter npu devices %v in container (%s)", devicesIDs, c.Id)

	if len(devicesIDs) != 0 {
		if deviceInfo, err = makeUpDeviceInfo(c); err == nil {
			deviceInfo.Devices = devicesIDs
			return deviceInfo, nil
		}
		hwlog.RunLog.Error(err)
		return DevicesInfo{}, err
	}

	return DevicesInfo{}, nil
}

func (dp *DevicesParser) parseDeviceInIsula(ctx context.Context, c *CommonContainer, rs chan<- DevicesInfo) error {
	if rs == nil {
		return errors.New("empty result channel")
	}

	deviceInfo := DevicesInfo{}
	defer func(di *DevicesInfo) {
		rs <- *di
	}(&deviceInfo)

	if len(c.Id) > maxCgroupPath {
		return fmt.Errorf("the containerId (%s) is too long", c.Id)
	}
	containerInfo, err := dp.RuntimeOperator.GetIsulaContainerInfoByID(ctx, c.Id)
	if err != nil {
		return contactError(err, fmt.Sprintf("getting config of container(%s) fail", c.Id))
	}
	if containerInfo.HostConfig == nil || containerInfo.Config == nil {
		return errors.New("empty container info")
	}

	envs := containerInfo.Config.Env
	for _, env := range envs {
		if strings.Contains(env, ascendDeviceInfo) {
			deviceInfo, err = dp.getDevicesWithAscendRuntime(env, c)
			return err
		}
	}

	deviceInfo, err = dp.getDevWithoutAscendRuntimeInIsula(containerInfo, c)
	return err
}

func (dp *DevicesParser) collect(ctx context.Context, r <-chan DevicesInfo, ct int32) (DevicesInfos, error) {
	if r == nil {
		return nil, errors.New("receiving channel is empty")
	}
	if ct < 0 {
		return nil, nil
	}

	results := make(map[string]DevicesInfo, ct)
	for {
		select {
		case info, ok := <-r:
			if !ok {
				return nil, nil
			}
			if info.ID != "" {
				results[info.ID] = info
			}
			if ct -= 1; ct <= 0 {
				return results, nil
			}
		case _, ok := <-ctx.Done():
			if !ok {
				return nil, nil
			}
			dp.err <- ErrFromContext
			return nil, nil
		}
	}
}

func (dp *DevicesParser) doParse(resultOut chan<- DevicesInfos) {
	var result DevicesInfos = nil
	defer func(rslt DevicesInfos) {
		if resultOut != nil {
			resultOut <- rslt
			close(resultOut)
		}
	}(result)

	ctx := context.Background()
	containers, err := dp.RuntimeOperator.GetContainers(ctx)
	if err != nil {
		dp.err <- err
		return
	}

	l := len(containers)
	if l == 0 || l > maxContainers {
		hwlog.RunLog.Debugf("get %d containers from cri interface, return empty data", l)
		dp.result <- make(DevicesInfos)
		return
	}

	r := make(chan DevicesInfo)
	defer close(r)
	wg := sync.WaitGroup{}
	wg.Add(l)

	for _, container := range containers {
		go func(container *CommonContainer, c context.Context) {
			if err := dp.parseDevices(c, container, r); err != nil {
				dp.err <- err
			}
			wg.Done()
		}(container, ctx)
	}
	ctx, cancelFn := context.WithTimeout(ctx, withDefault(dp.Timeout, parsingNpuDefaultTimeout))
	defer cancelFn()
	if result, err = dp.collect(ctx, r, int32(l)); result != nil && err == nil {
		dp.result <- result
	}
	wg.Wait()
}

// FetchAndParse triggers the asynchronous process of querying and analyzing all containers
// resultOut channel is for fetching the current result
func (dp *DevicesParser) FetchAndParse(resultOut chan<- DevicesInfos) {
	if dp.err == nil {
		hwlog.RunLog.Debug("device paster is not initialized")
		return
	}
	go dp.doParse(resultOut)
}

func withDefault(v time.Duration, d time.Duration) time.Duration {
	if v == 0 {
		return d
	}

	return v
}

// query the MajorID of NPU devices
func getNPUMajorID() ([]string, error) {
	const (
		deviceCount   = 2
		maxSearchLine = 512
	)

	path, err := utils.CheckPath("/proc/devices")
	if err != nil {
		return nil, err
	}
	majorID := make([]string, 0, deviceCount)
	f, err := os.Open(path)
	if err != nil {
		return majorID, err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			hwlog.RunLog.Error(err)
		}
	}()
	s := bufio.NewScanner(f)
	count := 0
	for s.Scan() {
		// prevent from searching too many lines
		if count > maxSearchLine {
			break
		}
		count++
		text := s.Text()
		matched, err := regexp.MatchString("^[0-9]{1,3}\\s[v]?devdrv-cdev$", text)
		if err != nil {
			return majorID, err
		}
		if !matched {
			continue
		}
		fields := strings.Fields(text)
		majorID = append(majorID, fields[0])
	}
	return majorID, nil
}

func npuMajor() []string {
	npuMajorFetchCtrl.Do(func() {
		var err error
		npuMajorID, err = getNPUMajorID()
		if err != nil {
			return
		}
	})
	return npuMajorID
}

func contains(slice []string, target string) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

func contactError(err error, msg string) error {
	return fmt.Errorf("%s->%s", err.Error(), msg)
}

func filterNPUDevices(spec v1.Spec) ([]int, error) {
	if spec.Linux == nil || spec.Linux.Resources == nil {
		return nil, errors.New("empty spec info")
	}

	const base = 10
	devIDs := make([]int, 0, sliceLen8)
	majorIDs := npuMajor()
	for _, dev := range spec.Linux.Resources.Devices {
		if dev.Minor == nil || dev.Major == nil {
			// do not monitor privileged container
			continue
		}
		if *dev.Minor > math.MaxInt32 {
			return nil, fmt.Errorf("get wrong device ID (%v)", dev.Minor)
		}
		major := strconv.FormatInt(*dev.Major, base)
		if dev.Type == charDevice && contains(majorIDs, major) {
			devIDs = append(devIDs, int(*dev.Minor))
		}
	}

	return devIDs, nil
}

// filterNPUDevicesInIsula get id of device from containerJson(containerInfo)
func filterNPUDevicesInIsula(containerInfo isula.ContainerJson) ([]int, error) {
	privileged := containerInfo.HostConfig.Privileged
	if privileged {
		return nil, errors.New("it's a privileged container and skip it")
	}

	devIDs := make([]int, 0, sliceLen8)
	devices := containerInfo.HostConfig.Devices
	for _, dev := range devices {
		Id, err := getDevIdFromPath(devicePathPattern, dev.PathInContainer)
		if err != nil {
			hwlog.RunLog.Warn(err)
			continue
		}
		devIDs = append(devIDs, Id)
	}

	return devIDs, nil
}

func getDevIdFromPath(pattern, path string) (int, error) {
	if match, err := regexp.MatchString(pattern, path); !match || err != nil {
		return -1, fmt.Errorf("unexpected path of device: %s", path)
	}
	number := regexp.MustCompile(`\d+`)
	IdStr := number.FindString(path)
	Id, err := strconv.Atoi(IdStr)
	if err != nil {
		return -1, fmt.Errorf("unexpected device ID (%v)", IdStr)
	}
	if Id > math.MaxInt32 {
		return -1, fmt.Errorf("get wrong device ID (%v)", Id)
	}
	return Id, nil
}
