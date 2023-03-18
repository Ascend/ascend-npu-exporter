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
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/opencontainers/runc/libcontainer/cgroups"
	"k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	"huawei.com/npu-exporter/v5/collector/container/v1"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/common-utils/utils"
)

const (
	procMountInfoColSep         = " "
	cgroupControllerDevices     = "devices"
	expectSystemdCgroupPathCols = 3
	expectProcMountInfoColNum   = 10
	systemdSliceHierarchySep    = "-"
	suffixSlice                 = ".slice"
	suffixScope                 = ".scope"
	defaultSlice                = "system.slice"
	devicesList                 = "devices.list"
	expectDevicesListColNum     = 3
	expectDeviceIDNum           = 2
	maxNpuCardsNum              = 512
	namespaceMoby               = "moby"   // Docker
	namespaceK8s                = "k8s.io" // CRI + Containerd
	sliceLen8                   = 8
	cgroupIndex                 = 4
	mountPointIdx               = 3
	cgroupPrePath               = 1
	cgroupSuffixPath            = 2

	ascendDeviceInfo = "ASCEND_VISIBLE_DEVICES"
	ascendEnvPart    = 2

	charDevice = "c"
)

// don't change the order
// all capabilities presents privileged
var privilegeCaps = []string{"CAP_AUDIT_CONTROL", "CAP_AUDIT_READ", "CAP_AUDIT_WRITE", "CAP_BLOCK_SUSPEND",
	"CAP_BPF", "CAP_CHECKPOINT_RESTORE", "CAP_CHOWN", "CAP_DAC_OVERRIDE", "CAP_DAC_READ_SEARCH", "CAP_FOWNER",
	"CAP_FSETID", "CAP_IPC_LOCK", "CAP_IPC_OWNER", "CAP_KILL", "CAP_LEASE", "CAP_LINUX_IMMUTABLE", "CAP_MAC_ADMIN",
	"CAP_MAC_OVERRIDE", "CAP_MKNOD", "CAP_NET_ADMIN", "CAP_NET_BIND_SERVICE", "CAP_NET_BROADCAST", "CAP_NET_RAW",
	"CAP_PERFMON", "CAP_SETFCAP", "CAP_SETGID", "CAP_SETPCAP", "CAP_SETUID", "CAP_SYSLOG", "CAP_SYS_ADMIN",
	"CAP_SYS_BOOT", "CAP_SYS_CHROOT", "CAP_SYS_MODULE", "CAP_SYS_NICE", "CAP_SYS_PACCT", "CAP_SYS_PTRACE",
	"CAP_SYS_RAWIO", "CAP_SYS_RESOURCE", "CAP_SYS_TIME", "CAP_SYS_TTY_CONFIG", "CAP_WAKE_ALARM"}

const (
	// EndpointTypeContainerd K8S + Containerd
	EndpointTypeContainerd = iota
	// EndpointTypeDockerd Docker with or without K8S
	EndpointTypeDockerd
)

var (
	// ErrUnknownCgroupsPathType cgroups path format not recognized
	ErrUnknownCgroupsPathType = errors.New("unknown cgroupsPath type")
	// ErrParseFail parsing devices.list fail
	ErrParseFail = errors.New("parsing fail")
	// ErrNoCgroupController no such cgroup controller
	ErrNoCgroupController = errors.New("no cgroup controller")
	// ErrNoCgroupHierarchy cgroup path not found
	ErrNoCgroupHierarchy = errors.New("no cgroup hierarchy")
	// ErrFromContext error is from the context
	ErrFromContext = errors.New("error from context")

	npuMajorID               []string
	npuMajorFetchCtrl        sync.Once
	parsingNpuDefaultTimeout = 3 * time.Second
	procMountInfoGet         sync.Once
	procMountInfo            string
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
		runtimeOperator.EndpointType = EndpointTypeContainerd
	case EndpointTypeDockerd:
		runtimeOperator.Namespace = namespaceMoby
		parser.RuntimeOperator = runtimeOperator
		runtimeOperator.CriEndpoint = opts.CriEndpoint
		runtimeOperator.OciEndpoint = opts.OciEndpoint
		runtimeOperator.EndpointType = EndpointTypeDockerd

	default:
		hwlog.RunLog.Errorf("Invalid type value %d", opts.EndpointType)
	}

	return parser
}

