// Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

// Package container for monitoring containers' npu allocation
package container

import (
	"github.com/containerd/containerd/integration/remote"
	"github.com/pkg/errors"
	"huawei.com/npu-exporter/hwlog"
	cri "k8s.io/cri-api/pkg/apis"
	"time"
)

const (
	namePartSeperator    = "_"
	labelK8sPodNamespace = "io.kubernetes.pod.namespace"
	labelK8sPodName      = "io.kubernetes.pod.name"
)

var (
	criConnectionTimeout = 5 * time.Second
)

// CriNameFetcher implements NameFetcher interface for situation of k8s + containerd
type CriNameFetcher struct {
	client cri.RuntimeService
	Endpoint string
}

// Init implements NameFetcher interface
func (f *CriNameFetcher) Init() error {
	var err error
	f.client, err = remote.NewRuntimeService(f.Endpoint, criConnectionTimeout)
	if err != nil {
		return errors.Wrapf(err, "connecting to cri server failed")
	}
	return nil
}

// Name implements NameFetcher interface
func (f *CriNameFetcher) Name(id string) string {
	containerStatus, err := f.client.ContainerStatus(id)
	if err != nil {
		hwlog.Errorf("CRI name fetcher: cannot get status of container %s: %v", id, err)
		return ""
	}

	name := containerStatus.Metadata.Name
	podName, found := containerStatus.Labels[labelK8sPodName]
	if found {
		name += namePartSeperator + podName
	}
	podNamespace, found := containerStatus.Labels[labelK8sPodNamespace]
	if found {
		name += namePartSeperator + podNamespace
	}

	return name
}

// Close implements NameFetcher interface
func (f *CriNameFetcher) Close() error {
	return nil
}
