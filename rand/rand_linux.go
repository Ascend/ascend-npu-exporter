//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

//go:build linux || freebsd || dragonfly || solaris
// +build linux freebsd dragonfly solaris

// Package rand implement the security rand
package rand

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	maxReadSize = 1<<25 - 1
)

// A randomReader satisfies reads by reading the file named name.
type randomReader struct {
	f  io.Reader
	mu sync.Mutex
}

func init() {
	Reader = &randomReader{}
}

func warnBlocked() {
	fmt.Println("mindx-security/rand: blocked for 60 seconds waiting to read random data from the kernel")
}

// Read implements the interface of io.Reader
func (r *randomReader) Read(b []byte) (int, error) {
	t := time.AfterFunc(time.Minute, warnBlocked)
	defer t.Stop()
	if len(b) > maxReadSize {
		return 0, errors.New("byte size is too large")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if runtime.GOOS != "linux" {
		return 0, errors.New("not supported")
	}
	f, err := os.Open("/dev/random")
	if f == nil || err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Read(b)
}
