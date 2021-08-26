//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

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
	"huawei.com/npu-exporter/kmclog"
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
		hwlog.RunLog.Fatal("program error")
	}
	if len(bytePassword) > maxLen {
		hwlog.RunLog.Fatal("input too long")
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
		PaddingAndCleanSlice(pd)
		if err != nil {
			if err == x509.IncorrectPasswordError {
				return nil, err
			}
			return nil, errors.New("cannot decode encrypted private keys")
		}
	} else {
		hwlog.RunLog.Warn("detect that you provided private key is not encrypted")
	}
	return &pem.Block{
		Type:    block.Type,
		Headers: nil,
		Bytes:   buf,
	}, nil

}

// CheckCRL validate crl file
func CheckCRL(crlFile string) ([]byte, error) {
	if crlFile == "" {
		return nil, nil
	}
	crl, err := filepath.Abs(crlFile)
	if err != nil {
		return nil, errors.New("the crlFile is invalid")
	}
	if !IsExists(crl) {
		return nil, nil
	}
	crlBytes, err := ioutil.ReadFile(crl)
	if err != nil {
		return nil, errors.New("read crlFile failed")
	}
	_, err = x509.ParseCRL(crlBytes)
	if err != nil {
		return nil, errors.New("parse crlFile failed")
	}
	return crlBytes, nil
}

var osMkdirAll = os.MkdirAll

