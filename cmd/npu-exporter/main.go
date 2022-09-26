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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/limiter"
	mytls "huawei.com/mindx/common/tls"
	"huawei.com/mindx/common/utils"
	myx509 "huawei.com/mindx/common/x509"
	"huawei.com/npu-exporter/collector"
	"huawei.com/npu-exporter/collector/container"
	"huawei.com/npu-exporter/versions"
)

var (
	port           int
	updateTime     int
	certificate    *tls.Certificate
	ip             string
	enableHTTP     bool
	caBytes        []byte
	version        bool
	tlsSuites      int
	cipherSuites   uint16
	concurrency    int
	containerMode  string
	containerd     string
	endpoint       string
	crlcerList     *pkix.CertificateList
	warningDays    int
	checkInterval  int
	limitIPReq     string
	limitIPConn    int
	limitTotalConn int
	cacheSize      int
)

const (
	dirPrefix               = "/etc/mindx-dl/npu-exporter/"
	portConst               = 8082
	updateTimeConst         = 5
	cacheTime               = 65 * time.Second
	portLeft                = 1025
	portRight               = 40000
	oneMinute               = 60
	keyStore                = dirPrefix + mytls.KeyStore
	keyBkpStore             = dirPrefix + mytls.KeyBackup
	certStore               = dirPrefix + mytls.CertStore
	certBkpStore            = dirPrefix + mytls.CertBackup
	caStore                 = dirPrefix + mytls.CaStore
	caBkpStore              = dirPrefix + mytls.CaBackup
	crlStore                = dirPrefix + mytls.CrlStore
	crlBkpStore             = dirPrefix + mytls.CrlBackup
	passFile                = dirPrefix + mytls.PassFile
	passFileBackUp          = dirPrefix + mytls.PassFileBackUp
	defaultConcurrency      = 5
	defaultLogFile          = "/var/log/mindx-dl/npu-exporter/npu-exporter.log"
	containerModeDocker     = "docker"
	containerModeContainerd = "containerd"
	unixPre                 = "unix://"
	timeout                 = 10
	maxHeaderBytes          = 1024
	defaultWarningDays      = 100
	// weekDays one week days
	weekDays = 7
	// yearDays one year days
	yearDays = 365
	// tenDays ten days
	tenDays           = 10
	aes256gcm         = 9
	maxIPConnLimit    = 128
	maxConcurrency    = 512
	defaultConnection = 20
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
	conf := initConfig()
	s, limitLs := newServerAndListener(conf)
	if s == nil || limitLs == nil {
		return
	}
	if certificate != nil {
		tlsConf, err := mytls.NewTLSConfig(caBytes, *certificate, []uint16{cipherSuites})
		if err != nil {
			hwlog.RunLog.Error(err)
			return
		}
		s.TLSConfig = tlsConf
		s.Handler, err = limiter.NewLimitHandlerV2(myx509.Interceptor(http.DefaultServeMux, crlcerList), conf)
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
	if checkInterval < 1 || checkInterval > weekDays {
		return errors.New("certificate check interval time invalidate")
	}
	if warningDays < tenDays || warningDays > yearDays {
		return errors.New("certificate warning time invalidate")
	}
	myx509.SetPeriodCheckParam(warningDays, checkInterval)
	// key file exist and need init kmc
	hwlog.RunLog.Info("start load imported certificate files")
	pathMap := map[string]string{
		myx509.CertStorePath:       certStore,
		myx509.CertStoreBackupPath: certBkpStore,
		myx509.KeyStorePath:        keyStore,
		myx509.KeyStoreBackupPath:  keyBkpStore,
		myx509.PassFilePath:        passFile,
		myx509.PassFileBackUpPath:  passFileBackUp,
	}
	tlsCert, err := mytls.LoadCertPair(pathMap, aes256gcm)
	if err != nil {
		return err
	}
	certificate = tlsCert
	if err = loadCRL(); err != nil {
		return err
	}
	if err = loadCA(); err != nil {
		return err
	}
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

func containerSockCheck() error {
	if endpoint != "" && !strings.Contains(endpoint, unixPre) {
		endpoint = unixPre + endpoint
	}
	if containerd != "" && !strings.Contains(containerd, unixPre) {
		containerd = unixPre + containerd
	}
	if endpoint != "" && !strings.Contains(endpoint, ".sock") {
		return errors.New("endpoint file not sock address")
	}
	if containerd != "" && !strings.Contains(containerd, ".sock") {
		return errors.New("containerd file not sock address")
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
	flag.IntVar(&checkInterval, "checkInterval", 1,
		"the Interval time for certificate validate period check, range is [1, 7]")
	flag.IntVar(&warningDays, "warningDays", defaultWarningDays,
		"the Ahead days of warning for certificate overdue, range is [10, 365]")
	flag.IntVar(&cacheSize, "cacheSize", limiter.DefaultCacheSize, "the cacheSize for ip limit,"+
		"range  is [1,1024000],keep default normally")
	flag.IntVar(&limitIPConn, "limitIPConn", defaultConcurrency, "the tcp connection limit for each Ip,"+
		"range  is [1,128]")
	flag.IntVar(&limitTotalConn, "limitTotalConn", defaultConnection, "the tcp connection limit for all"+
		" request,range  is [1,512]")
	flag.StringVar(&limitIPReq, "limitIPReq", "20/1",
		"the http request limit counts for each Ip,20/1 means allow 20 request in 1 seconds")
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
	b, err := json.Marshal(myx509.GetCertStatus())
	if err != nil {
		hwlog.RunLog.Error("fail to marshal cert status")
	}
	_, err = w.Write(b)
	if err != nil {
		hwlog.RunLog.Error("Write to response error")
	}
}

func loadCRL() error {
	crlInstance, err := myx509.NewBKPInstance(nil, crlStore, crlBkpStore)
	if err != nil {
		return err
	}
	crlBytes, err := crlInstance.ReadFromDisk(utils.FileMode, false)
	if err != nil || crlBytes == nil {
		hwlog.RunLog.Info("no crl file found")
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

func loadCA() error {
	caInstance, err := myx509.NewBKPInstance(nil, caStore, caBkpStore)
	if err != nil {
		return err
	}
	caBytes, err = caInstance.ReadFromDisk(utils.FileMode, false)
	if err != nil || len(caBytes) == 0 {
		hwlog.RunLog.Info("no ca file found")
		return nil
	}
	return myx509.VerifyCaCert(caBytes, myx509.InvalidNum)
}

func initHwLogger() error {
	if err := hwlog.InitRunLogger(hwLogConfig, context.Background()); err != nil {
		fmt.Printf("hwlog init failed, error is %#v", err)
		return err
	}
	return nil
}