// DevicesInfo the container device information struct
type DevicesInfo struct {
	ID      string
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

func (dp *DevicesParser) parseDevices(ctx context.Context, c *v1alpha2.Container, rs chan<- DevicesInfo) error {
	if cgroups.IsCgroup2UnifiedMode() {
		hwlog.RunLog.Debugf("now use cgroup v2 to get container (%s) npu devices", c.Id)
		return dp.parseDevicesV2(ctx, c, rs)
	}

	hwlog.RunLog.Debugf("now use cgroup v1 or hybrid cgroup to get npu devices")
	return dp.parseDevicesV1(ctx, c, rs)
}

func (dp *DevicesParser) parseDevicesV2(ctx context.Context, c *v1alpha2.Container, rs chan<- DevicesInfo) error {
	if rs == nil {
		return errors.New("empty result channel")
	}
	deviceInfo := DevicesInfo{}
	defer func(di *DevicesInfo) {
		rs <- *di
	}(&deviceInfo)

	spec, err := dp.RuntimeOperator.GetContainerInfoByID(ctx, c.Id)
	if err != nil {
		return contactError(err, fmt.Sprintf("cannot get container devices by container id (%#v)", c.Id))
	}
	if spec.Linux == nil || len(spec.Linux.Devices) > maxDevicesNum {
		return contactError(errors.New("device error"), fmt.Sprintf("devices in container is too much (%v) or empty",
			maxDevicesNum))
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

func (dp *DevicesParser) getDevicesWithoutAscendRuntime(spec v1.Spec, c *v1alpha2.Container) (DevicesInfo, error) {
	deviceInfo := DevicesInfo{}
	devicesIDs, err := filterNPUDevices(spec)
	if err != nil {
		hwlog.RunLog.Debugf("filter npu devices failed by container id (%#v), err is %v", c.Id, err)
		return DevicesInfo{}, err
	}
	hwlog.RunLog.Debugf("filter npu devices %#v in container (%s)", devicesIDs, c.Id)

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

func (dp *DevicesParser) getDevicesWithAscendRuntime(ascendDevEnv string, c *v1alpha2.Container) (DevicesInfo, error) {
	hwlog.RunLog.Debugf("get device info by env (%#v) in %s, error is %s", ascendDevEnv, c.Id)
	devInfo := strings.Split(ascendDevEnv, "=")
	if len(devInfo) != ascendEnvPart {
		return DevicesInfo{}, fmt.Errorf("an invalid %s env(%#v)", ascendDeviceInfo, ascendDevEnv)
	}
	devList := strings.Split(devInfo[1], ",")

	devicesIDs := make([]int, len(devList))
	for _, devID := range devList {
		id, err := strconv.Atoi(devID)
		if err != nil {
			hwlog.RunLog.Errorf("container (%#v) has an invalid device ID (%#v) in %s, error is %s", c.Id, devID,
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

func (dp *DevicesParser) parseDevicesV1(ctx context.Context, c *v1alpha2.Container, rs chan<- DevicesInfo) error {
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
	p, err := dp.RuntimeOperator.CgroupPath(ctx, c.Id)
	if err != nil {
		return contactError(err, fmt.Sprintf("getting cgroup path of container(%#v) fail", c.Id))
	}

	p, err = GetCgroupPath(cgroupControllerDevices, p)
	if err != nil {
		return contactError(err, "parsing cgroup path from spec fail")
	}
	devicesIDs, hasAscend, err := ScanForAscendDevices(filepath.Join(p, devicesList))
	hwlog.RunLog.Debugf("filter npu devices %#v in container (%s)", devicesIDs, c.Id)
	if err == ErrNoCgroupHierarchy {
		return nil
	} else if err != nil {
		return contactError(err, fmt.Sprintf("parsing Ascend devices of container %s fail", c.Id))
	}

	if hasAscend {
		if deviceInfo, err = makeUpDeviceInfo(c); err == nil {
			deviceInfo.Devices = devicesIDs
			return nil
		}
		hwlog.RunLog.Error(err)
		return err
	}
	return nil
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
		return
	}

	r := make(chan DevicesInfo)
	defer close(r)
	wg := sync.WaitGroup{}
	wg.Add(l)

	for _, container := range containers {
		go func(container *v1alpha2.Container) {
			if err := dp.parseDevices(ctx, container, r); err != nil {
				dp.err <- err
			}
			wg.Done()
		}(container)
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

// GetCgroupPath the method of caculate cgroup path of device.list
func GetCgroupPath(controller, specCgroupsPath string) (string, error) {
	devicesController, err := getCgroupControllerPath(controller)
	if err != nil {
		return "", contactError(err, "getting mount point of cgroup devices subsystem fail")
	}

	hierarchy, err := toCgroupHierarchy(specCgroupsPath)
	if err != nil {
		return "", contactError(err, "parsing cgroups path of spec to cgroup hierarchy fail")
	}

	return filepath.Join(devicesController, hierarchy), nil
}

func getCgroupControllerPath(controller string) (string, error) {
	procMountInfoGet.Do(func() {
		pid := os.Getpid()
		procMountInfo = "/proc/" + strconv.Itoa(pid) + "/mountinfo"
	})
	path, err := utils.CheckPath(procMountInfo)
	if err != nil {
		return "", err
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			hwlog.RunLog.Error(err)
		}
	}()

	// parsing the /proc/self/mountinfo file content to find the mount point of specified
	// cgroup subsystem.
	// the format of the file is described in proc man page.
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), procMountInfoColSep)
		l := len(split)
		if l < expectProcMountInfoColNum {
			return "", contactError(ErrParseFail,
				fmt.Sprintf("mount info record has less than %d columns", expectProcMountInfoColNum))
		}

		// finding cgroup mount point, ignore others
		if split[l-mountPointIdx] != "cgroup" {
			continue
		}

		// finding the specified cgroup controller
		for _, opt := range strings.Split(split[l-1], ",") {
			if opt == controller {
				// returns the path of specified cgroup controller in fs
				return split[cgroupIndex], nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", ErrNoCgroupController
}

func toCgroupHierarchy(cgroupsPath string) (string, error) {
	var hierarchy string
	if strings.HasPrefix(cgroupsPath, "/") {
		// as cgroupfs
		hierarchy = cgroupsPath
	} else if strings.ContainsRune(cgroupsPath, ':') {
		// as systemd cgroup
		hierarchy = parseSystemdCgroup(cgroupsPath)
	} else {
		return "", ErrUnknownCgroupsPathType
	}
	if hierarchy == "" {
		return "", contactError(ErrParseFail, fmt.Sprintf("failed to parse cgroupsPath value %s", cgroupsPath))
	}
	return hierarchy, nil
}

func parseSystemdCgroup(cgroup string) string {
	pathsArr := strings.Split(cgroup, ":")
	if len(pathsArr) != expectSystemdCgroupPathCols {
		hwlog.RunLog.Error("systemd cgroup path must have three parts separated by colon")
		return ""
	}

	slicePath := parseSlice(pathsArr[0])
	if slicePath == "" {
		hwlog.RunLog.Error("failed to parse the slice part of the cgroupsPath")
		return ""
	}
	return filepath.Join(slicePath, getUnit(pathsArr[cgroupPrePath], pathsArr[cgroupSuffixPath]))
}

func parseSlice(slice string) string {
	if slice == "" {
		return defaultSlice
	}

	if len(slice) < len(suffixSlice) ||
		!strings.HasSuffix(slice, suffixSlice) ||
		strings.ContainsRune(slice, '/') {
		hwlog.RunLog.Errorf("invalid slice %s when parsing slice part of systemd cgroup path", slice)
		return ""
	}

	sliceMain := strings.TrimSuffix(slice, suffixSlice)
	if sliceMain == systemdSliceHierarchySep {
		return "/"
	}

	b := new(strings.Builder)
	prefix := ""
	for _, part := range strings.Split(sliceMain, systemdSliceHierarchySep) {
		if part == "" {
			hwlog.RunLog.Errorf("slice %s contains invalid content of continuous double dashes", slice)
			return ""
		}
		_, err := b.WriteRune('/')
		_, err = b.WriteString(prefix)
		_, err = b.WriteString(part)
		_, err = b.WriteString(suffixSlice)
		if err != nil {
			return "" // err is always nil
		}
		prefix += part + "-"
	}

	return b.String()
}

func getUnit(prefix, name string) string {
	if strings.HasSuffix(name, suffixSlice) {
		return name
	}
	return prefix + "-" + name + suffixScope
}

// ScanForAscendDevices scan ascend devices from device.list file
func ScanForAscendDevices(devicesListFile string) ([]int, bool, error) {
	minorNumbers := make([]int, 0, sliceLen8)
	majorID := npuMajor()
	if len(majorID) == 0 {
		return nil, false, fmt.Errorf("majorID is null")
	}

	f, err := os.Open(devicesListFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, ErrNoCgroupHierarchy
		}
		return nil, false, contactError(err, fmt.Sprintf("error while opening devices cgroup file %q",
			utils.MaskPrefix(strings.TrimPrefix(devicesListFile, unixPrefix+"://"))))
	}
	defer func() {
		err = f.Close()
		if err != nil {
			hwlog.RunLog.Error(err)
		}
	}()

	s := bufio.NewScanner(f)
	for s.Scan() {
		text := s.Text()
		fields := strings.Fields(text)
		if len(fields) != expectDevicesListColNum {
			return nil, false, fmt.Errorf("cgroup entry %q must have three whitespace-separated fields", text)
		}

		majorMinor := strings.Split(fields[1], ":")
		if len(majorMinor) != expectDeviceIDNum {
			return nil, false, fmt.Errorf("second field of cgroup entry %q should have one colon", text)
		}

		if fields[0] == charDevice && contains(majorID, majorMinor[0]) {
			if majorMinor[1] == "*" {
				return nil, false, nil
			}
			minorNumber, err := strconv.Atoi(majorMinor[1])
			if err != nil {
				return nil, false, fmt.Errorf("cgroup entry %q: minor number is not integer", text)
			}

			// the max NPU cards supported number is 64
			if minorNumber < maxNpuCardsNum {
				minorNumbers = append(minorNumbers, minorNumber)
			}
		}
	}

	return minorNumbers, len(minorNumbers) > 0, nil
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
	var caps []string
	if spec.Process != nil && spec.Process.Capabilities != nil {
		caps = spec.Process.Capabilities.Permitted
	}
	sort.Strings(caps)
	same := isSameStringSlice(caps, privilegeCaps)
	if same {
		hwlog.RunLog.Debug("it's a privileged container and skip it")
		return nil, nil
	}

	const base = 10
	devIDs := make([]int, 0, sliceLen8)
	majorIDs := npuMajor()
	for _, dev := range spec.Linux.Devices {
		if dev.Minor > math.MaxInt32 {
			hwlog.RunLog.Debugf("get wrong device ID (%v)", dev.Minor)
			continue
		}
		major := strconv.FormatInt(dev.Major, base)
		if dev.Type == charDevice && contains(majorIDs, major) {
			devIDs = append(devIDs, int(dev.Minor))
		}
	}

	return devIDs, nil
}
