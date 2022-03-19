//  Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.

// Package main
package main

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"huawei.com/npu-exporter/collector"
	"huawei.com/npu-exporter/collector/container"
	"huawei.com/npu-exporter/hwlog"
	"huawei.com/npu-exporter/limiter"
	"huawei.com/npu-exporter/utils"
)

var (
	port             int
	updateTime       int
	certificate      *tls.Certificate
	ip               string
	enableHTTP       bool
	caBytes          []byte
	encryptAlgorithm int
	version          bool
	tlsSuites        int
	cipherSuites     uint16
	concurrency      int
	containerMode    string
	containerd       string
	endpoint         string
	crlcerList       *pkix.CertificateList
)

const (
	dirPrefix               = "/etc/mindx-dl/npu-exporter/"
	portConst               = 8082
	updateTimeConst         = 5
	cacheTime               = 65 * time.Second
	portLeft                = 1025
	portRight               = 40000
	oneMinute               = 60
	keyStore                = dirPrefix + utils.KeyStore
	certStore               = dirPrefix + utils.CertStore
	caStore                 = dirPrefix + utils.CaStore
	crlStore                = dirPrefix + utils.CrlStore
	passFile                = dirPrefix + utils.PassFile
	passFileBackUp          = dirPrefix + utils.PassFileBackUp
	defaultConcurrency      = 5
	defaultLogFile          = "/var/log/mindx-dl/npu-exporter/npu-exporter.log"
	containerModeDocker     = "docker"
	containerModeContainerd = "containerd"
	maxConcurrency          = 50
	unixPre                 = "unix://"
	timeout                 = 10
	maxConnection           = 20
	maxHeaderBytes          = 1024
)

var hwLogConfig = &hwlog.LogConfig{LogFileName: defaultLogFile}

func main() {
	flag.Parse()
	if version {
		fmt.Printf("NPU-exporter version: %s \n", hwlog.BuildVersion)
		os.Exit(0)
	}
	stopCH := make(chan struct{})
	defer close(stopCH)
	// init hwlog
	initHwLogger(stopCH)
	validate()
	hwlog.RunLog.Infof("npu exporter starting and the version is %s", hwlog.BuildVersion)

	opts := readCntMonitoringFlags()
	stop := make(chan os.Signal)
	defer close(stop)
	signal.Notify(stop, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	reg := regPrometheus(stop, opts)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
	http.Handle("/", http.HandlerFunc(indexHandler))
	http.Handle("/v1/certstatus", http.HandlerFunc(getCertStatus))
	s, limitLs := newServerAndListener()
	if certificate != nil {
		tlsConf, err := utils.NewTLSConfig(caBytes, *certificate, cipherSuites)
		if err != nil {
			hwlog.RunLog.Fatal(err)
		}
		s.TLSConfig = tlsConf
		s.Handler = limiter.NewLimitHandlerWithMethod(concurrency, maxConcurrency,
			utils.Interceptor(http.DefaultServeMux, crlcerList), true, http.MethodGet)
		hwlog.RunLog.Info("start https server now...")
		if err := s.ServeTLS(limitLs, "", ""); err != nil {
			hwlog.RunLog.Fatal("Https server error and stopped")
		}
	}
	hwlog.RunLog.Warn("enable unsafe http server")
	if err := s.Serve(limitLs); err != nil {
		hwlog.RunLog.Fatal("Http server error and stopped")
	}
}

func newServerAndListener() (*http.Server, net.Listener) {
	s := &http.Server{
		Addr: ip + ":" + strconv.Itoa(port),
		Handler: limiter.NewLimitHandlerWithMethod(concurrency, maxConcurrency, http.DefaultServeMux, true,
			http.MethodGet),
		ReadTimeout:    timeout * time.Second,
		WriteTimeout:   timeout * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
	}
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		hwlog.RunLog.Fatal(err)
	}
	limitLs, err := limiter.LimitListener(ln, maxConnection)
	if err != nil {
		hwlog.RunLog.Fatal(err)
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
	default:
		hwlog.RunLog.Fatal("invalid container mode setting")
	}
	if containerd != "" {
		opts.OciEndpoint = containerd
		opts.UserBackUp = false
	}
	if endpoint != "" {
		opts.CriEndpoint = endpoint
	}
	return opts
}

func regPrometheus(stop chan os.Signal, opts container.CntNpuMonitorOpts) *prometheus.Registry {
	deviceParser := container.MakeDevicesParser(opts)
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collector.NewNpuCollector(cacheTime, time.Duration(updateTime)*time.Second, stop, deviceParser),
	)
	return reg
}

func validate() {
	baseParamValid()
	if enableHTTP {
		return
	}
	if utils.CheckInterval < 1 || utils.CheckInterval > utils.WeekDays {
		hwlog.RunLog.Fatal("certificate check interval time invalidate")
	}
	if utils.WarningDays < utils.TenDays || utils.WarningDays > utils.YearDays {
		hwlog.RunLog.Fatal("certificate warning time invalidate")
	}
	// key file exist and need init kmc
	hwlog.RunLog.Info("start load imported certificate files")
	tlsCert, err := utils.LoadCertPair(certStore, keyStore, passFile, passFileBackUp, encryptAlgorithm)
	if err != nil {
		hwlog.RunLog.Fatal(err)
	}
	certificate = tlsCert
	loadCRL()
	caBytes, err = utils.CheckCaCert(caStore)
	if err != nil {
		hwlog.RunLog.Fatal(err)
	}
}

