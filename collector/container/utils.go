// Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

// Package container for monitoring containers' npu allocation
package container

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"huawei.com/npu-exporter/hwlog"
	"huawei.com/npu-exporter/utils"
	"net"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout = 5 * time.Second
	unixProtocol   = "unix"
)

// GetConnection return the grpc connection
func GetConnection(endPoint string) (*grpc.ClientConn, error) {
	if endPoint == "" {
		return nil, fmt.Errorf("endpoint is not set")
	}
	var conn *grpc.ClientConn
	hwlog.RunLog.Debugf("connect using endpoint '%s' with '%s' timeout", utils.MaskPrefix(strings.TrimPrefix(endPoint,
		unixProtocol+"://")), defaultTimeout)
	addr, dialer, err := getAddressAndDialer(endPoint)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, err
	}
	ctx, canceFn := context.WithTimeout(context.Background(), defaultTimeout)
	defer canceFn()
	conn, err = grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithContextDialer(dialer))
	if err != nil {
		return nil, err
	}
	hwlog.RunLog.Debugf("connected successfully using endpoint: %s", utils.MaskPrefix(strings.TrimPrefix(endPoint,
		unixProtocol+"://")))
	return conn, nil
}

func parseEndpoint(endpoint string) (string, string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", err
	}

	switch u.Scheme {
	case "unix":
		return "unix", u.Path, nil
	case "tcp":
		return "tcp", u.Host, nil
	default:
		return u.Scheme, "", fmt.Errorf("protocol %q not supported", u.Scheme)
	}
}

// getAddressAndDialer returns the address parsed from the given endpoint and a context dialer.
func getAddressAndDialer(endpoint string) (string, func(ctx context.Context, addr string) (net.Conn, error), error) {
	protocol, addr, err := parseEndpoint(endpoint)
	if err != nil {
		return "", nil, err
	}
	if protocol != unixProtocol {
		return "", nil, fmt.Errorf("only support unix socket endpoint")
	}
	return addr, dial, nil
}
func dial(ctx context.Context, addr string) (net.Conn, error) {
	return (&net.Dialer{}).DialContext(ctx, unixProtocol, addr)
}
