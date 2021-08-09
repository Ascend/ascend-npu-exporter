// Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

// Package container for monitoring containers' npu allocation
package container

import (
	"context"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/pkg/errors"
)

var (
	// ErrNoContainers means no containers are discovered
	ErrNoContainers = errors.New("no containers")
)

// RuntimeOperator wraps operations against container runtime
type RuntimeOperator interface {
	Init() error
	Close() error
	ContainerIDs(ctx context.Context) ([]string, error)
	CgroupsPath(ctx context.Context, id string) (string, error)
}

// ContainerdRuntimeOperator implements RuntimeOperator interface
type ContainerdRuntimeOperator struct {
	client *containerd.Client

	// inputs
	Endpoint string
	Namespace string
}

// Init initializes container runtime operator
func (operator *ContainerdRuntimeOperator) Init() error {
	var err error
	operator.client, err = containerd.New(operator.Endpoint)
	if err != nil {
		return errors.Wrapf(err, "connecting to containerd daemon failed")
	}
	return nil
}

// Close closes container runtime operator
func (operator *ContainerdRuntimeOperator) Close() error {
	return operator.client.Close()
}

func (operator *ContainerdRuntimeOperator) namespacedCtx(ctx context.Context) context.Context {
	return namespaces.WithNamespace(ctx, operator.Namespace)
}

// ContainerIDs returns all containers' IDs
func (operator *ContainerdRuntimeOperator) ContainerIDs(ctx context.Context) ([]string, error) {
	containers, err := operator.client.Containers(operator.namespacedCtx(ctx))
	if err != nil {
		return nil, errors.Wrapf(err, "listing all containers fail")
	}

	l := len(containers)
	if l == 0 {
		return nil, ErrNoContainers
	}

	ids := make([]string, 0, l)
	for _, c := range containers {
		ids = append(ids, c.ID())
	}

	return ids, nil
}

// CgroupsPath returns the cgroup path from spec of specified container
func (operator *ContainerdRuntimeOperator) CgroupsPath(ctx context.Context, id string) (string, error) {
	ctx = operator.namespacedCtx(ctx)
	c, err := operator.client.LoadContainer(ctx, id)
	if err != nil {
		return "", errors.Wrapf(err, "reading container info of %s fail", id)
	}

	spec, err := c.Spec(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "reading spec of container %s fail", id)
	}

	return spec.Linux.CgroupsPath, nil
}
