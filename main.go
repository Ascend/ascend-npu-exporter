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
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"huawei.com/kmc/pkg/adaptor/inbound/api"
	"huawei.com/kmc/pkg/adaptor/inbound/api/kmc"
	"huawei.com/kmc/pkg/adaptor/inbound/api/kmc/vo"
	"huawei.com/kmc/pkg/adaptor/outbound/log"
	"huawei.com/kmc/pkg/application/gateway"
	"huawei.com/kmc/pkg/application/gateway/loglevel"
	"huawei.com/npu-exporter/collector"
	"huawei.com/npu-exporter/utils"
	"io/ioutil"
	"k8s.io/klog"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
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
	aes128gcm          = 8
	aes256gcm          = 9
	rsaLength          = 2048
	eccLength          = 256
	fileMode           = 0400
	overdueTime        = 100
	dayHours           = 24
	keyStore           = "/etc/npu-exporter/.config/config1"
	certStore          = "/etc/npu-exporter/.config/config2"
	caStore            = "/etc/npu-exporter/.config/config3"
	crlStore           = "/etc/npu-exporter/.config/config4"
	defaultConcurrency = 5
)

var revokedCertificates []pkix.RevokedCertificate

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
		klog.Fatal("maxConcurrency parameter error")
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
	validate()
	listenAddress := ip + ":" + strconv.Itoa(port)
	klog.Infof("npu exporter starting and the version is %s", collector.BuildVersion)
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	reg := regPrometheus(stop)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
	http.Handle("/", http.HandlerFunc(indexHandler))
	s := &http.Server{
		Addr:    listenAddress,
		Handler: newLimitHandler(concurrency, http.DefaultServeMux),
	}
	if certificate != nil {
		tlsConfig := &tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: []tls.Certificate{*certificate},
			MinVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{cipherSuites},
		}
		if len(caBytes) > 0 {
			// Two-way SSL
			pool := x509.NewCertPool()
			ok := pool.AppendCertsFromPEM(caBytes)
			if !ok {
				klog.Fatalf("append the CA file failed")
			}
			tlsConfig.ClientCAs = pool
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
			klog.Info("enable Two-way SSL mode")
		} else {
			// One-way SSL
			tlsConfig.ClientAuth = tls.NoClientCert
			klog.Info("enable One-way SSL mode")
		}
		s.TLSConfig = tlsConfig
		s.Handler = newLimitHandler(concurrency, interceptor(http.DefaultServeMux))
		klog.Info("start https server now...")
		if err := s.ListenAndServeTLS("", ""); err != nil {
			klog.Fatal("Https server error and stopped")
		}
	}
	klog.Warning("enable unsafe http server")
	if err := s.ListenAndServe(); err != nil {
		klog.Fatal("Http server error and stopped")
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
func validate() {
	if version {
		fmt.Printf("NPU-exporter version: %s \n", collector.BuildVersion)
		os.Exit(0)
	}
	if (certFile == "" && keyFile == "") && enableHTTP {
		baseParamValid()
		return
	}
	kmcInit()
	// start to import certificate and keys
	importCertFiles(certFile, keyFile, caFile, crlFile)
	baseParamValid()
	// key file exist and need init kmc
	klog.Info("start load imported certificate files")
	tlsc := handleCert()
	certificate = &tlsc
	if certificate == nil {
		return
	}
	cc, err := x509.ParseCertificate(certificate.Certificate[0])
	if err != nil {
		klog.Fatal("parse certificate failed")
	}
	go periodCheck(cc)
	loadCRL()
	checkCaCert(caStore)
}

func baseParamValid() {
	if port < portLeft || port > portRight {
		klog.Fatalf("the port is invalid")
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		klog.Fatalf("the listen ip is invalid")
	}
	ip = parsedIP.String()
	klog.Infof("listen on: %s", ip)
	if updateTime > oneMinute || updateTime < 1 {
		klog.Fatalf("the updateTime is invalid")
	}
	if encryptAlgorithm != aes128gcm && encryptAlgorithm != aes256gcm {
		encryptAlgorithm = aes256gcm
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
		klog.Fatalf("the caFile is invalid")
	}
	if !utils.IsExists(caFile) {
		return nil
	}
	caBytes, err = ioutil.ReadFile(ca)
	if err != nil {
		klog.Fatalf("failed to load caFile")
	}
	caCrt, err := loadCertsFromPEM(caBytes)
	if err != nil {
		klog.Fatal("convert ca cert failed")
	}
	err = caCrt.CheckSignature(caCrt.SignatureAlgorithm, caCrt.RawTBSCertificate, caCrt.Signature)
	if err != nil {
		klog.Fatal("check ca certificate signature failed")
	}
	klog.Infof("certificate signature check pass")
	return caBytes
}

func handleCert() tls.Certificate {
	certBytes, err := ioutil.ReadFile(certStore)
	if err != nil {
		klog.Fatal("there is no certFile provided")
	}
	keyBytes, err := ioutil.ReadFile(keyStore)
	if err != nil {
		klog.Fatal("there is no keyFile provided")
	}
	keyBytes = handlePrivateKey(keyBytes)
	// preload cert and key files
	return validateX509Pair(certBytes, keyBytes)
}

func handlePrivateKey(keyBytes []byte) []byte {
	suffix := []byte("npu-exporter-encoded")
	if !bytes.HasSuffix(keyBytes, suffix) {
		klog.Fatal("npu-exporter config file invalid")
	}
	klog.Info("got Encrypted key file and start to decrypt")
	keyBytes = bytes.TrimSuffix(keyBytes, suffix)
	var err error
	keyBytes, err = decrypt(0, keyBytes)
	if err != nil {
		klog.Info("decrypt failed")
	}
	klog.Info("decrypt success")
	bootstrap.Shutdown()
	return keyBytes
}

func init() {
	klog.InitFlags(nil)
	flag.IntVar(&port, "port", portConst,
		"the server port of the http service,range[1025-40000]")
	flag.StringVar(&ip, "ip", "",
		"the listen ip of the service,0.0.0.0 is not recommended when install on Multi-NIC host")
	flag.IntVar(&updateTime, "updateTime", updateTimeConst,
		"Interval (seconds) to update the npu metrics cache,range[1-60]")
	flag.BoolVar(&needGoInfo, "needGoInfo", false,
		"If true,show golang metrics (default false)")
	flag.StringVar(&caFile, "caFile", "", "the root certificate file path")
	flag.StringVar(&certFile, "certFile", "", "the certificate file path")
	flag.StringVar(&keyFile, "keyFile", "",
		"the key file path,If both the certificate and key file exist,system will enable https")
	flag.StringVar(&crlFile, "crlFile", "", "the offline CRL file path")
	flag.BoolVar(&enableHTTP, "enableHTTP", false,
		"If true, the program will not check certificate files and enable http server")
	flag.IntVar(&encryptAlgorithm, "encryptAlgorithm", aes256gcm,
		"use 8 for aes128gcm,9 for aes256gcm(default)")
	flag.IntVar(&tlsSuites, "tlsSuites", 1,
		"use 0 for TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 ,1 for TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256")
	flag.BoolVar(&version, "version", false,
		"If true,query the version of the program")
	flag.IntVar(&concurrency, "concurrency", defaultConcurrency,
		"the max concurrency of the http server")
}

func interceptor(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(revokedCertificates) > 0 && checkRevokedCert(r) {
			return
		}
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		h.ServeHTTP(w, r)
	})
}

