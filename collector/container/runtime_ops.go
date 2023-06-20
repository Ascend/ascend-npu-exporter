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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	"huawei.com/npu-exporter/v5/collector/container/isula"
	"huawei.com/npu-exporter/v5/collector/container/v1"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/common-utils/utils"
)

const (
	labelK8sPodNamespace = "io.kubernetes.pod.namespace"
	labelK8sPodName      = "io.kubernetes.pod.name"
	labelContainerName   = "io.kubernetes.container.name"

	// DefaultIsuladAddr default isulad sock adress
	DefaultIsuladAddr = "unix:///run/isulad.sock"
	// DefaultDockerShim default docker shim sock address
	DefaultDockerShim = "unix:///run/dockershim.sock"
	// DefaultCRIDockerd default cri-dockerd  sock address
	DefaultCRIDockerd = "unix:///run/cri-dockerd.sock"
	// DefaultContainerdAddr default containerd sock address
	DefaultContainerdAddr = "unix:///run/containerd/containerd.sock"
	// DefaultDockerAddr default docker containerd sock address
	DefaultDockerAddr    = "unix:///run/docker/containerd/docker-containerd.sock"
	defaultDockerOnEuler = "unix:///run/docker/containerd/containerd.sock"
	grpcHeader           = "containerd-namespace"
	unixPre              = "unix://"

	IsulaContainer   = "isula"
	DefaultContainer = "docker-containerd"
)

// CommonContainer wraps some common container attribute of isulad and containerd
type CommonContainer struct {
	Id     string
	Labels map[string]string
}

// RuntimeOperator wraps operations against container runtime
type RuntimeOperator interface {
	Init() error
	Close() error
	GetContainers(ctx context.Context) ([]*CommonContainer, error)
	GetContainerInfoByID(ctx context.Context, id string) (v1.Spec, error)
	GetIsulaContainerInfoByID(ctx context.Context, id string) (isula.ContainerJson, error)
	GetContainerType() string
}

