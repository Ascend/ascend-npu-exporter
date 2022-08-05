//  Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.

// Package main
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/limiter"
	"huawei.com/npu-exporter/collector"
	"huawei.com/npu-exporter/collector/container"
	"huawei.com/npu-exporter/utils"
	"huawei.com/npu-exporter/versions"
)

var (
	port          int
	updateTime    int
	certificate   *tls.Certificate
	ip            string
	enableHTTP    bool
	caBytes       []byte
	version       bool
	tlsSuites     int
	cipherSuites  uint16
	concurrency   int
	containerMode string
	containerd    string
	endpoint      string
	crlcerList    *pkix.CertificateList
	warningDays   int
	checkInterval int
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
	defaultWarningDays      = 100
)

var hwLogConfig = &hwlog.LogConfig{LogFileName: defaultLogFile}

func main() {
	flag.Parse()
	if version {
		fmt.Printf("NPU-exporter version: %s \n", versions.BuildVersion)
		return
	}
	if err := initHwLogger(); err != nil {
		return
	}
	if err := validate(); err != nil {
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
	http.Handle("/v1/certstatus", http.HandlerFunc(getCertStatus))
	s, limitLs := newServerAndListener()
	if s == nil || limitLs == nil {
		return
	}
	if certificate != nil {
		tlsConf, err := utils.NewTLSConfig(caBytes, *certificate, cipherSuites)
		if err != nil {
			hwlog.RunLog.Error(err)
			return
		}
		s.TLSConfig = tlsConf
		s.Handler, err = limiter.NewLimitHandlerWithMethod(concurrency, maxConcurrency,
			utils.Interceptor(http.DefaultServeMux, crlcerList), true, http.MethodGet)
		if err != nil {
			hwlog.RunLog.Error(err)
			return
		}
		hwlog.RunLog.Info("start https server now...")
		if err = s.ServeTLS(limitLs, "", ""); err != nil {
			hwlog.RunLog.Error("Https server error and stopped")
			return
		}
	}
	hwlog.RunLog.Warn("enable unsafe http server")
	if err := s.Serve(limitLs); err != nil {
		hwlog.RunLog.Error("Http server error and stopped")
	}
}

func newServerAndListener() (*http.Server, net.Listener) {
	handler, err := limiter.NewLimitHandlerWithMethod(concurrency, maxConcurrency, http.DefaultServeMux, true,
		http.MethodGet)
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
	}
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, nil
	}
	limitLs, err := limiter.LimitListener(ln, maxConnection)
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

func validate() error {
	if err := baseParamValid(); err != nil {
		return err
	}
	if enableHTTP {
		return nil
	}
	if checkInterval < 1 || checkInterval > utils.WeekDays {
		return errors.New("certificate check interval time invalidate")
	}
	if warningDays < utils.TenDays || warningDays > utils.YearDays {
		return errors.New("certificate warning time invalidate")
	}
	utils.SetPeriodCheckParam(warningDays, checkInterval)
	// key file exist and need init kmc
	hwlog.RunLog.Info("start load imported certificate files")
	tlsCert, err := utils.LoadCertPair(certStore, keyStore, passFile, passFileBackUp, utils.Aes256gcm)
	if err != nil {
		return err
	}
	certificate = tlsCert
	if err = loadCRL(); err != nil {
		return err
	}
	caBytes, err = utils.CheckCaCert(caStore)
	return err
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
	if endpoint != "" {
		ep, err := utils.CheckPath(strings.TrimPrefix(endpoint, unixPre))
		if err != nil {
			return err
		}
		endpoint = unixPre + ep
	}
	if containerd != "" {
		cnd, err := utils.CheckPath(strings.TrimPrefix(containerd, unixPre))
		if err != nil {
			return err
		}
		containerd = unixPre + cnd
	}
	if enableHTTP {
		return nil
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
	return nil
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
	flag.IntVar(&checkInterval, "checkInterval", 1,
		"the Interval time for certificate validate period check, range is [1, 7]")
	flag.IntVar(&warningDays, "warningDays", defaultWarningDays,
		"the Ahead days of warning for certificate overdue, range is [10, 365]")
}

func indexHandler(w http.ResponseWriter, _ *http.Request) {
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
	b, err := json.Marshal(utils.GetCertStatus())
	if err != nil {
		hwlog.RunLog.Error("fail to marshal cert status")
	}
	_, err = w.Write(b)
	if err != nil {
		hwlog.RunLog.Error("Write to response error")
	}
}

func loadCRL() error {
	crlBytes, err := utils.CheckCRL(crlStore)
	if err != nil {
		return err
	}
	if len(crlBytes) == 0 {
		return nil
	}
	crlList, err := x509.ParseCRL(crlBytes)
	if err != nil {
		return errors.New("parse crlFile failed")
	}
	// skip check CRL update time when load it,only check when import CRL file
	if crlList != nil {
		crlcerList = crlList
		hwlog.RunLog.Infof("load CRL success")
	}
	return nil
}

func initHwLogger() error {
	if err := hwlog.InitRunLogger(hwLogConfig, context.Background()); err != nil {
		fmt.Printf("hwlog init failed, error is %#v", err)
		return err
	}
	return nil
}
