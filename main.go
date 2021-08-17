//  Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main
package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"huawei.com/npu-exporter/collector"
	"huawei.com/npu-exporter/hwlog"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	port        int
	updateTime  int
	needGoInfo  bool
	ip          string
	version     bool
	concurrency int
)

const (
	portConst          = 8082
	updateTimeConst    = 5
	cacheTime          = 65 * time.Second
	portLeft           = 1025
	portRight          = 40000
	oneMinute          = 60
	defaultConcurrency = 5
	defaultLogFile     = "/var/log/mindx-dl/npu-exporter/npu-exporter.log"
)

var hwLogConfig = &hwlog.LogConfig{LogFileName: defaultLogFile}

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
		h.httpHandler.ServeHTTP(w, req)
		h.concurrency <- struct{}{}
	default:
		http.Error(w, "503 too busy", http.StatusServiceUnavailable)
	}
}

func newLimitHandler(maxConcur int, handler http.Handler) http.Handler {
	if maxConcur < 1 || maxConcur > math.MaxInt16 {
		hwlog.Fatal("maxConcurrency parameter error")
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

func main() {
	flag.Parse()
	if version {
		fmt.Printf("NPU-exporter version: %s \n", collector.BuildVersion)
		os.Exit(0)
	}
	stopCH := make(chan struct{})
	defer close(stopCH)
	// init hwlog
	initHwLogger(stopCH)
	baseParamValid()
	listenAddress := ip + ":" + strconv.Itoa(port)
	hwlog.Infof("npu exporter starting and the version is %s", collector.BuildVersion)
	stop := make(chan os.Signal)
	defer close(stop)
	signal.Notify(stop, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	reg := regPrometheus(stop)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
	http.Handle("/", http.HandlerFunc(indexHandler))
	s := &http.Server{
		Addr:    listenAddress,
		Handler: newLimitHandler(concurrency, http.DefaultServeMux),
	}
	hwlog.Warn("enable unsafe http server")
	if err := s.ListenAndServe(); err != nil {
		hwlog.Fatal("Http server error and stopped")
	}
}

func regPrometheus(stop chan os.Signal) *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collector.NewNpuCollector(cacheTime, time.Duration(updateTime)*time.Second, stop),
	)
	if needGoInfo {
		reg.MustRegister(prometheus.NewGoCollector())
	}
	return reg
}

func baseParamValid() {
	if port < portLeft || port > portRight {
		hwlog.Fatalf("the port is invalid")
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		hwlog.Fatalf("the listen ip is invalid")
	}
	ip = parsedIP.String()
	hwlog.Infof("listen on: %s", ip)
	if updateTime > oneMinute || updateTime < 1 {
		hwlog.Fatalf("the updateTime is invalid")
	}
}

func init() {
	flag.IntVar(&port, "port", portConst,
		"The server port of the http service,range[1025-40000]")
	flag.StringVar(&ip, "ip", "",
		"The listen ip of the service,0.0.0.0 is not recommended when install on Multi-NIC host")
	flag.IntVar(&updateTime, "updateTime", updateTimeConst,
		"Interval (seconds) to update the npu metrics cache,range[1-60]")
	flag.BoolVar(&needGoInfo, "needGoInfo", false,
		"If true,show golang metrics (default false)")
	flag.BoolVar(&version, "version", false,
		"If true,query the version of the program (default false)")
	flag.IntVar(&concurrency, "concurrency", defaultConcurrency,
		"The max concurrency of the http server")

	// hwlog configuration
	flag.IntVar(&hwLogConfig.LogLevel, "logLevel", 0,
		"Log level, -1-debug, 0-info, 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal(default 0)")
	flag.IntVar(&hwLogConfig.MaxAge, "maxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup log files")
	flag.BoolVar(&hwLogConfig.IsCompress, "isCompress", false,
		"Whether backup files need to be compressed (default false)")
	flag.StringVar(&hwLogConfig.LogFileName, "logFile", defaultLogFile, "Log file path")
	flag.IntVar(&hwLogConfig.MaxBackups, "maxBackups", hwlog.DefaultMaxBackups, "Maximum number of backup log files")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var proposal = "http"
	_, err := w.Write([]byte(
		`<html>
			<head><title>NPU-Exporter</title></head>
			<body>
			<h1 align="center">NPU-Exporter</h1>
			<p align="center">Welcome to use NPU-Exporter,the Prometheus metrics url is ` + proposal + `://ip ` +
			strconv.Itoa(port) + `/metrics: <a href="./metrics">Metrics</a></p>
			</body>
			</html>`))
	if err != nil {
		hwlog.Error("Write to response error")
	}
}

func initHwLogger(stopCh <-chan struct{}) {
	if err := hwlog.Init(hwLogConfig, stopCh); err != nil {
		fmt.Printf("hwlog init failed, error is %v", err)
		os.Exit(-1)
	}
}