// MakeSureDir make sure the directory was existed
func MakeSureDir(path string) error {
	dir := filepath.Dir(path)
	if !IsExists(dir) {
		err := osMkdirAll(dir, dirMode)
		if err != nil {
			return errors.New("create config directory failed")
		}
	}
	return nil
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
		strings.Contains(signAl, "SHA1") || signAl == "0" {
		return errors.New("the signature algorithm is unsafe,please use safe algorithm ")
	}
	hwlog.RunLog.Info("signature algorithm validation passed")
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
func CheckRevokedCert(r *http.Request, crlcerList *pkix.CertificateList) bool {
	if crlcerList == nil || r.TLS == nil {
		hwlog.RunLog.Warnf("certificate or revokelist is nil")
		return false
	}
	revokedCertificates := crlcerList.TBSCertList.RevokedCertificates
	if len(revokedCertificates) == 0 {
		hwlog.RunLog.Warnf("revoked certificate length is 0")
		return false
	}
	// r.TLS.VerifiedChains [][]*x509.Certificate ,certificateChain[0] : current chain
	// certificateChain[0][0] : current certificate, certificateChain[0][1] :  certificate's issuer
	certificateChain := r.TLS.VerifiedChains
	if len(certificateChain) == 0 || len(certificateChain[0]) <= 1 {
		hwlog.RunLog.Warnf("VerifiedChains length is 0,or certificate is Cafile cannot revoke")
		return false
	}
	hwlog.RunLog.Infof("VerifiedChains length: %d,CertificatesChains length %d",
		len(certificateChain), len(certificateChain[0]))
	// CheckCRLSignature check CRL's issuer is certificate's issuer
	error := certificateChain[0][1].CheckCRLSignature(crlcerList)
	if error != nil {
		hwlog.RunLog.Warnf("CRL's issuer is not certificate's issuer")
		return false
	}
	for _, revokeCert := range revokedCertificates {
		for _, cert := range r.TLS.PeerCertificates {
			if cert.SerialNumber.Cmp(revokeCert.SerialNumber) == 0 {
				hwlog.RunLog.Warnf("revoked certificate SN: %s", cert.SerialNumber)
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
func ValidateX509Pair(certBytes []byte, keyBytes []byte) (*tls.Certificate, error) {
	c, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return nil, errors.New("failed to load X509KeyPair")
	}
	cc, err := x509.ParseCertificate(c.Certificate[0])
	if err != nil {
		return nil, errors.New("parse certificate failed")
	}
	if err = CheckSignatureAlgorithm(cc); err != nil {
		return nil, err
	}
	if err = CheckValidityPeriod(cc); err != nil {
		return nil, err
	}
	keyLen, keyType, err := GetPrivateKeyLength(cc, &c)
	if err != nil {
		return nil, err
	}
	// ED25519 private key length is stable and no need to verify
	if "RSA" == keyType && keyLen < rsaLength || "ECC" == keyType && keyLen < eccLength {
		hwlog.RunLog.Warn("the private key length is not enough")
	}
	return &c, nil
}

// DecryptPrivateKeyWithPd  decrypt Private key By password
func DecryptPrivateKeyWithPd(keyFile string, passwd []byte) (*pem.Block, error) {
	keyBytes, err := ReadBytes(keyFile)
	if err != nil {
		return nil, err
	}
	block, err := ParsePrivateKeyWithPassword(keyBytes, passwd)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// GetRandomPass produce the new password
func GetRandomPass() []byte {
	k := make([]byte, byteSize, byteSize)
	if _, err := rand.Read(k); err != nil {
		hwlog.RunLog.Error("get random words failed")
	}
	len := base64.RawStdEncoding.EncodedLen(byteSize)
	if len > capacity || len < byteSize {
		hwlog.RunLog.Warn("the len of slice is abnormal")
	}
	dst := make([]byte, len, len)
	base64.RawStdEncoding.Encode(dst, k)
	return dst
}

// ReadOrUpdatePd  read or update the password file
func ReadOrUpdatePd(mainPath, backPath string, mode os.FileMode) []byte {
	mainPd, err := ReadBytes(mainPath)
	if err != nil {
		hwlog.RunLog.Warn("there is no main passwd,start to find backup files")
		backPd, err := ReadBytes(backPath)
		if err != nil {
			hwlog.RunLog.Warn("there is no backup file found")
			return []byte{}
		}
		if err = ioutil.WriteFile(mainPath, backPd, mode); err != nil {
			hwlog.RunLog.Warn("revert passwd failed")
		}
		return backPd
	}
	if err = ioutil.WriteFile(backPath, mainPd, mode); err != nil {
		hwlog.RunLog.Warn("backup passwd failed")
	}
	return mainPd

}

// KmcInit init kmc component
var KmcInit = func(sdpAlgID int, primaryKey, standbyKey string) {
	if Bootstrap == nil {
		defaultLogLevel := loglevel.Info
		var defaultLogger gateway.CryptoLogger = &kmclog.KmcLoggerAdaptor{}
		defaultInitConfig := vo.NewKmcInitConfigVO()
		if primaryKey == "" {
			primaryKey = "/etc/mindx-dl/kmc_primary_store/master.ks"
		}
		if standbyKey == "" {
			standbyKey = "/etc//mindx-dl/.config/backup.ks"
		}
		defaultInitConfig.PrimaryKeyStoreFile = primaryKey
		defaultInitConfig.StandbyKeyStoreFile = standbyKey
		if sdpAlgID == 0 {
			sdpAlgID = Aes256gcm
		}
		defaultInitConfig.SdpAlgId = sdpAlgID
		Bootstrap = kmc.NewManualBootstrap(0, defaultLogLevel, &defaultLogger, defaultInitConfig)
	}
	var err error
	cryptoAPI, err = Bootstrap.Start()
	if err != nil {
		hwlog.RunLog.Fatal("initial kmc failed,please make sure the LD_LIBRARY_PATH include the kmc-ext.so ")
	}
}

// Encrypt encrypt the data
var Encrypt = func(domainID int, data []byte) ([]byte, error) {
	return cryptoAPI.EncryptByAppId(domainID, data)
}

// Decrypt decrypt the data
var Decrypt = func(domainID int, data []byte) ([]byte, error) {
	return cryptoAPI.DecryptByAppId(domainID, data)
}

// EncryptPrivateKeyAgain encrypt PrivateKey with local password again, and encrypted save password into files
func EncryptPrivateKeyAgain(key *pem.Block, psFile, psBkFile string, encrypt int) (*pem.Block, error) {
	// generate new passwd for private key
	pd := GetRandomPass()
	KmcInit(encrypt, "", "")
	encryptedPd, err := Encrypt(0, pd)
	if err != nil {
		hwlog.RunLog.Fatal("encrypt passwd failed")
	}
	hwlog.RunLog.Info("encrypt new passwd successfully")
	if err := OverridePassWdFile(psFile, encryptedPd, FileMode); err != nil {
		hwlog.RunLog.Fatal("write encrypted passwd to file failed")
	}
	hwlog.RunLog.Info("create or update  passwd file successfully")
	if err = OverridePassWdFile(psBkFile, encryptedPd, FileMode); err != nil {
		hwlog.RunLog.Fatal("write encrypted passwd to back file failed")
	}
	hwlog.RunLog.Info("create or update  passwd backup file successfully")
	encryptedBlock, err := x509.EncryptPEMBlock(rand.Reader, key.Type, key.Bytes, pd, x509.PEMCipherAES256)
	if err != nil {
		hwlog.RunLog.Fatal("encrypted private key failed")
	}
	hwlog.RunLog.Info("encrypt private key by new passwd successfully")
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
				hwlog.RunLog.Warn("the certificate is already overdue")
				continue
			}
			gapHours := cert.NotAfter.Sub(now).Hours()
			overdueDays := gapHours / dayHours
			if overdueDays > math.MaxInt64 {
				overdueDays = math.MaxInt64
			}
			if overdueDays < overdueTime && overdueDays > 0 {
				hwlog.RunLog.Warnf("the certificate will overdue after %d days later", int64(overdueDays))
			} else {
				hwlog.RunLog.Error("the certificate was expired")
			}
		}
	}
}

// OverridePassWdFile override password file with 0,1,random and then write new data
func OverridePassWdFile(path string, data []byte, mode os.FileMode) error {
	// Override with zero
	overrideByte := make([]byte, byteSize*maxLen, byteSize*maxLen)
	if err := write(path, overrideByte, mode); err != nil {
		return err
	}
	for i := range overrideByte {
		overrideByte[i] = 1
	}
	if err := write(path, overrideByte, mode); err != nil {
		return err
	}
	if _, err := rand.Read(overrideByte); err != nil {
		err := errors.New("get random words failed")
		return err
	}
	if err := write(path, overrideByte, mode); err != nil {
		return err
	}
	if err := write(path, data, mode); err != nil {
		return err
	}
	return nil
}

// CheckCaCert check the import ca cert
func CheckCaCert(caFile string) ([]byte, error) {
	if caFile == "" {
		return nil, nil
	}
	ca, err := filepath.Abs(caFile)
	if err != nil {
		return nil, errors.New("the caFile is invalid")
	}
	if !IsExists(caFile) {
		return nil, nil
	}
	caBytes, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, errors.New("failed to load caFile")
	}
	caCrt, err := LoadCertsFromPEM(caBytes)
	if err != nil {
		return nil, errors.New("convert ca certificate failed")
	}
	if !caCrt.IsCA {
		return nil, errors.New("this is not ca certificate")
	}
	err = CheckValidityPeriod(caCrt)
	if err != nil {
		return nil, errors.New("ca certificate is overdue")
	}
	if err = caCrt.CheckSignature(caCrt.SignatureAlgorithm, caCrt.RawTBSCertificate, caCrt.Signature); err != nil {
		return nil, errors.New("check ca certificate signature failed")
	}
	hwlog.RunLog.Infof("certificate signature check pass")
	return caBytes, nil
}

// LoadCertPair load and valid encrypted certificate and private key
func LoadCertPair(cert, key, psFile, psFileBk string, encryptAlgorithm int) (*tls.Certificate, error) {
	certBytes, err := ioutil.ReadFile(cert)
	if err != nil {
		return nil, errors.New("there is no certFile provided")
	}
	encodedPd := ReadOrUpdatePd(psFile, psFileBk, FileMode)
	KmcInit(encryptAlgorithm, "", "")
	pd, err := Decrypt(0, encodedPd)
	if err != nil {
		return nil, errors.New("decrypt passwd failed")
	}
	hwlog.RunLog.Info("decrypt passwd successfully")
	keyBlock, err := DecryptPrivateKeyWithPd(key, pd)
	if err != nil {
		return nil, err
	}
	hwlog.RunLog.Info("decrypt success")
	if Bootstrap != nil {
		Bootstrap.Shutdown()
	}
	PaddingAndCleanSlice(pd)
	// preload cert and key files
	c, err := ValidateX509Pair(certBytes, pem.EncodeToMemory(keyBlock))
	if err != nil {
		return nil, err
	}
	return c, nil
}

// NewTLSConfig  create the tls config struct
func NewTLSConfig(caBytes []byte, certificate tls.Certificate, cipherSuites uint16) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{cipherSuites},
	}
	if len(caBytes) > 0 {
		// Two-way SSL
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(caBytes); !ok {
			return nil, errors.New("append the CA file failed")
		}
		tlsConfig.ClientCAs = pool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		hwlog.RunLog.Info("enable Two-way SSL mode")
	} else {
		// One-way SSL
		tlsConfig.ClientAuth = tls.NoClientCert
		hwlog.RunLog.Info("enable One-way SSL mode")
	}
	return tlsConfig, nil
}

func write(path string, overideByte []byte, mode os.FileMode) error {
	if err := ioutil.WriteFile(path, overideByte, mode); err != nil {
		return errors.New("write encrypted key to config failed")
	}
	return nil
}