func checkRevokedCert(r *http.Request) bool {
	for _, revokeCert := range revokedCertificates {
		for _, cert := range r.TLS.PeerCertificates {
			if cert != nil && cert.SerialNumber.Cmp(revokeCert.SerialNumber) == 0 {
				klog.Warningf("revoked certificate SN: %s", cert.SerialNumber)
				return true
			}
		}
	}
	return false
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
		klog.Error("Write to response error")
	}
}

var cryptoAPI api.CryptoApi
var bootstrap *kmc.ManualBootstrap

func kmcInit() {
	defaultLogLevel := loglevel.Info
	var defaultLogger gateway.CryptoLogger = log.NewDefaultLogger()
	defaultInitConfig := vo.NewKmcInitConfigVO()
	defaultInitConfig.PrimaryKeyStoreFile = "/etc/npu-exporter/kmc_primary_store/master.ks"
	defaultInitConfig.StandbyKeyStoreFile = "/etc/npu-exporter/.config/backup.ks"
	defaultInitConfig.SdpAlgId = encryptAlgorithm
	bootstrap = kmc.NewManualBootstrap(0, defaultLogLevel, &defaultLogger, defaultInitConfig)
	var err error
	cryptoAPI, err = bootstrap.Start()
	if err != nil {
		klog.Fatal("initial kmc failed,please make sure the LD_LIBRARY_PATH include the kmc-ext.so ")
	}
}

