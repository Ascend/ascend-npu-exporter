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
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"huawei.com/npu-exporter/collector"
	"huawei.com/npu-exporter/hwlog"
	"huawei.com/npu-exporter/utils"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

var (
	port             int
	updateTime       int
	needGoInfo       bool
	certFile         string
	keyFile          string
	caFile           string
	crlFile          string
	certificate      *tls.Certificate
	ip               string
	enableHTTP       bool
	caBytes          []byte
	encryptAlgorithm int
	version          bool
	tlsSuites        int
	cipherSuites     uint16
	concurrency      int
)

const (
	portConst          = 8082
	updateTimeConst    = 5
	cacheTime          = 65 * time.Second
	portLeft           = 1025
	portRight          = 40000
	oneMinute          = 60
	keyStore           = "/etc/npu-exporter/.config/config1"
	certStore          = "/etc/npu-exporter/.config/config2"
	caStore            = "/etc/npu-exporter/.config/config3"
	crlStore           = "/etc/npu-exporter/.config/config4"
	passFile           = "/etc/npu-exporter/.config/config5"
	passFileBackUp     = "/etc/npu-exporter/.conf"
	defaultConcurrency = 5
	defaultLogFile     = "/var/log/mindx-dl/npu-exporter/npu-exporter.log"
)

var revokedCertificates []pkix.RevokedCertificate
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
	validate()
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
	if certificate != nil {
		s.TLSConfig = newTLSConfig(caBytes)
		s.Handler = newLimitHandler(concurrency, interceptor(http.DefaultServeMux))
		hwlog.Info("start https server now...")
		if err := s.ListenAndServeTLS("", ""); err != nil {
			hwlog.Fatal("Https server error and stopped")
		}
	}
	hwlog.Warn("enable unsafe http server")
	if err := s.ListenAndServe(); err != nil {
		hwlog.Fatal("Http server error and stopped")
	}
}

func newTLSConfig(caBytes []byte) *tls.Config {
	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{*certificate},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{cipherSuites},
	}
	if len(caBytes) > 0 {
		// Two-way SSL
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(caBytes); !ok {
			hwlog.Fatalf("append the CA file failed")
		}
		tlsConfig.ClientCAs = pool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		hwlog.Info("enable Two-way SSL mode")
	} else {
		// One-way SSL
		tlsConfig.ClientAuth = tls.NoClientCert
		hwlog.Info("enable One-way SSL mode")
	}
	return tlsConfig
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
func validate() {
	if (certFile == "" && keyFile == "") && enableHTTP {
		baseParamValid()
		return
	}
	utils.KmcInit(encryptAlgorithm, "", "")
	// start to import certificate and keys
	importCertFiles(certFile, keyFile, caFile, crlFile)
	baseParamValid()
	// key file exist and need init kmc
	hwlog.Info("start load imported certificate files")
	tlsCert := handleCert()
	certificate = &tlsCert
	if certificate == nil {
		return
	}
	cc, err := x509.ParseCertificate(certificate.Certificate[0])
	if err != nil {
		hwlog.Fatal("parse certificate failed")
	}
	go utils.PeriodCheck(cc)
	loadCRL()
	checkCaCert(caStore)
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
	if encryptAlgorithm != utils.Aes128gcm && encryptAlgorithm != utils.Aes256gcm {
		encryptAlgorithm = utils.Aes256gcm
	}
	if tlsSuites != 0 && tlsSuites != 1 {
		tlsSuites = 0
	}
	if tlsSuites == 0 {
		cipherSuites = tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
	} else {
		cipherSuites = tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	}
}

func checkCaCert(caFile string) []byte {
	if caFile == "" {
		return nil
	}
	ca, err := filepath.Abs(caFile)
	if err != nil {
		hwlog.Fatalf("the caFile is invalid")
	}
	if !utils.IsExists(caFile) {
		return nil
	}
	caBytes, err = ioutil.ReadFile(ca)
	if err != nil {
		hwlog.Fatalf("failed to load caFile")
	}
	caCrt, err := utils.LoadCertsFromPEM(caBytes)
	if err != nil {
		hwlog.Fatal("convert ca certificate failed")
	}
	if !caCrt.IsCA {
		hwlog.Fatal("this is not ca certificate")
	}
	err = utils.CheckValidityPeriod(caCrt)
	if err != nil {
		hwlog.Fatal("ca certificate is overdue")
	}
	if err = caCrt.CheckSignature(caCrt.SignatureAlgorithm, caCrt.RawTBSCertificate, caCrt.Signature); err != nil {
		hwlog.Fatal("check ca certificate signature failed")
	}
	hwlog.Infof("certificate signature check pass")
	return caBytes
}

