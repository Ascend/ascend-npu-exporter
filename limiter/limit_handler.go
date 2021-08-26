//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package limiter implement a token bucket limiter
package limiter

import (
	"huawei.com/npu-exporter/hwlog"
	"net/http"
)

type limitHandler struct {
	concurrency chan struct{}
	httpHandler http.Handler
}

// ServeHTTP implement http.Handler
func (h *limitHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	select {
	case _, ok := <-h.concurrency:
		if !ok {
			return
		}
		hwlog.RunLog.Infof("received request:%s\t%s\t%s%s\t%s", req.Method, req.Proto, req.Host,
			req.URL.String(), req.UserAgent())
		h.httpHandler.ServeHTTP(w, req)
		h.concurrency <- struct{}{}
	default:
		hwlog.RunLog.Warnf("rejected request:%s\t%s\t%s%s\t%s", req.Method, req.Proto, req.Host,
			req.URL.String(), req.UserAgent())
		http.Error(w, "503 too busy", http.StatusServiceUnavailable)
	}
}

// NewLimitHandler new a bucket-token limiter
func NewLimitHandler(maxConcur, maxConcurrency int, handler http.Handler) http.Handler {
	if maxConcur < 1 || maxConcur > maxConcurrency {
		hwlog.RunLog.Fatal("maxConcurrency parameter error")
	}
	h := &limitHandler{
		concurrency: make(chan struct{}, maxConcur),
		httpHandler: handler,
	}
	for i := 0; i < maxConcur; i++ {
		h.concurrency <- struct{}{}
	}
	return h
}
