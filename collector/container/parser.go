// Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

// Package container for monitoring containers' npu allocation
package container

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	"huawei.com/npu-exporter/dsmi"
	"huawei.com/npu-exporter/hwlog"
	"huawei.com/npu-exporter/utils"
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
)

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
	case EndpointTypeDockerd:
		runtimeOperator.Namespace = namespaceMoby
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
		return errors.Wrapf(err, "connecting to container runtime failed")
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
	if rs == nil {
		hwlog.RunLog.Fatal("empty result channel")
	}

	deviceInfo := DevicesInfo{}
	defer func(di *DevicesInfo) {
		rs <- *di
	}(&deviceInfo)

	p, err := dp.RuntimeOperator.CgroupsPath(ctx, c.Id)
	if err != nil {
		return errors.Wrapf(err, "getting cgroup path of container fail")
	}

	p, err = GetCgroupPath(cgroupControllerDevices, p)
	if err != nil {
		return errors.Wrapf(err, "parsing cgroup path from spec fail")
	}
	devicesIDs, hasAscend, err := ScanForAscendDevices(filepath.Join(p, devicesList))
	if err == ErrNoCgroupHierarchy {
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "parsing Ascend devices of container %s fail", c.Id)
	}
	ns := c.Labels[labelK8sPodNamespace]
	err = validDNSRe(ns)
	if err != nil {
		return err
	}
	podName := c.Labels[labelK8sPodName]
	err = validDNSRe(podName)
	if err != nil {
		return err
	}
	if hasAscend {
		deviceInfo.ID = c.Id
		deviceInfo.Name = ns + "_" + podName + "_" + c.Metadata.Name
		deviceInfo.Devices = devicesIDs
	}
	return nil
}

func (dp *DevicesParser) collect(ctx context.Context, r <-chan DevicesInfo, ct int32) DevicesInfos {
	if r == nil {
		hwlog.RunLog.Fatal("receiving channel is empty")
	}
	if ct < 0 {
		return nil
	}

	results := make(map[string]DevicesInfo, ct)
	for {
		select {
		case info, ok := <-r:
			if !ok {
				return nil
			}
			if info.ID != "" {
				results[info.ID] = info
			}
			if ct -= 1; ct <= 0 {
				return results
			}
		case _, ok := <-ctx.Done():
			if !ok {
				return nil
			}
			dp.err <- ErrFromContext
			return nil
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
	if l == 0 {
		return
	}

	r := make(chan DevicesInfo)
	defer close(r)
	ctx, cancelFn := context.WithTimeout(ctx, withDefault(dp.Timeout, parsingNpuDefaultTimeout))
	defer cancelFn()
	for _, container := range containers {
		go func(container *v1alpha2.Container) {
			if err := dp.parseDevices(ctx, container, r); err != nil {
				dp.err <- err
			}
		}(container)
	}

	if result = dp.collect(ctx, r, int32(l)); result != nil {
		dp.result <- result
	}
}

// FetchAndParse triggers the asynchronous process of querying and analyzing all containers
// resultOut channel is for fetching the current result
func (dp *DevicesParser) FetchAndParse(resultOut chan<- DevicesInfos) {
	go dp.doParse(resultOut)
}

func withDefault(v time.Duration, d time.Duration) time.Duration {
	if v == 0 {
		return d
	}

	return v
}

// GetCgroupPath the method of caculate cgroup path of device.list
var GetCgroupPath = func(controller, specCgroupsPath string) (string, error) {
	devicesController, err := getCgroupControllerPath(controller)
	if err != nil {
		return "", errors.Wrapf(err, "getting mount point of cgroup devices subsystem fail")
	}

	hierarchy, err := toCgroupHierarchy(specCgroupsPath)
	if err != nil {
		return "", errors.Wrapf(err, "parsing cgroups path of spec to cgroup hierarchy fail")
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
	defer f.Close()

	// parsing the /proc/self/mountinfo file content to find the mount point of specified
	// cgroup subsystem.
	// the format of the file is described in proc man page.
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), procMountInfoColSep)
		l := len(split)
		if l < expectProcMountInfoColNum {
			return "", errors.Wrapf(ErrParseFail,
				"mount info record has less than %d columns", expectProcMountInfoColNum)
		}

		// finding cgroup mount point, ignore others
		if split[l-3] != "cgroup" {
			continue
		}

		// finding the specified cgroup controller
		for _, opt := range strings.Split(split[l-1], ",") {
			if opt == controller {
				// returns the path of specified cgroup controller in fs
				return split[4], nil
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
		return "", errors.Wrapf(ErrParseFail, "failed to parse cgroupsPath value %s", cgroupsPath)
	}
	return hierarchy, nil
}

func parseSystemdCgroup(cgroup string) string {
	cols := strings.Split(cgroup, ":")
	if len(cols) != expectSystemdCgroupPathCols {
		hwlog.RunLog.Error("systemd cgroup path must have three parts separated by colon")
		return ""
	}

	slicePath := parseSlice(cols[0])
	if slicePath == "" {
		hwlog.RunLog.Error("failed to parse the slice part of the cgroupsPath")
		return ""
	}

	return filepath.Join(slicePath, getUnit(cols[1], cols[2]))
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
var ScanForAscendDevices = func(devicesListFile string) ([]int, bool, error) {
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
		return nil, false, errors.Wrapf(err, "error while opening devices cgroup file %q",
			utils.MaskPrefix(strings.TrimPrefix(devicesListFile, unixProtocol+"://")))
	}
	defer f.Close()

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

		if fields[0] == "c" && contains(majorID, majorMinor[0]) {
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

func npuMajor() []string {
	npuMajorFetchCtrl.Do(func() {
		var err error
		npuMajorID, err = dsmi.GetDeviceManager().GetNPUMajorID()
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
