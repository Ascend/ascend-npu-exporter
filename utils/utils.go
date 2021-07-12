//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.
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

// Package utils offer the some utils for certificate handling
package utils

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"huawei.com/kmc/pkg/adaptor/inbound/api"
	"huawei.com/kmc/pkg/adaptor/inbound/api/kmc"
	"huawei.com/kmc/pkg/adaptor/inbound/api/kmc/vo"
	"huawei.com/kmc/pkg/application/gateway"
	"huawei.com/kmc/pkg/application/gateway/loglevel"
	"huawei.com/npu-exporter/hwlog"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	rsaLength = 2048
	eccLength = 256
	maxLen    = 2048
	capacity  = 64
	byteSize  = 32
	dirMode   = 0700
	// FileMode file privilege
	FileMode = 0400
	// Aes128gcm AES128-GCM
	Aes128gcm = 8
	// Aes256gcm AES256-GCM
	Aes256gcm   = 9
	overdueTime = 100
	dayHours    = 24
)

var cryptoAPI api.CryptoApi

// Bootstrap kmc bootstrap
var Bootstrap *kmc.ManualBootstrap

// ReadBytes read contents from file path
func ReadBytes(path string) ([]byte, error) {
	key, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.New("the file path is invalid")
	}
	bytesData, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, errors.New("read file failed")
	}
	return bytesData, nil
}

// IsExists judge the file or directory exist or not
func IsExists(file string) bool {
	_, err := os.Stat(file)
	if err == nil {
		return true
	}
	if os.IsExist(err) {
		return true
	}
	return false
}

// ReadPassWd scan the screen and input the password info
func ReadPassWd() []byte {
	fmt.Print("Enter Private Key Password: ")
	bytePassword, err := terminal.ReadPassword(0)
	if err != nil {
		hwlog.Fatal("program error")
	}
	if len(bytePassword) > maxLen {
		hwlog.Fatal("input too long")
	}
	return bytePassword
}

// ParsePrivateKeyWithPassword  decode the private key
func ParsePrivateKeyWithPassword(keyBytes []byte, pd []byte) (*pem.Block, error) {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("decode key file failed")
	}
	buf := block.Bytes
	if x509.IsEncryptedPEMBlock(block) {
		if len(pd) == 0 {
			pd = ReadPassWd()
		}
		var err error
		buf, err = x509.DecryptPEMBlock(block, pd)
		if err != nil {
			if err == x509.IncorrectPasswordError {
				return nil, err
			}
			return nil, errors.New("cannot decode encrypted private keys")
		}
	} else {
		hwlog.Warn("detect that you provided private key is not encrypted")
	}
	return &pem.Block{
		Type:    block.Type,
		Headers: nil,
		Bytes:   buf,
	}, nil

}

// CheckCRL validate crl file
func CheckCRL(crlFile string) []byte {
	if crlFile == "" {
		return nil
	}
	crl, err := filepath.Abs(crlFile)
	if err != nil {
		hwlog.Fatalf("the crlFile is invalid")
	}
	if !IsExists(crl) {
		return nil
	}
	crlBytes, err := ioutil.ReadFile(crl)
	if err != nil {
		hwlog.Fatal("read crlFile failed")
	}
	_, err = x509.ParseCRL(crlBytes)
	if err != nil {
		hwlog.Fatal("parse crlFile failed")
	}
	return crlBytes
}

// MakeSureDir make sure the directory was existed
func MakeSureDir(path string) {
	dir := filepath.Dir(path)
	if !IsExists(dir) {
		err := os.Mkdir(dir, dirMode)
		if err != nil {
			hwlog.Fatal("create config directory failed")
		}
	}
}

// CheckValidityPeriod check certification validity period
func CheckValidityPeriod(cert *x509.Certificate) error {
	if time.Now().After(cert.NotAfter) || time.Now().Before(cert.NotBefore) {
		return errors.New("the certificate overdue ")
	}
	return nil
}

