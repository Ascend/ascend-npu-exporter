//  Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"huawei.com/npu-exporter/collector"
	"k8s.io/klog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	port       int
	updateTime int
	needGoInfo bool
)

const (
	portConst       = 8082
	updateTimeConst = 5
	cacheTime       = 65 * time.Second
	portLeft        = 1025
	portRight       = 40000
	oneMinute       = 60
)

func main() {
	flag.Parse()
	validate()
	portStr := ":" + strconv.Itoa(port)
	klog.Infof("npu exporter starting and the version is %s", collector.BuildVersion)
	reg := prometheus.NewRegistry()
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	reg.MustRegister(
		collector.NewNpuCollector(cacheTime, time.Duration(updateTime)*time.Second, stop),
	)
	if needGoInfo {
		reg.MustRegister(prometheus.NewGoCollector())
	}
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(
			`<html>
			<head><title>NPU-Exporter</title></head>
			<body>
			<h1 align="center">NPU-Exporter</h1>
			<p align="center">Welcome to use NPU-Exporter,the Prometheus metrics url is http://ip ` + portStr + `/metrics: <a href="./metrics">Metrics</a></p>
			</body>
			</html>`))
		if err != nil {
			klog.Error("Write to response error")
		}
	})
	if err := http.ListenAndServe(portStr, nil); err != nil {
		klog.Fatal("Server error and Stopped")
	}
}
func validate() {
	if port < portLeft || port > portRight {
		klog.Fatalf("the port is invalid")
	}
	if updateTime > oneMinute || updateTime < 1 {
		klog.Fatalf("the updateTime is invalid")
	}
}

func init() {
	klog.InitFlags(nil)
	flag.IntVar(&port, "port", portConst,
		"the server port of the http service,range[1025-40000]")
	flag.IntVar(&updateTime, "updateTime", updateTimeConst,
		"Interval (seconds) to update the npu metrics cache,range[1-60]")
	flag.BoolVar(&needGoInfo, "needGoInfo", false,
		"If true,show golang metrics (default false)")

}