func encrypt(domainID int, data []byte) ([]byte, error) {
	return cryptoAPI.EncryptByAppId(domainID, data)
}

func decrypt(domainID int, data []byte) ([]byte, error) {
	return cryptoAPI.DecryptByAppId(domainID, data)
}

func checkSignatureAlgorithm(cert *x509.Certificate) error {
	var signAl = cert.SignatureAlgorithm.String()
	if strings.Contains(signAl, "MD2") || strings.Contains(signAl, "MD5") ||
		strings.Contains(signAl, "SHA1") {
		return errors.New("the signature algorithm is unsafe,please use safe algorithm ")
	}
	klog.Info("signature algorithm validation passed")
	return nil
}

func checkValidDate(cert *x509.Certificate) error {
	if time.Now().After(cert.NotAfter) || time.Now().Before(cert.NotBefore) {
		return errors.New("the certificate overdue ")
	}
	return nil
}

func checkPrivateKeyLength(cert *x509.Certificate, certificate *tls.Certificate) (int, string, error) {
	if certificate == nil {
		return 0, "", errors.New("certificate is nil")
	}
	switch cert.PublicKey.(type) {
	case *rsa.PublicKey:
		priv, ok := certificate.PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return 0, "RSA", errors.New("get rsa key length failed")
		}
		return priv.N.BitLen(), "RSA", nil
	case *ecdsa.PublicKey:
		priv, ok := certificate.PrivateKey.(*ecdsa.PrivateKey)
		if !ok {
			return 0, "ECC", errors.New("get ecdsa key length failed")
		}
		return priv.X.BitLen(), "ECC", nil
	case ed25519.PublicKey:
		priv, ok := certificate.PrivateKey.(ed25519.PrivateKey)
		if !ok {
			return 0, "ED25519", errors.New("get ed25519 key length failed")
		}
		return len(priv.Public().(ed25519.PublicKey)), "ED25519", nil
	default:
		return 0, "", errors.New("get key length failed")
	}
}

func loadCertsFromPEM(pemCerts []byte) (*x509.Certificate, error) {
	if len(pemCerts) <= 0 {
		return nil, errors.New("wrong input")
	}
	var block *pem.Block
	block, pemCerts = pem.Decode(pemCerts)
	if block == nil {
		return nil, errors.New("parse cert failed")
	}
	if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
		return nil, errors.New("invalid cert bytes")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.New("parse cert failed")
	}
	return cert, nil
}

func periodCheck(cert *x509.Certificate) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			now := time.Now()
			if now.After(cert.NotAfter) || now.Before(cert.NotBefore) {
				klog.Warning("the certificate is already overdue")
				continue
			}
			gapHours := cert.NotAfter.Sub(now).Hours()
			overdueDays := gapHours / dayHours
			if overdueDays > math.MaxInt64 {
				overdueDays = math.MaxInt64
			}
			if overdueDays < overdueTime {
				klog.Warningf("the certificate will overdue after %d days later", int64(overdueDays))
			}
		}
	}
}

