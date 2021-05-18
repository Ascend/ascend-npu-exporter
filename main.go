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
	"crypto/tls"
	"crypto/x509"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"huawei.com/npu-exporter/collector"
	"io/ioutil"
	"k8s.io/klog"
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
	port        int
	updateTime  int
	needGoInfo  bool
	certFile    string
	keyFile     string
	caFile      string
	certificate *tls.Certificate
	ip          string
	enableHTTP  bool
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
	listenAddress := ip + ":" + strconv.Itoa(port)
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
	http.Handle("/", http.HandlerFunc(indexHandler))
	if certificate != nil {
		tlsConfig := &tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: []tls.Certificate{*certificate},
			MinVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
		}
		if caFile != "" {
			// Two-way SSL
			pool := x509.NewCertPool()
			crt, err := ioutil.ReadFile(caFile)
			if err != nil {
				klog.Fatalf("load CA file failed")
			}
			ok := pool.AppendCertsFromPEM(crt)
			if !ok {
				klog.Fatalf("append the CA file fialed")
			}
			tlsConfig.ClientCAs = pool
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			// One-way SSL
			tlsConfig.ClientAuth = tls.NoClientCert
		}

		s := &http.Server{
			Addr:      listenAddress,
			TLSConfig: tlsConfig,
			Handler:   interceptor(http.DefaultServeMux),
		}

		if err := s.ListenAndServeTLS("", ""); err != nil {
			klog.Fatal("Https server error and stopped")
		}
	}
	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		klog.Fatal("Http server error and stopped")
	}
}
func validate() {
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
	if (certFile == "" || keyFile == "") && enableHTTP {
		return
	}
	cert, err := filepath.Abs(certFile)
	if err != nil {
		klog.Fatalf("the certFile is invalid")
	}
	key, err := filepath.Abs(keyFile)
	if err != nil {
		klog.Fatalf("the keyFile is invalid")
	}
	certFile = cert
	keyFile = key
	// preload cert and key files
	c, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		klog.Fatalf("failed to load X509KeyPair")
	}
	certificate = &c
	if caFile == "" {
		return
	}
	ca, err := filepath.Abs(caFile)
	if err != nil {
		klog.Fatalf("the caFile is invalid")
	}
	caFile = ca
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
	flag.BoolVar(&enableHTTP, "enableHTTP", false,
		"If true, the program will not check certificate files and enable http server")
}

func interceptor(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		klog.Error("Write to response error")
	}
}
