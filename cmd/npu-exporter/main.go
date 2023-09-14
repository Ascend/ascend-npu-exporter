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

// Package main
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf/plugins/common/shim"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"huawei.com/npu-exporter/v5/collector"
	"huawei.com/npu-exporter/v5/collector/container"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/common-utils/limiter"
	_ "huawei.com/npu-exporter/v5/plugins/inputs/npu"
	"huawei.com/npu-exporter/v5/versions"
)

var (
	port           int
	updateTime     int
	ip             string
	version        bool
	concurrency    int
	containerMode  string
	containerd     string
	endpoint       string
	limitIPReq     string
	platform       string
	limitIPConn    int
	limitTotalConn int
	cacheSize      int
	pollInterval   time.Duration
)

const (
	portConst               = 8082
	updateTimeConst         = 5
	cacheTime               = 65 * time.Second
	portLeft                = 1025
	portRight               = 40000
	oneMinute               = 60
	defaultConcurrency      = 5
	defaultLogFile          = "/var/log/mindx-dl/npu-exporter/npu-exporter.log"
	containerModeDocker     = "docker"
	containerModeContainerd = "containerd"
	containerModeIsula      = "isula"
	unixPre                 = "unix://"
	prometheusPlatform      = "Prometheus"
	telegrafPlatform        = "Telegraf"
	timeout                 = 10
	maxHeaderBytes          = 1024
	// tenDays ten days
	tenDays           = 10
	maxIPConnLimit    = 128
	maxConcurrency    = 512
	defaultConnection = 20
)

var hwLogConfig = &hwlog.LogConfig{LogFileName: defaultLogFile, ExpiredTime: hwlog.DefaultExpiredTime,
	CacheSize: hwlog.DefaultCacheSize}

func main() {
	flag.Parse()
	if version {
		fmt.Printf("NPU-exporter version: %s \n", versions.BuildVersion)
		return
	}

	switch platform {
	case prometheusPlatform:
		prometheusProcess()
	case telegrafPlatform:
		telegrafProcess()
	default:
		fmt.Fprintf(os.Stderr, "err platform input")
		os.Exit(1)
	}
}

func initConfig() *limiter.HandlerConfig {
	conf := &limiter.HandlerConfig{
		PrintLog:         true,
		Method:           http.MethodGet,
		LimitBytes:       limiter.DefaultDataLimit,
		TotalConCurrency: concurrency,
		IPConCurrency:    limitIPReq,
		CacheSize:        limiter.DefaultCacheSize,
	}
	return conf
}

func newServerAndListener(conf *limiter.HandlerConfig) (*http.Server, net.Listener) {
	handler, err := limiter.NewLimitHandlerV2(http.DefaultServeMux, conf)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, nil
	}
	s := &http.Server{
		Addr:           ip + ":" + strconv.Itoa(port),
		Handler:        handler,
		ReadTimeout:    timeout * time.Second,
		WriteTimeout:   timeout * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
		ErrorLog:       log.New(&hwlog.SelfLogWriter{}, "", log.Lshortfile),
	}
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, nil
	}
	limitLs, err := limiter.LimitListener(ln, limitTotalConn, limitIPConn, limiter.DefaultCacheSize)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, nil
	}
	return s, limitLs
}

func readCntMonitoringFlags() container.CntNpuMonitorOpts {
	opts := container.CntNpuMonitorOpts{UserBackUp: true}
	switch containerMode {
	case containerModeDocker:
		opts.EndpointType = container.EndpointTypeDockerd
		opts.OciEndpoint = container.DefaultDockerAddr
		opts.CriEndpoint = container.DefaultDockerShim
	case containerModeContainerd:
		opts.EndpointType = container.EndpointTypeContainerd
		opts.OciEndpoint = container.DefaultContainerdAddr
		opts.CriEndpoint = container.DefaultContainerdAddr
	case containerModeIsula:
		opts.EndpointType = container.EndpointTypeIsula
		opts.OciEndpoint = container.DefaultIsuladAddr
		opts.CriEndpoint = container.DefaultIsuladAddr
	default:
		hwlog.RunLog.Error("invalid container mode setting,reset to docker")
		opts.EndpointType = container.EndpointTypeDockerd
		opts.OciEndpoint = container.DefaultDockerAddr
		opts.CriEndpoint = container.DefaultDockerShim
	}
	if containerd != "" {
		opts.OciEndpoint = containerd
		opts.UserBackUp = false
	}
	if endpoint != "" {
		opts.CriEndpoint = endpoint
		opts.UserBackUp = false
	}
	return opts
}

func regPrometheus(opts container.CntNpuMonitorOpts) (*prometheus.Registry, error) {
	deviceParser := container.MakeDevicesParser(opts)
	reg := prometheus.NewRegistry()
	c, err := collector.NewNpuCollector(context.Background(), cacheTime, time.Duration(updateTime)*time.Second,
		deviceParser)
	if err != nil {
		return nil, err
	}
	reg.MustRegister(c)
	return reg, nil
}

func baseParamValid() error {
	if port < portLeft || port > portRight {
		return errors.New("the port is invalid")
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return errors.New("the listen ip is invalid")
	}
	ip = parsedIP.String()
	hwlog.RunLog.Infof("listen on: %s", ip)
	if updateTime > oneMinute || updateTime < 1 {
		return errors.New("the updateTime is invalid")
	}
	if err := containerSockCheck(); err != nil {
		return err
	}
	reg := regexp.MustCompile(limiter.IPReqLimitReg)
	if !reg.Match([]byte(limitIPReq)) {
		return errors.New("limitIPReq format error")
	}
	if limitIPConn < 1 || limitIPConn > maxIPConnLimit {
		return errors.New("limitIPConn range error")
	}
	if limitTotalConn < 1 || limitTotalConn > maxConcurrency {
		return errors.New("limitTotalConn range error")
	}
	if cacheSize < 1 || cacheSize > limiter.DefaultCacheSize*tenDays {
		return errors.New("cacheSize range error")
	}
	if concurrency < 1 || concurrency > maxConcurrency {
		return errors.New("concurrency range error")
	}
	return nil
}