func handleCert() tls.Certificate {
	certBytes, err := ioutil.ReadFile(certStore)
	if err != nil {
		hwlog.Fatal("there is no certFile provided")
	}
	encodedPd := utils.ReadOrUpdatePd(passFile, passFileBackUp, utils.FileMode)
	pd, err := utils.Decrypt(0, encodedPd)
	if err != nil {
		hwlog.Info("decrypt passwd failed")
	}
	hwlog.Info("decrypt passwd successfully")
	keyBlock, err := utils.DecryptPrivateKeyWithPd(keyStore, pd)
	if err != nil {
		hwlog.Fatal(err)
	}
	hwlog.Info("decrypt success")
	utils.Bootstrap.Shutdown()
	utils.PaddingAndCleanSlice(pd)
	// preload cert and key files
	c, err := utils.ValidateX509Pair(certBytes, pem.EncodeToMemory(keyBlock))
	if err != nil {
		hwlog.Fatal(err)
	}
	return *c
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
	flag.StringVar(&caFile, "caFile", "", "The root certificate file path")
	flag.StringVar(&certFile, "certFile", "", "The certificate file path")
	flag.StringVar(&keyFile, "keyFile", "",
		"The key file path,If both the certificate and key file exist,system will enable https")
	flag.StringVar(&crlFile, "crlFile", "", "The offline CRL file path")
	flag.BoolVar(&enableHTTP, "enableHTTP", false,
		"If true, the program will not check certificate files and enable http server (default false)")
	flag.IntVar(&encryptAlgorithm, "encryptAlgorithm", utils.Aes256gcm,
		"Use 8 for aes128gcm,9 for aes256gcm")
	flag.IntVar(&tlsSuites, "tlsSuites", 1,
		"Use 0 for TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 ,1 for TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256")
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

func interceptor(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(revokedCertificates) > 0 && utils.CheckRevokedCert(r, revokedCertificates) {
			return
		}
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		h.ServeHTTP(w, r)
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(
		`<html>
			<head><title>NPU-Exporter</title></head>
			<body>
			<h1 align="center">NPU-Exporter</h1>
			<p align="center">Welcome to use NPU-Exporter,the Prometheus metrics url is http://ip ` + strconv.Itoa(port) + `/metrics: <a href="./metrics">Metrics</a></p>
			</body>
			</html>`))
	if err != nil {
		hwlog.Error("Write to response error")
	}
}

func loadCRL() {
	crlBytes, err := utils.CheckCRL(crlStore)
	if err != nil {
		hwlog.Fatal(err)
	}
	if len(crlBytes) == 0 {
		return
	}
	crlList, err := x509.ParseCRL(crlBytes)
	if err != nil {
		hwlog.Fatal("parse crlFile failed")
	}
	if crlList != nil {
		revokedCertificates = crlList.TBSCertList.RevokedCertificates
		hwlog.Infof("load CRL success")
	}
}

func importCertFiles(certFile, keyFile, caFile, crlFile string) {
	if certFile == "" && keyFile == "" && caFile == "" && crlFile == "" {
		hwlog.Info("no new certificate files need to be imported")
		return
	}
	if certFile == "" || keyFile == "" {
		hwlog.Fatal("need input certFile and keyFile together")
	}
	keyBlock, err := utils.DecryptPrivateKeyWithPd(keyFile, nil)
	if err != nil {
		hwlog.Fatal(err)
	}
	// start to import the  certificate file
	certBytes, err := utils.ReadBytes(certFile)
	if err != nil {
		hwlog.Fatal("read certFile failed")
	}
	// validate certification and private key, if not pass, program will exit
	if _, err = utils.ValidateX509Pair(certBytes, pem.EncodeToMemory(keyBlock)); err != nil {
		hwlog.Fatal(err)
	}
	// encrypt private key again with passwd
	encryptedBlock, err := utils.EncryptPrivateKeyAgain(keyBlock, passFile, passFileBackUp)
	if err = utils.MakeSureDir(keyStore); err != nil {
		hwlog.Fatal(err)
	}
	if err := utils.OverridePassWdFile(keyStore, pem.EncodeToMemory(encryptedBlock), utils.FileMode); err != nil {
		hwlog.Fatal(err)
	}
	if err = ioutil.WriteFile(certStore, certBytes, utils.FileMode); err != nil {
		hwlog.Fatal("write certBytes to config failed ")
	}
	// start to import the ca certificate file
	if caBytes = checkCaCert(caFile); len(caBytes) != 0 {
		if err = ioutil.WriteFile(caStore, caBytes, utils.FileMode); err != nil {
			hwlog.Fatal("write caBytes to config failed ")
		}
	}
	// start to import the crl file
	crlBytes, err := utils.CheckCRL(crlFile)
	if err != nil {
		hwlog.Fatal(err)
	}
	if len(crlBytes) != 0 {
		if err = ioutil.WriteFile(crlStore, crlBytes, utils.FileMode); err != nil {
			hwlog.Fatal("write crlBytes to config failed ")
		}
	}
	hwlog.Fatal("import certificate successfully")
}

func initHwLogger(stopCh <-chan struct{}) {
	if err := hwlog.Init(hwLogConfig, stopCh); err != nil {
		fmt.Printf("hwlog init failed, error is %v", err)
		os.Exit(-1)
	}
}
