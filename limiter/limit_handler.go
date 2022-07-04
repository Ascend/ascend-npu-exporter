//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package limiter implement a token bucket limiter
package limiter

import (
	"context"
	"math"
	"net/http"
	"syscall"
	"time"

	"huawei.com/npu-exporter/hwlog"
	"huawei.com/npu-exporter/utils"
)

const (
	kilo                  = 1000.0
	defaultDataLimit      = 1024 * 1024 * 10
	defaultMaxConcurrency = 1024
	second5               = 5
)

type limitHandler struct {
	concurrency chan struct{}
	httpHandler http.Handler
	log         bool
	method      string
	limitBytes  int64
}

// ServeHTTP implement http.Handler
func (h *limitHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, h.limitBytes)
	ctx := context.TODO()
	reqID := req.Header.Get(hwlog.ReqID.String())
	if reqID != "" {
		ctx = context.WithValue(context.Background(), hwlog.ReqID, reqID)
	}
	id := req.Header.Get(hwlog.UserID.String())
	if id != "" {
		ctx = context.WithValue(ctx, hwlog.UserID, id)
	}
	path := req.URL.Path
	clientIP := utils.ClientIP(req)
	clientUserAgent := req.UserAgent()
	select {
	case _, ok := <-h.concurrency:
		if !ok {
			//  channel closed and no need return token
			return
		}

		if h.method != "" && req.Method != h.method {
			http.NotFound(w, req)
			//  return to token bucket
			h.concurrency <- struct{}{}
			return
		}
		hwlog.RunLog.Debugf("token count:%d", len(h.concurrency))
		ctx := context.Background()
		cancelCtx, cancelFunc := context.WithCancel(ctx)
		start := time.Now()
		go returnToken(cancelCtx, h.concurrency)
		h.httpHandler.ServeHTTP(w, req)
		stop := time.Since(start)
		cancelFunc()
		if stop < second5*time.Second {
			h.concurrency <- struct{}{}
		}
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / kilo / kilo))
		if h.log {
			hwlog.RunLog.InfofWithCtx(ctx, "%s %s: %s <%3d> (%dms) |%15s |%s |%d", req.Proto, req.Method, path,
				http.StatusOK, latency, clientIP, clientUserAgent, syscall.Getuid())
		}
	default:
		hwlog.RunLog.WarnfWithCtx(ctx, "Reject Request:%s: %s <%3d> |%15s |%s |%d ", req.Method, path,
			http.StatusServiceUnavailable, clientIP, clientUserAgent, syscall.Getuid())
		http.Error(w, "503 too busy", http.StatusServiceUnavailable)
	}
}

func returnToken(ctx context.Context, concurrency chan struct{}) {
	defer func() {
		if err := recover(); err != nil {
			hwlog.RunLog.Errorf("go routine failed with %#v", err)
		}
	}()
	timeAfterTrigger := time.After(time.Second * second5)
	if concurrency == nil || timeAfterTrigger == nil {
		hwlog.RunLog.Error("return token error")
		return
	}

	for {
		select {
		case _, ok := <-timeAfterTrigger:
			if !ok {
				return
			}
			concurrency <- struct{}{}
			// reach the return deadline
			hwlog.RunLog.Debugf("recover token num：%d", len(concurrency))
			return
		case _, ok := <-ctx.Done():
			hwlog.RunLog.Debugf("goroutine canceled")
			if !ok {
				return
			}
			return
		}

	}

}

// NewLimitHandler new a bucket-token limiter
func NewLimitHandler(maxConcur, maxConcurrency int, handler http.Handler, printLog bool) http.Handler {
	return NewLimitHandlerWithMethod(maxConcur, maxConcurrency, handler, printLog, "")
}

// NewLimitHandlerWithMethod  new a bucket-token limiter with specific http method
func NewLimitHandlerWithMethod(maxConcur, maxConcurrency int, handler http.Handler, printLog bool,
	httpMethod string) http.Handler {
	if maxConcur < 1 || maxConcur > maxConcurrency {
		hwlog.RunLog.Fatal("maxConcurrency parameter error")
	}
	concur := make(chan struct{}, maxConcur)
	return createHandler(concur, handler, printLog, httpMethod, defaultDataLimit)
}

// NewLimitHandlerWithBodyLimit  new a bucket-token limiter with bodysize limit
func NewLimitHandlerWithBodyLimit(maxConcur int, handler http.Handler, printLog bool,
	httpMethod string, bodySizeLimit int64) http.Handler {
	if maxConcur < 1 || maxConcur > defaultMaxConcurrency {
		hwlog.RunLog.Fatal("maxConcurrency parameter error")
	}
	concur := make(chan struct{}, maxConcur)
	return createHandler(concur, handler, printLog, httpMethod, bodySizeLimit)
}

func createHandler(concur chan struct{}, handler http.Handler, printLog bool,
	httpMethod string, bodySizeLimit int64) *limitHandler {
	h := &limitHandler{
		concurrency: concur,
		httpHandler: handler,
		log:         printLog,
		method:      httpMethod,
		limitBytes:  bodySizeLimit,
	}
	for i := 0; i < cap(concur); i++ {
		h.concurrency <- struct{}{}
	}
	return h
}