// RuntimeOperatorTool implements RuntimeOperator interface
type RuntimeOperatorTool struct {
	criConn   *grpc.ClientConn
	conn      *grpc.ClientConn
	criClient interface{}
	client    interface{}
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
func (operator *RuntimeOperatorTool) Init() error {
	start := syscall.Getuid()
	hwlog.RunLog.Debugf("the init uid is:%d", start)
	if start != 0 {
		err := syscall.Setuid(0)
		if err != nil {
			return errors.New("raise uid failed")
		}
		hwlog.RunLog.Debugf("raise uid to:%d", 0)
		defer func() {
			err = syscall.Setuid(start)
			if err != nil {
				hwlog.RunLog.Error("recover uid failed")
			}
			hwlog.RunLog.Debugf("recover uid to:%d", start)
		}()
	}
	if err := sockCheck(operator); err != nil {
		hwlog.RunLog.Error("check socket path failed")
		return err
	}

	if err := operator.initCriClient(); err != nil {
		return fmt.Errorf("init CRI client failed, %s", err)
	}

	if err := operator.initOciClient(); err != nil {
		return fmt.Errorf("init OCI client failed, %s", err)
	}
	return nil
}

func (operator *RuntimeOperatorTool) initCriClient() error {
	criConn, err := GetConnection(operator.CriEndpoint)
	if err != nil || criConn == nil {
		hwlog.RunLog.Warn("connecting to CRI server failed")
		if operator.UseBackup {
			hwlog.RunLog.Warn("use cri-dockerd address to try again")
			if utils.IsExist(strings.TrimPrefix(DefaultCRIDockerd, unixPre)) {
				criConn, err = GetConnection(DefaultCRIDockerd)
			}
		}
	}
	if err != nil {
		return errors.New("connecting to CRI server failed")
	}
	if operator.CriEndpoint == DefaultIsuladAddr {
		operator.criClient = isula.NewRuntimeServiceClient(criConn)
	} else {
		operator.criClient = v1alpha2.NewRuntimeServiceClient(criConn)
	}
	operator.criConn = criConn
	return nil
}

func (operator *RuntimeOperatorTool) initOciClient() error {
	conn, err := GetConnection(operator.OciEndpoint)
	if err != nil || conn == nil {
		hwlog.RunLog.Warn("failed to get OCI connection")
		if operator.UseBackup {
			hwlog.RunLog.Warn("use backup address to try again")
			if utils.IsExist(strings.TrimPrefix(DefaultContainerdAddr, unixPre)) {
				conn, err = GetConnection(DefaultContainerdAddr)

			} else if utils.IsExist(strings.TrimPrefix(defaultDockerOnEuler, unixPre)) {
				conn, err = GetConnection(defaultDockerOnEuler)
			}
		}
	}
	if err != nil {
		return err
	}
	if operator.OciEndpoint == DefaultIsuladAddr {
		operator.client = isula.NewContainerServiceClient(conn)
	} else {
		operator.client = v1.NewContainersClient(conn)
	}
	operator.conn = conn
	return nil
}

func sockCheck(operator *RuntimeOperatorTool) error {
	if _, err := utils.CheckPath(strings.TrimPrefix(operator.CriEndpoint, unixPre)); err != nil {
		return err
	}
	if _, err := utils.CheckPath(strings.TrimPrefix(operator.OciEndpoint, unixPre)); err != nil {
		return err
	}
	return nil
}

// Close closes container runtime operator
func (operator *RuntimeOperatorTool) Close() error {
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
func (operator *RuntimeOperatorTool) GetContainers(ctx context.Context) ([]*CommonContainer, error) {
	if utils.IsNil(operator.criClient) || operator.criConn == nil {
		return nil, errors.New("criClient is empty")
	}
	if client, ok := operator.criClient.(v1alpha2.RuntimeServiceClient); ok {
		return getContainersByContainerd(ctx, client)
	}
	if client, ok := operator.criClient.(isula.RuntimeServiceClient); ok {
		return getContainersByIsulad(ctx, client)
	}

	hwlog.RunLog.Errorf("client %#v is unexpected", operator.criClient)
	return nil, errors.New("unexpected client type")
}

// GetContainerInfoByID use oci interface to get container
func (operator *RuntimeOperatorTool) GetContainerInfoByID(ctx context.Context, id string) (v1.Spec, error) {
	if utils.IsNil(operator.client) || operator.conn == nil {
		return v1.Spec{}, errors.New("oci client is empty")
	}

	s := v1.Spec{}
	if client, ok := operator.client.(v1.ContainersClient); ok {
		resp, err := client.Get(setGrpcNamespaceHeader(ctx, operator.Namespace), &v1.GetContainerRequest{
			Id: id,
		})
		if err != nil {
			hwlog.RunLog.Error("get call OCI get method failed")
			return v1.Spec{}, err
		}
		if err = json.Unmarshal(resp.Container.Spec.Value, &s); err != nil {
			hwlog.RunLog.Error("unmarshal OCI response failed")
			return v1.Spec{}, err
		}
		return s, nil
	}

	return s, errors.New("unexpected containerd client")
}

func (operator *RuntimeOperatorTool) GetIsulaContainerInfoByID(ctx context.Context,
	id string) (isula.ContainerJson, error) {
	containerJsonInfo := isula.ContainerJson{}
	if utils.IsNil(operator.client) || operator.conn == nil {
		return containerJsonInfo, errors.New("oci client is empty")
	}

	if client, ok := operator.client.(isula.ContainerServiceClient); ok {
		resp, err := client.Inspect(setGrpcNamespaceHeader(ctx, operator.Namespace), &isula.InspectContainerRequest{
			Id: id,
		})
		if err != nil {
			hwlog.RunLog.Error("call isula OCI Inspect method failed")
			return containerJsonInfo, err
		}
		if err = json.Unmarshal([]byte(resp.ContainerJSON), &containerJsonInfo); err != nil {
			hwlog.RunLog.Errorf("unmarshal err: %#v", err)
			return containerJsonInfo, err
		}
		return containerJsonInfo, nil
	}

	return containerJsonInfo, errors.New("unexpected isula client")
}

func (operator *RuntimeOperatorTool) GetContainerType() string {
	if operator.OciEndpoint == DefaultIsuladAddr {
		return IsulaContainer
	}
	return DefaultContainer
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

func getContainersByContainerd(ctx context.Context, client v1alpha2.RuntimeServiceClient) ([]*CommonContainer, error) {
	var allContainers []*CommonContainer
	request := genContainerRequest()
	r, err := client.ListContainers(ctx, request)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, err
	}
	for _, container := range r.Containers {
		allContainers = append(allContainers, &CommonContainer{
			Id:     container.Id,
			Labels: container.Labels,
		})
	}
	return allContainers, nil
}

func getContainersByIsulad(ctx context.Context, client isula.RuntimeServiceClient) ([]*CommonContainer, error) {
	var allContainers []*CommonContainer
	request := genIsulaRequest()
	r, err := client.ListContainers(ctx, request)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, err
	}
	for _, container := range r.Containers {
		allContainers = append(allContainers, &CommonContainer{
			Id:     container.Id,
			Labels: container.Labels,
		})
	}
	return allContainers, nil
}

func genContainerRequest() *v1alpha2.ListContainersRequest {
	filter := &v1alpha2.ContainerFilter{}
	st := &v1alpha2.ContainerStateValue{}
	st.State = v1alpha2.ContainerState_CONTAINER_RUNNING
	filter.State = st
	request := &v1alpha2.ListContainersRequest{
		Filter: filter,
	}
	return request
}

func genIsulaRequest() *isula.ListContainersRequest {
	filter := &isula.ContainerFilter{}
	st := &isula.ContainerStateValue{}
	st.State = isula.ContainerState_CONTAINER_RUNNING
	filter.State = st
	request := &isula.ListContainersRequest{
		Filter: filter,
	}
	return request
}