func loadCRL() {
	crlBytes := utils.CheckCRL(crlStore)
	if len(crlBytes) == 0 {
		return
	}
	crlList, err := x509.ParseCRL(crlBytes)
	if err != nil {
		klog.Fatal("parse crlFile failed")
	}
	if crlList != nil {
		revokedCertificates = crlList.TBSCertList.RevokedCertificates
		klog.Infof("load CRL success")
	}
}

func importCertFiles(certFile, keyFile, caFile, crlFile string) {
	if certFile == "" && keyFile == "" && caFile == "" && crlFile == "" {
		klog.Info("no new certificate files need to be imported")
		return
	}
	if certFile == "" || keyFile == "" {
		klog.Fatal("need input certFile and keyFile together")
	}
	keyBytes, encodeKey := getPrivateBytes(keyFile)
	// wait certificate verify passed and then write key to file together
	if bootstrap != nil {
		bootstrap.Shutdown()
	}
	// start to import the  certificate file
	certBytes := checkX509Pair(certFile, keyBytes)
	utils.MakeSureDir(keyStore)
	err := ioutil.WriteFile(keyStore, encodeKey, fileMode)
	if err != nil {
		klog.Fatalf("write encrypted key to config failed:%v ", err)
	}
	err = ioutil.WriteFile(certStore, certBytes, fileMode)
	if err != nil {
		klog.Fatal("write certBytes to config failed ")
	}
	// start to import the ca certificate file
	caBytes = checkCaCert(caFile)
	if len(caBytes) != 0 {
		err = ioutil.WriteFile(caStore, caBytes, fileMode)
		if err != nil {
			klog.Fatal("write caBytes to config failed ")
		}
	}
	// start to import the crl file
	crlBytes := utils.CheckCRL(crlFile)
	if len(crlBytes) != 0 {
		err = ioutil.WriteFile(crlStore, crlBytes, fileMode)
		if err != nil {
			klog.Fatal("write crlBytes to config failed ")
		}
	}
	klog.Fatal("import certificate successfully")
}

func getPrivateBytes(keyFile string) ([]byte, []byte) {
	pd := utils.ReadPassWd()
	keyBytes, err := utils.ReadBytes(keyFile)
	if err != nil {
		klog.Fatal("read keyFile failed")
	}
	suffix := []byte("npu-exporter-encoded")
	keyBytes, err = utils.ParsePrivateKeyWithPassword(keyBytes, []byte(pd))
	if err != nil {
		klog.Fatal(err)
	}
	encodeKey, err := encrypt(0, keyBytes)
	if err != nil {
		klog.Warning("encrypt failed")
	}
	encodeKey = append(encodeKey, suffix...)
	return keyBytes, encodeKey
}

func checkX509Pair(certFile string, keyBytes []byte) (cert []byte) {
	certBytes, err := utils.ReadBytes(certFile)
	if err != nil {
		klog.Fatal("read certFile failed")
	}
	validateX509Pair(certBytes, keyBytes)
	return certBytes
}

func validateX509Pair(certBytes []byte, keyBytes []byte) tls.Certificate {
	c, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		klog.Fatal("failed to load X509KeyPair")
	}
	cc, err := x509.ParseCertificate(c.Certificate[0])
	if err != nil {
		klog.Fatalf("parse certificate failed")
	}
	err = checkSignatureAlgorithm(cc)
	if err != nil {
		klog.Fatalf(err.Error())
	}
	err = checkValidDate(cc)
	if err != nil {
		klog.Fatalf(err.Error())
	}
	keyLen, keyType, err := checkPrivateKeyLength(cc, &c)
	if err != nil {
		klog.Fatalf(err.Error())
	}
	// ED25519 private key length is stable and no need to verify
	if "RSA" == keyType && keyLen < rsaLength || "ECC" == keyType && keyLen < eccLength {
		klog.Warning("the private key length is not enough")
	}
	return c
}
