// Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

// Package cert for web certificate util
// 1. load cert, ca, crl
// 2. return http client/server for one/two way auth
package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"huawei.com/npu-exporter/hwlog"
	"huawei.com/npu-exporter/utils"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	// OneWay one way auth
	OneWay = "one"
	// TwoWay two way auth
	TwoWay = "two"
	// Overdue Overdue
	Overdue = "overdue"
	// ReApply ReApply
	ReApply = "reapply"
)

// CertificateUtils CertificateUtils
type CertificateUtils struct {
	PrivateKey *rsa.PrivateKey
	CrlList    *pkix.CertificateList
	Cert       *tls.Certificate
	Lock       sync.RWMutex

	CaBytes     []byte
	OverdueTime int
	DNSName     string

	KeyStore       string
	CertStore      string
	CaStore        string
	CrlStore       string
	PassFile       string
	PassFileBackUp string
}

// NewCertificateUtils NewCertificateUtils
func NewCertificateUtils(dnsName, dirName string, overdueTime int) *CertificateUtils {
	cu := &CertificateUtils{
		Lock:           sync.RWMutex{},
		OverdueTime:    overdueTime,
		DNSName:        dnsName,
		KeyStore:       dirName + ".config/config1",
		CertStore:      dirName + ".config/config2",
		CaStore:        dirName + ".config/config3",
		CrlStore:       dirName + ".config/config4",
		PassFile:       dirName + ".config/config5",
		PassFileBackUp: dirName + ".conf",
	}
	cu.ClearCertificateMap()

	return cu
}

// ClearCertificateMap ClearCertificateMap
func (cu *CertificateUtils) ClearCertificateMap() {
	utils.CertificateMap = map[string]*utils.CertStatus{}
}

// LoadCertAndKeyOnStart load cert and private key on start
func (cu *CertificateUtils) LoadCertAndKeyOnStart(algorithm int) (*tls.Certificate, error) {
	pathMap := map[string]string{
		utils.CertStorePath:      cu.CertStore,
		utils.KeyStorePath:       cu.KeyStore,
		utils.PassFilePath:       cu.PassFile,
		utils.PassFileBackUpPath: cu.PassFileBackUp,
	}
	certByte, keyByte, err := utils.LoadCertPairByte(pathMap, algorithm, utils.RWMode)
	if err == os.ErrNotExist {
		err = errors.New("needs to reapply certification")
		hwlog.RunLog.Error(err)
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	tlsCert, err := utils.ValidateCertPair(certByte, keyByte, false, cu.OverdueTime)
	if err != nil {
		if strings.Contains(err.Error(), Overdue) {
			err = errors.New("needs to reapply certification")
			hwlog.RunLog.Error(err)
			return nil, err
		}
		return nil, err
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(keyByte)
	if err != nil {
		return nil, err
	}

	cu.PrivateKey = privateKey
	cu.Cert = tlsCert
	return tlsCert, nil
}

// LoadCRLOnStart load crlList and private key on start
func (cu *CertificateUtils) LoadCRLOnStart() (*pkix.CertificateList, error) {
	crlBytes, err := utils.LoadFile(cu.CrlStore)
	if err != nil {
		return nil, err
	}
	if crlBytes == nil {
		return nil, nil
	}
	crlCertList, err := utils.ValidateCRL(crlBytes)
	if err != nil {
		return nil, err
	}

	cu.CrlList = crlCertList
	return crlCertList, nil
}

// LoadCAOnStart load ca on start
func (cu *CertificateUtils) LoadCAOnStart(authMode string) ([]byte, error) {
	caBytes, err := utils.CheckCaCertV2(cu.CaStore, cu.OverdueTime)
	// if the ca is overdue, needs to update in future
	if err != nil {
		return nil, err
	}
	if len(caBytes) == 0 && authMode == TwoWay {
		// reapply ca
		err = errors.New("needs to reapply ca certification")
		hwlog.RunLog.Error(err)
		return nil, err
	}

	cu.CaBytes = caBytes
	return caBytes, nil
}

// GetCertificateFunc return func get cert
func (cu *CertificateUtils) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		cu.Lock.RLock()
		defer cu.Lock.RUnlock()
		return cu.Cert, nil
	}
}

// GeneratePrivateKey generate private key
func (cu *CertificateUtils) GeneratePrivateKey(privateKeyLength int, algorithm int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, privateKeyLength)
	if err != nil {
		return nil, err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	encryptedBlock, err := utils.EncryptPrivateKeyAgainWithMode(keyBlock, cu.PassFile, cu.PassFileBackUp, algorithm,
		utils.RWMode)
	if err = utils.OverridePassWdFile(cu.KeyStore, pem.EncodeToMemory(encryptedBlock), utils.RWMode); err != nil {
		return nil, err
	}
	hwlog.RunLog.Info("successed in creating private key")
	cu.PrivateKey = privateKey
	return privateKey, nil
}

// GetTwoWayAuthRequestClient get http client for two way auth
func (cu *CertificateUtils) GetTwoWayAuthRequestClient() (*http.Client, error) {
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(cu.CaBytes); !ok {
		return nil, errors.New("tls config append the CA file failed")
	}
	cu.Lock.RLock()
	defer cu.Lock.RUnlock()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{*cu.Cert},
				RootCAs:      pool,
			},
		},
	}

	return client, nil
}
