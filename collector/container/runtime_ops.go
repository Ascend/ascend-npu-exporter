// Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

// Package container for monitoring containers' npu allocation
package container

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	"huawei.com/npu-exporter/collector/container/v1"
	"huawei.com/npu-exporter/hwlog"
)

const (
	labelK8sPodNamespace = "io.kubernetes.pod.namespace"
	labelK8sPodName      = "io.kubernetes.pod.name"
	// DefaultDockerShim default docker shim sock address
	DefaultDockerShim = "unix:///var/run/dockershim.sock"
	// DefaultContainerdAddr default containerd sock address
	DefaultContainerdAddr = "unix:///run/containerd/containerd.sock"
	// DefaultDockerAddr default docker containerd sock address
	DefaultDockerAddr = "unix:///var/run/docker/containerd/docker-containerd.sock"
	grpcHeader        = "containerd-namespace"
)

var (
	// ErrNoContainers means no containers are discovered
	ErrNoContainers = errors.New("no containers")
)

// RuntimeOperator wraps operations against container runtime
type RuntimeOperator interface {
	Init() error
	Close() error
	GetContainers(ctx context.Context) ([]*v1alpha2.Container, error)
	CgroupsPath(ctx context.Context, id string) (string, error)
}

// ContainerdRuntimeOperator implements RuntimeOperator interface
type ContainerdRuntimeOperator struct {
	criConn   *grpc.ClientConn
	conn      *grpc.ClientConn
	criClient v1alpha2.RuntimeServiceClient
	client    v1.ContainersClient
	// CriEndpoint CRI server endpoint
	CriEndpoint string
	// OciEndpoint containerd Server endpoint
	OciEndpoint string
	// Namespace the namespace of containerd
	Namespace string
	// UseBackup use back up address or not
	UseBackup bool
}

// Init initializes container runtime operator
func (operator *ContainerdRuntimeOperator) Init() error {
	criConn, err := GetConnection(operator.CriEndpoint)
	if err != nil || criConn == nil {
		return errors.New("connecting to CRI server failed")
	}
	operator.criClient = v1alpha2.NewRuntimeServiceClient(criConn)
	operator.criConn = criConn

	conn, err := GetConnection(operator.OciEndpoint)
	if err != nil || conn == nil {
		hwlog.RunLog.Errorf("failed to get OCI connection")
		if operator.UseBackup {
			hwlog.RunLog.Errorf("try again")
			conn, err = GetConnection(DefaultContainerdAddr)
			if err != nil {
				return err
			}
		}
	}
	operator.client = v1.NewContainersClient(conn)
	operator.conn = conn
	return nil
}

// Close closes container runtime operator
func (operator *ContainerdRuntimeOperator) Close() error {
	err := operator.conn.Close()
	if err != nil {
		return err
	}
	err = operator.criConn.Close()
	if err != nil {
		return err
	}
	return nil
}

// GetContainers returns all containers' IDs
func (operator *ContainerdRuntimeOperator) GetContainers(ctx context.Context) ([]*v1alpha2.Container, error) {
	filter := &v1alpha2.ContainerFilter{}
	st := &v1alpha2.ContainerStateValue{}
	st.State = v1alpha2.ContainerState_CONTAINER_RUNNING
	filter.State = st
	request := &v1alpha2.ListContainersRequest{
		Filter: filter,
	}
	if operator.criClient == nil {
		return nil, errors.New("criClient is empty")
	}
	r, err := operator.criClient.ListContainers(context.Background(), request)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, err
	}
	return r.Containers, nil
}

// CgroupsPath returns the cgroup path from spec of specified container
func (operator *ContainerdRuntimeOperator) CgroupsPath(ctx context.Context, id string) (string, error) {
	if operator.client == nil {
		return "", errors.New("oci client is empty")
	}
	resp, err := operator.client.Get(setGrpcNamespaceHeader(ctx, operator.Namespace), &v1.GetContainerRequest{
		Id: id,
	})
	if err != nil {
		hwlog.RunLog.Error("get call OCI get method failed")
		return "", err
	}
	s := v1.Spec{}
	if err := json.Unmarshal(resp.Container.Spec.Value, &s); err != nil {
		hwlog.RunLog.Error("unmarshal OCI response failed")
		return "", err
	}
	return s.Linux.CgroupsPath, nil
}

type nsKey struct{}

func setGrpcNamespaceHeader(ctx context.Context, namespace string) context.Context {
	context.WithValue(ctx, nsKey{}, namespace)
	ns := metadata.Pairs(grpcHeader, namespace)
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = ns
	} else {
		md = metadata.Join(ns, md)
	}
	return metadata.NewOutgoingContext(ctx, md)
}