func baseParamValid() {
	if port < portLeft || port > portRight {
		hwlog.RunLog.Fatalf("the port is invalid")
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		hwlog.RunLog.Fatalf("the listen ip is invalid")
	}
	ip = parsedIP.String()
	hwlog.RunLog.Infof("listen on: %s", ip)
	if updateTime > oneMinute || updateTime < 1 {
		hwlog.RunLog.Fatalf("the updateTime is invalid")
	}
	if endpoint != "" {
		ep, err := utils.CheckPath(strings.TrimPrefix(endpoint, unixPre))
		if err != nil {
			hwlog.RunLog.Fatal(err)
		}
		endpoint = unixPre + ep
	}
	if containerd != "" {
		cnd, err := utils.CheckPath(strings.TrimPrefix(containerd, unixPre))
		if err != nil {
			hwlog.RunLog.Fatal(err)
		}
		containerd = unixPre + cnd
	}
	if enableHTTP {
		return
	}
	if encryptAlgorithm != utils.Aes128gcm && encryptAlgorithm != utils.Aes256gcm {
		hwlog.RunLog.Warn("reset invalid encryptAlgorithm ")
		encryptAlgorithm = utils.Aes256gcm
	}
	if tlsSuites != 0 && tlsSuites != 1 {
		hwlog.RunLog.Warn("reset invalid tlsSuites = 1 ")
		tlsSuites = 1
	}
	if tlsSuites == 0 {
		cipherSuites = tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
	} else {
		cipherSuites = tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	}
}

func init() {
	flag.IntVar(&port, "port", portConst,
		"The server port of the http service,range[1025-40000]")
	flag.StringVar(&ip, "ip", "",
		"The listen ip of the service,0.0.0.0 is not recommended when install on Multi-NIC host")
	flag.IntVar(&updateTime, "updateTime", updateTimeConst,
		"Interval (seconds) to update the npu metrics cache,range[1-60]")
	flag.BoolVar(&enableHTTP, "enableHTTP", false,
		"If true, the program will not check certificate files and enable http server (default false)")
	flag.IntVar(&encryptAlgorithm, "encryptAlgorithm", utils.Aes256gcm,
		"Use 8 for aes128gcm,9 for aes256gcm,not recommended config it in general")
	flag.IntVar(&tlsSuites, "tlsSuites", 1,
		"Use 0 for TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 ,1 for TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256")
	flag.BoolVar(&version, "version", false,
		"If true,query the version of the program (default false)")
	flag.StringVar(&containerMode, "containerMode", containerModeDocker,
		"Set 'docker' for monitoring docker containers or 'containerd' for CRI & containerd")
	flag.StringVar(&containerd, "containerd", "",
		"The endpoint of containerd used for listening containers' events")
	flag.StringVar(&endpoint, "endpoint", "",
		"The endpoint of the CRI  server to which will be connected")
	flag.IntVar(&concurrency, "concurrency", defaultConcurrency,
		"The max concurrency of the http server, range is [1-50]")

	// hwlog configuration
	flag.IntVar(&hwLogConfig.LogLevel, "logLevel", 0,
		"Log level, -1-debug, 0-info, 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal (default 0)")
	flag.IntVar(&hwLogConfig.MaxAge, "maxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup log files, must be greater than or equal to 7 days")
	flag.StringVar(&hwLogConfig.LogFileName, "logFile", defaultLogFile,
		"Log file path. If the file size exceeds 20MB, will be rotated")
	flag.IntVar(&hwLogConfig.MaxBackups, "maxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup log files, range is (0, 30]")
	flag.IntVar(&utils.CheckInterval, "checkInterval", utils.CheckInterval,
		"the Interval time for certificate validate period check, range is [1, 7]")
	flag.IntVar(&utils.WarningDays, "warningDays", utils.WarningDays,
		"the Ahead days of warning for certificate overdue, range is [10, 365]")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var proposal = "http"
	if certificate != nil {
		proposal = "https"
	}
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

func getCertStatus(w http.ResponseWriter, _ *http.Request) {
	b, err := json.Marshal(utils.CertificateMap)
	if err != nil {
		hwlog.RunLog.Error("fail to marshal cert status")
	}
	_, err = w.Write(b)
	if err != nil {
		hwlog.RunLog.Error("Write to response error")
	}
}

func loadCRL() {
	crlBytes, err := utils.CheckCRL(crlStore)
	if err != nil {
		hwlog.RunLog.Fatal(err)
	}
	if len(crlBytes) == 0 {
		return
	}
	crlList, err := x509.ParseCRL(crlBytes)
	if err != nil {
		hwlog.RunLog.Fatal("parse crlFile failed")
	}
	// skip check CRL update time when load it,only check when import CRL file
	if crlList != nil {
		crlcerList = crlList
		hwlog.RunLog.Infof("load CRL success")
	}
}

func initHwLogger(stopCh <-chan struct{}) {
	if err := hwlog.InitRunLogger(hwLogConfig, stopCh); err != nil {
		fmt.Printf("hwlog init failed, error is %v", err)
		os.Exit(-1)
	}
}
