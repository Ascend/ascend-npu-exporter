// Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

// Package container for monitoring containers' npu allocation
package container

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"huawei.com/npu-exporter/hwlog"
	"strings"
)

// DockerNameFetcher implements NameFetcher for situation of docker
type DockerNameFetcher struct {
	cli *client.Client
	Endpoint string
}

// Init implements NameFetcher interface
func (f *DockerNameFetcher) Init() error {
	opts := make([]client.Opt, 0, 1)
	if f.Endpoint != "" {
		opts = append(opts, client.WithHost(f.Endpoint))
	} else {
		opts = append(opts, client.FromEnv)
	}

	var err error
	f.cli, err = client.NewClientWithOpts(opts...)
	if err != nil {
		return errors.Wrapf(err, "connecting to docker daemon fail")
	}
	f.cli.NegotiateAPIVersion(context.TODO())

	return nil
}

// Name implements NameFetcher interface
func (f *DockerNameFetcher) Name(id string) string {
	containerJSON, err := f.cli.ContainerInspect(context.TODO(), id)
	if err != nil {
		hwlog.RunLog.Errorf("Docker name fetcher: inspecting container %s fail: %v", id, err)
		return ""
	}
	return strings.TrimPrefix(containerJSON.Name, "/")
}

// Close implements NameFetcher interface
func (f *DockerNameFetcher) Close() error {
	return f.cli.Close()
}
