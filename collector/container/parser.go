// Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

// Package container for monitoring containers' npu allocation
package container

import (
	"bufio"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"huawei.com/npu-exporter/dsmi"
	"huawei.com/npu-exporter/hwlog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	procMountInfo               = "/proc/self/mountinfo"
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
	maxNpuCardsNum              = 64
	namespaceMoby               = "moby"   // Docker
	namespaceK8s                = "k8s.io" // CRI + Containerd
)

const (
	EndpointTypeContainerd = iota // K8S + Containerd
	EndpointTypeDockerd           // Docker with or without K8S
)

var (
	ErrUnknownCgroupsPathType = errors.New("unknown cgroupsPath type")
	ErrParseFail              = errors.New("parsing fail")
	ErrNoCgroupController     = errors.New("no cgroup controller")
	ErrNoCgroupHierarchy      = errors.New("no cgroup hierarchy")
	ErrFromContext            = errors.New("error from context")

	npuMajorID               string
	npuMajorFetchCtrl        sync.Once
	parsingNpuDefaultTimeout = 5 * time.Second
)

type CntNpuMonitorOpts struct {
	ContainerdAddress string
	EndpointType      int
	Endpoint          string
}

// MakeDevicesParser evaluates option settings and make an instance according to it
func MakeDevicesParser(opts CntNpuMonitorOpts) *DevicesParser {
	runtimeOperator := &ContainerdRuntimeOperator{Endpoint: opts.ContainerdAddress}
	parser := &DevicesParser{}

	switch opts.EndpointType {
	case EndpointTypeContainerd:
		runtimeOperator.Namespace = namespaceK8s
		parser.RuntimeOperator = runtimeOperator
		parser.NameFetcher = &CriNameFetcher{
			Endpoint: opts.Endpoint,
		}
	case EndpointTypeDockerd:
		runtimeOperator.Namespace = namespaceMoby
		parser.RuntimeOperator = runtimeOperator
		parser.NameFetcher = &DockerNameFetcher{
			Endpoint: opts.Endpoint,
		}
	default:
		hwlog.Errorf("Invalid type value %d", opts.EndpointType)
	}

	return parser
}

type DevicesInfo struct {
	ID      string
	Name    string
	Devices []int
}

type DevicesInfos = map[string]DevicesInfo

type DevicesParser struct {
	// instances
	result chan DevicesInfos
	err    chan error
	// configuration
	RuntimeOperator RuntimeOperator
	NameFetcher     NameFetcher
	Timeout         time.Duration
}

// Init initializes connection to containerd daemon and to CRI server or dockerd daemon based on name fetcher setting
func (dp *DevicesParser) Init() error {
	if err := dp.RuntimeOperator.Init(); err != nil {
		return errors.Wrapf(err, "connecting to container runtime failed")
	}

	if err := dp.NameFetcher.Init(); err != nil {
		return errors.Wrapf(err, "init name fetcher fail")
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
	_ = dp.NameFetcher.Close()
}

func (dp *DevicesParser) parseDevices(ctx context.Context, containerID string, result chan<- DevicesInfo) error {
	if result == nil {
		hwlog.Fatal("empty result channel")
	}

	deviceInfo := DevicesInfo{}
	defer func(di *DevicesInfo) {
		result <- *di
	}(&deviceInfo)

	p, err := dp.RuntimeOperator.CgroupsPath(ctx, containerID)
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
		return errors.Wrapf(err, "parsing Ascend devices of container %s fail", containerID)
	}

	if hasAscend {
		deviceInfo.ID = containerID
		deviceInfo.Name = dp.NameFetcher.Name(containerID)
		deviceInfo.Devices = devicesIDs
	}
	return nil
}

func (dp *DevicesParser) collect(ctx context.Context, r <-chan DevicesInfo, counter int32) DevicesInfos {
	if r == nil {
		hwlog.Fatal("receiving channel is empty")
	}
	if counter < 0 {
		return nil
	}

	results := make(map[string]DevicesInfo, counter)
	for {
		select {
		case info := <-r:
			if info.ID != "" {
				results[info.ID] = info
			}
			if counter -= 1; counter <= 0 {
				return results
			}
		case <-ctx.Done():
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
	ids, err := dp.RuntimeOperator.ContainerIDs(ctx)
	if err == ErrNoContainers {
		return
	} else if err != nil {
		dp.err <- err
		return
	}

	l := len(ids)
	if l == 0 {
		return
	}

	r := make(chan DevicesInfo)
	defer close(r)
	ctx, cancelFn := context.WithTimeout(ctx, withDefault(dp.Timeout, parsingNpuDefaultTimeout))
	defer cancelFn()
	for _, id := range ids {
		go func(containerId string) {
			if err := dp.parseDevices(ctx, containerId, r); err != nil {
				dp.err <- err
			}
		}(id)
	}

	if result = dp.collect(ctx, r, int32(l)); result != nil {
		dp.result <- result
	}
}

// FetchAndParse triggers the asynchronized process of querying and analyzing all containers
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
	f, err := os.Open(procMountInfo)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), procMountInfoColSep)
		l := len(cols)
		if l < expectProcMountInfoColNum {
			return "", errors.Wrapf(ErrParseFail,
				"mount info record has less than %d columns", expectProcMountInfoColNum)
		}

		if cols[l-3] != "cgroup" {
			continue
		}

		for _, opt := range strings.Split(cols[l-1], ",") {
			if opt == controller {
				return cols[4], nil
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
		hwlog.Error("systemd cgroup path must have three parts separated by colon")
		return ""
	}

	slicePath := parseSlice(cols[0])
	if slicePath == "" {
		hwlog.Error("failed to parse the slice part of the cgroupsPath")
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
		hwlog.Errorf("invalid slice %s when parsing slice part of systemd cgroup path", slice)
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
			hwlog.Errorf("slice %s contains invalid content of continuous double dashes", slice)
			return ""
		}
		b.WriteRune('/')
		b.WriteString(prefix)
		b.WriteString(part)
		b.WriteString(suffixSlice)
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

var ScanForAscendDevices = func(devicesListFile string) ([]int, bool, error) {
	minorNumbers := make([]int, 0, 8)
	majorID := npuMajor()
	if majorID == "" {
		return nil, false, fmt.Errorf("majorID is null")
	}

	f, err := os.Open(devicesListFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, ErrNoCgroupHierarchy
		}
		return nil, false, errors.Wrapf(err, "error while opening devices cgroup file %q", devicesListFile)
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

		if fields[0] == "c" && majorMinor[0] == majorID {
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

func npuMajor() string {
	npuMajorFetchCtrl.Do(func() {
		var err error
		npuMajorID, err = dsmi.GetDeviceManager().GetNPUMajorID()
		if err != nil {
			return
		}
	})
	return npuMajorID
}
