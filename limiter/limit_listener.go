//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package limiter implement a token bucket limit listener, refer to "golang.org/x/net/netutil" and
// change the acquire method, if acquire failed, return false immediately
package limiter

import (
	"errors"
	"huawei.com/npu-exporter/hwlog"
	"net"
	"sync"
)

const (
	maxConnection = 1024
)

// LimitListener returns a Listener that accepts at most n connections at the same time
func LimitListener(l net.Listener, n int) (net.Listener, error) {
	if n < 0 || n > maxConnection {
		return nil, errors.New("the parameter n is illegal")
	}
	bucket := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		bucket <- struct{}{}
	}
	return &localLimitListener{
		Listener: l,
		buckets:  bucket,
	}, nil
}

type localLimitListener struct {
	net.Listener
	buckets   chan struct{}
	closeOnce sync.Once
}

// acquire acquires the limiting semaphore. Returns true if successfully
// accquired, false if the listener is closed or  reach the max limit
func (l *localLimitListener) acquire() bool {
	select {
	case _, ok := <-l.buckets:
		if !ok {
			return false
		}
		return true
	default:
		return false
	}
}
func (l *localLimitListener) release() { l.buckets <- struct{}{} }

// Accept implement  net.Listener interface
func (l *localLimitListener) Accept() (net.Conn, error) {
	acquired := l.acquire()
	c, err := l.Listener.Accept()
	if err != nil {
		l.release()
		return nil, err
	}
	if !acquired {
		// once the connection reach the max limit, force close the connection
		hwlog.RunLog.Warn("limit forbidden, connection will to force closed")
		err := c.(*net.TCPConn).SetLinger(0)
		if err != nil {
			hwlog.RunLog.Warnf("Error when setting linger: %s", err)
		}
		err = c.Close()
		if err != nil {
			hwlog.RunLog.Warn(err)
		}
	}
	return &limitListenerConn{Conn: c, release: l.release}, nil
}

// close implement  net.Listener interface
func (l *localLimitListener) Close() error {
	err := l.Listener.Close()
	l.closeOnce.Do(func() { close(l.buckets) })
	return err
}

type limitListenerConn struct {
	net.Conn
	releaseOnce sync.Once
	release     func()
}

// Close override  net.Conn interface
func (l *limitListenerConn) Close() error {
	err := l.Conn.Close()
	l.releaseOnce.Do(l.release)
	return err
}