func containerSockCheck() error {
	if endpoint != "" && !strings.Contains(endpoint, ".sock") {
		return errors.New("endpoint file not sock address")
	}
	if containerd != "" && !strings.Contains(containerd, ".sock") {
		return errors.New("containerd file not sock address")
	}
	if endpoint != "" && !strings.Contains(endpoint, unixPre) {
		endpoint = unixPre + endpoint
	}
	if containerd != "" && !strings.Contains(containerd, unixPre) {
		containerd = unixPre + containerd
	}
	return nil
}

func init() {
	flag.IntVar(&port, "port", portConst,
		"The server port of the http service,range[1025-40000]")
	flag.StringVar(&ip, "ip", "",
		"The listen ip of the service,0.0.0.0 is not recommended when install on Multi-NIC host")
	flag.IntVar(&updateTime, "updateTime", updateTimeConst,
		"Interval (seconds) to update the npu metrics cache,range[1-60]")
	flag.BoolVar(&version, "version", false,
		"If true,query the version of the program (default false)")
	flag.StringVar(&containerMode, "containerMode", containerModeDocker,
		"Set 'docker' for monitoring docker containers or 'containerd' for CRI & containerd")
	flag.StringVar(&containerd, "containerd", "",
		"The endpoint of containerd used for listening containers' events")
	flag.StringVar(&endpoint, "endpoint", "",
		"The endpoint of the CRI  server to which will be connected")
	flag.IntVar(&concurrency, "concurrency", defaultConcurrency,
		"The max concurrency of the http server, range is [1-512]")
	// hwlog configuration
	flag.IntVar(&hwLogConfig.LogLevel, "logLevel", 0,
		"Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical(default 0)")
	flag.IntVar(&hwLogConfig.MaxAge, "maxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup log files, range [7, 700] days")
	flag.StringVar(&hwLogConfig.LogFileName, "logFile", defaultLogFile,
		"Log file path. If the file size exceeds 20MB, will be rotated")
	flag.IntVar(&hwLogConfig.MaxBackups, "maxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup log files, range is (0, 30]")
	flag.IntVar(&cacheSize, "cacheSize", limiter.DefaultCacheSize, "the cacheSize for ip limit,"+
		"range  is [1,1024000],keep default normally")
	flag.IntVar(&limitIPConn, "limitIPConn", defaultConcurrency, "the tcp connection limit for each Ip,"+
		"range  is [1,128]")
	flag.IntVar(&limitTotalConn, "limitTotalConn", defaultConnection, "the tcp connection limit for all"+
		" request,range  is [1,512]")
	flag.StringVar(&limitIPReq, "limitIPReq", "20/1",
		"the http request limit counts for each Ip,20/1 means allow 20 request in 1 seconds")
	flag.StringVar(&platform, "platform", "Prometheus", "the data reporting platform, "+
		"just support Prometheus and Telegraf")
	flag.DurationVar(&pollInterval, "poll_interval", 1*time.Second, "how often to send metrics")
}

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	var proposal = "http"
	_, err := w.Write([]byte(
		`<html>
			<head><title>NPU-Exporter</title></head>
			<body>
			<h1 align="center">NPU-Exporter</h1>
			<p align="center">Welcome to use NPU-Exporter,the Prometheus metrics url is ` + proposal + `://ip:` +
			strconv.Itoa(port) + `/metrics: <a href="./metrics">Metrics</a></p>
			</body>
			</html>`))
	if err != nil {
		hwlog.RunLog.Error("Write to response error")
	}
}

func initHwLogger() error {
	if err := hwlog.InitRunLogger(hwLogConfig, context.Background()); err != nil {
		fmt.Printf("hwlog init failed, error is %#v\n", err)
		return err
	}
	return nil
}

func prometheusProcess() {
	if err := initHwLogger(); err != nil {
		return
	}
	if err := baseParamValid(); err != nil {
		hwlog.RunLog.Error(err)
		return
	}

	hwlog.RunLog.Infof("npu exporter starting and the version is %s", versions.BuildVersion)
	opts := readCntMonitoringFlags()
	reg, err := regPrometheus(opts)
	if err != nil {
		hwlog.RunLog.Errorf("register prometheus failed")
		return
	}
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
	http.Handle("/", http.HandlerFunc(indexHandler))
	conf := initConfig()
	s, limitLs := newServerAndListener(conf)
	if s == nil || limitLs == nil {
		return
	}
	hwlog.RunLog.Warn("enable unsafe http server")
	if err := s.Serve(limitLs); err != nil {
		hwlog.RunLog.Error("Http server error and stopped")
	}
}

func telegrafProcess() {
	// create the shim. This is what will run your plugins.
	shim := shim.New()

	// If no config is specified, all imported plugins are loaded.
	// otherwise follow what the config asks for.
	// Check for settings from a config toml file,
	// (or just use whatever plugins were imported above)
	configFile := ""
	err := shim.LoadConfig(&configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err loading input: %s\n", err)
		os.Exit(1)
	}

	// run the input plugin(s) until stdin closes, or we receive a termination signal
	if err := shim.Run(pollInterval); err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s\n", err)
		os.Exit(1)
	}
}