// CheckSignatureAlgorithm check signature algorithm of the certification
func CheckSignatureAlgorithm(cert *x509.Certificate) error {
	var signAl = cert.SignatureAlgorithm.String()
	if strings.Contains(signAl, "MD2") || strings.Contains(signAl, "MD5") ||
		strings.Contains(signAl, "SHA1") {
		return errors.New("the signature algorithm is unsafe,please use safe algorithm ")
	}
	hwlog.Info("signature algorithm validation passed")
	return nil
}

// GetPrivateKeyLength  return the length and type of private key
func GetPrivateKeyLength(cert *x509.Certificate, certificate *tls.Certificate) (int, string, error) {
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

// CheckRevokedCert check the revoked certification
func CheckRevokedCert(r *http.Request, revokedCertificates []pkix.RevokedCertificate) bool {
	for _, revokeCert := range revokedCertificates {
		for _, cert := range r.TLS.PeerCertificates {
			if cert != nil && cert.SerialNumber.Cmp(revokeCert.SerialNumber) == 0 {
				hwlog.Warnf("revoked certificate SN: %s", cert.SerialNumber)
				return true
			}
		}
	}
	return false
}

// LoadCertsFromPEM load the certification from pem
func LoadCertsFromPEM(pemCerts []byte) (*x509.Certificate, error) {
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

// ValidateX509Pair validate the x509pair
func ValidateX509Pair(certBytes []byte, keyBytes []byte) tls.Certificate {
	c, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		hwlog.Fatal("failed to load X509KeyPair")
	}
	cc, err := x509.ParseCertificate(c.Certificate[0])
	if err != nil {
		hwlog.Fatalf("parse certificate failed")
	}
	if err = CheckSignatureAlgorithm(cc); err != nil {
		hwlog.Fatalf(err.Error())
	}
	if err = CheckValidityPeriod(cc); err != nil {
		hwlog.Fatalf(err.Error())
	}
	keyLen, keyType, err := GetPrivateKeyLength(cc, &c)
	if err != nil {
		hwlog.Fatalf(err.Error())
	}
	// ED25519 private key length is stable and no need to verify
	if "RSA" == keyType && keyLen < rsaLength || "ECC" == keyType && keyLen < eccLength {
		hwlog.Warn("the private key length is not enough")
	}
	return c
}

// DecryptPrivateKeyWithPd  decrypt Private key By password
func DecryptPrivateKeyWithPd(keyFile string, passwd []byte) *pem.Block {
	keyBytes, err := ReadBytes(keyFile)
	if err != nil {
		hwlog.Fatal("read keyFile failed")
	}
	block, err := ParsePrivateKeyWithPassword(keyBytes, passwd)
	if err != nil {
		hwlog.Fatal(err)
	}
	return block
}

// GetRandomPass produce the new password
func GetRandomPass() []byte {
	k := make([]byte, byteSize, byteSize)
	if _, err := rand.Read(k); err != nil {
		hwlog.Error("get random words failed")
	}
	len := base64.RawStdEncoding.EncodedLen(byteSize)
	if len > capacity || len < byteSize {
		hwlog.Warn("the len of slice is abnormal")
	}
	dst := make([]byte, len, len)
	base64.RawStdEncoding.Encode(dst, k)
	return dst
}

// ReadOrUpdatePd  read or update the password file
func ReadOrUpdatePd(mainPath, backPath string) []byte {
	mainPd, err := ReadBytes(mainPath)
	if err != nil {
		hwlog.Warn("there is no main passwd,start to find backup files")
		backPd, err := ReadBytes(backPath)
		if err != nil {
			hwlog.Warn("there is no backup file found")
			return []byte{}
		}
		if err = ioutil.WriteFile(mainPath, backPd, FileMode); err != nil {
			hwlog.Warn("revert passwd failed")
		}
		return backPd
	}
	if err = ioutil.WriteFile(backPath, mainPd, FileMode); err != nil {
		hwlog.Warn("backup passwd failed")
	}
	return mainPd

}

// KmcInit init kmc component
func KmcInit(sdpAlgID int) {
	if Bootstrap == nil {
		defaultLogLevel := loglevel.Info
		var defaultLogger gateway.CryptoLogger = &hwlog.KmcLoggerApdaptor{}
		defaultInitConfig := vo.NewKmcInitConfigVO()
		defaultInitConfig.PrimaryKeyStoreFile = "/etc/npu-exporter/kmc_primary_store/master.ks"
		defaultInitConfig.StandbyKeyStoreFile = "/etc/npu-exporter/.config/backup.ks"
		if sdpAlgID == 0 {
			sdpAlgID = Aes256gcm
		}
		defaultInitConfig.SdpAlgId = sdpAlgID
		Bootstrap = kmc.NewManualBootstrap(0, defaultLogLevel, &defaultLogger, defaultInitConfig)
	}
	var err error
	cryptoAPI, err = Bootstrap.Start()
	if err != nil {
		hwlog.Fatal("initial kmc failed,please make sure the LD_LIBRARY_PATH include the kmc-ext.so ")
	}
}

// Encrypt encrypt the data
func Encrypt(domainID int, data []byte) ([]byte, error) {
	return cryptoAPI.EncryptByAppId(domainID, data)
}

// Decrypt decrypt the data
func Decrypt(domainID int, data []byte) ([]byte, error) {
	return cryptoAPI.DecryptByAppId(domainID, data)
}

// EncryptPrivateKeyAgain encrypt PrivateKey with local password again, and encrypted save password into files
func EncryptPrivateKeyAgain(keyBlock *pem.Block, passwdFile, passwdBackup string) (*pem.Block, error) {
	// generate new passwd for private key
	pd := GetRandomPass()
	KmcInit(0)
	encryptedPd, err := Encrypt(0, pd)
	if err != nil {
		hwlog.Fatal("encrypt passwd failed")
	}
	hwlog.Info("encrypt new passwd successfully")
	if err = ioutil.WriteFile(passwdFile, encryptedPd, FileMode); err != nil {
		hwlog.Fatal("write encrypted passwd to file failed")
	}
	hwlog.Info("create or update  passwd file successfully")
	if err = ioutil.WriteFile(passwdBackup, encryptedPd, FileMode); err != nil {
		hwlog.Fatal("write encrypted passwd to back file failed")
	}
	hwlog.Info("create or update  passwd backup file successfully")
	encryptedBlock, err := x509.EncryptPEMBlock(rand.Reader, keyBlock.Type, keyBlock.Bytes, pd, x509.PEMCipherAES256)
	if err != nil {
		hwlog.Fatal("encrypted private key failed")
	}
	hwlog.Info("encrypt private key by new passwd successfully")
	// clean password
	PaddingAndCleanSlice(pd)
	// wait certificate verify passed and then write key to file together
	if Bootstrap != nil {
		Bootstrap.Shutdown()
	}
	return encryptedBlock, nil
}

// PaddingAndCleanSlice fill slice wei zero
func PaddingAndCleanSlice(pd []byte) {
	for i := range pd {
		pd[i] = 0
	}
	pd = nil
}

// PeriodCheck  period check certificate
func PeriodCheck(cert *x509.Certificate) {
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
				hwlog.Warn("the certificate is already overdue")
				continue
			}
			gapHours := cert.NotAfter.Sub(now).Hours()
			overdueDays := gapHours / dayHours
			if overdueDays > math.MaxInt64 {
				overdueDays = math.MaxInt64
			}
			if overdueDays < overdueTime && overdueDays > 0 {
				hwlog.Warnf("the certificate will overdue after %d days later", int64(overdueDays))
			} else {
				hwlog.Error("the certificate was expired")
			}
		}
	}
}
