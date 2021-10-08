//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package limiter implement a token bucket limiter
package limiter

import (
	"huawei.com/npu-exporter/hwlog"
	"huawei.com/npu-exporter/utils"
	"math"
	"net/http"
	"time"
)

const (
	kilo = 1000.0
)

type limitHandler struct {
	concurrency chan struct{}
	httpHandler http.Handler
	log         bool
}

// ServeHTTP implement http.Handler
func (h *limitHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	clientIP := utils.ClientIP(req)
	clientUserAgent := req.UserAgent()
	select {
	case _, ok := <-h.concurrency:
		if !ok {
			return
		}
		start := time.Now()
		h.httpHandler.ServeHTTP(w, req)
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / kilo / kilo))
		if h.log {
			hwlog.RunLog.Infof("%s %s: %s <%3d> (%dms) |%15s |%s ", req.Proto, req.Method, path,
				http.StatusOK, latency, clientIP, clientUserAgent)
		}
		h.concurrency <- struct{}{}
	default:
		hwlog.RunLog.Warnf("Reject Request:%s: %s <%3d> |%15s |%s ", req.Method, path,
			http.StatusServiceUnavailable, clientIP, clientUserAgent)
		http.Error(w, "503 too busy", http.StatusServiceUnavailable)
	}
}

// NewLimitHandler new a bucket-token limiter
func NewLimitHandler(maxConcur, maxConcurrency int, handler http.Handler, printLog bool) http.Handler {
	if maxConcur < 1 || maxConcur > maxConcurrency {
		hwlog.RunLog.Fatal("maxConcurrency parameter error")
	}
	h := &limitHandler{
		concurrency: make(chan struct{}, maxConcur),
		httpHandler: handler,
		log:         printLog,
	}
	for i := 0; i < maxConcur; i++ {
		h.concurrency <- struct{}{}
	}
	return h
}
