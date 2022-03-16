//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package utils offer the some utils for certificate handling
package utils

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
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
	"net"
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
	// RWMode for read and write
	RWMode = 0600
	// Aes128gcm AES128-GCM
	Aes128gcm = 8
	// Aes256gcm AES256-GCM
	Aes256gcm = 9

	dayHours = 24
	x509v3   = 3
	// InvalidNum invalid num
	InvalidNum = -9999999

	// KeyStore KeyStore path
	KeyStore = ".config/config1"
	// CertStore CertStore path
	CertStore = ".config/config2"
	// CaStore CaStore path
	CaStore = ".config/config3"
	// CrlStore CrlStore path
	CrlStore = ".config/config4"
	// PassFile PassFile path
	PassFile = ".config/config5"
	// PassFileBackUp PassFileBackUp path
	PassFileBackUp = ".conf"
	// KubeCfgFile kubeconfig file store path
	KubeCfgFile = ".config/config6"
	// KeyStorePath KeyStorePath
	KeyStorePath = "KeyStorePath"
	// CertStorePath CertStorePath
	CertStorePath = "CertStorePath"
	// PassFilePath PassFilePath
	PassFilePath = "PassFilePath"
	// PassFileBackUpPath PassFileBackUpPath
	PassFileBackUpPath = "PassFileBackUpPath"
	yearHours          = 87600
	maskLen            = 2
	// WeekDays one week days
	WeekDays = 7
	// YearDays one year days
	YearDays = 365
	// TenDays ten days
	TenDays = 10
	// Size10M  bytes of 10M
	Size10M = 10 * 1024 * 1024
	maxSize = 1024 * 1024 * 1024
)

var (
	cryptoAPI api.CryptoApi
	// Bootstrap kmc bootstrap
	Bootstrap *kmc.ManualBootstrap
	// CertificateMap  using certificate information
	CertificateMap = make(map[string]*CertStatus, 4)
	// WarningDays cert warning day ,unit days
	WarningDays = 100
	// CheckInterval  cert period check interval,unit days
	CheckInterval = 1
)

// CertStatus  the certificate valid period
type CertStatus struct {
	NotBefore         time.Time `json:"not_before"`
	NotAfter          time.Time `json:"not_after"`
	IsCA              bool      `json:"is_ca"`
	FingerprintSHA256 string    `json:"fingerprint_sha256,omitempty"`
	FingerprintSHA1   string    `json:"fingerprint_sha1,omitempty"`
	FingerprintMD5    string    `json:"fingerprint_md5,omitempty"`
}

// ReadBytes read contents from file path
// Deprecated: replace with ReadLimitBytes
func ReadBytes(path string) ([]byte, error) {
	key, err := CheckPath(path)
	if err != nil {
		return nil, err
	}
	bytesData, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, errors.New("read file failed")
	}
	return bytesData, nil
}

// ReadLimitBytes read limit length of contents from file path
func ReadLimitBytes(path string, limitLength int) ([]byte, error) {
	key, err := CheckPath(path)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(key, os.O_RDONLY, FileMode)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	if limitLength < 0 || limitLength > maxSize {
		return nil, errors.New("the limit length is not valid")
	}
	buf := make([]byte, limitLength, limitLength)
	l, err := file.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[0:l], nil
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
	crlBytes, err := LoadFile(crlFile)
	if err != nil {
		return nil, err
	}
	if crlBytes == nil {
		return nil, nil
	}
	_, err = ValidateCRL(crlBytes)
	if err != nil {
		return nil, err
	}

	return crlBytes, nil
}

// ValidateCRL ValidateCRL
func ValidateCRL(crlBytes []byte) (*pkix.CertificateList, error) {
	crlList, err := x509.ParseCRL(crlBytes)
	if err != nil {
		return nil, errors.New("parse crlFile failed")
	}
	if time.Now().Before(crlList.TBSCertList.ThisUpdate) || time.Now().After(crlList.TBSCertList.NextUpdate) {
		return nil, errors.New("crlFile update time not match")
	}

	return crlList, nil
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
	overdueDays, err := GetValidityPeriod(cert)
	if err != nil {
		return err
	}
	if overdueDays < float64(WarningDays) && overdueDays > 0 {
		hwlog.RunLog.Warnf("the certificate will overdue after %d days later", int64(overdueDays))
	}

	return nil
}

// CheckValidityPeriodWithError if the time expires, an error is reported
func CheckValidityPeriodWithError(cert *x509.Certificate, overdueTime int) error {
	overdueDays, err := GetValidityPeriod(cert)
	if err != nil {
		return err
	}
	if overdueDays <= float64(overdueTime) {
		return fmt.Errorf("overdueDayes is (%v) need to update certification", overdueDays)
	}

	return nil
}

// GetValidityPeriod get certification validity period
func GetValidityPeriod(cert *x509.Certificate) (float64, error) {
	now := time.Now()
	if now.After(cert.NotAfter) || now.Before(cert.NotBefore) {
		return 0, errors.New("the certificate overdue ")
	}
	if cert.NotAfter.Sub(cert.NotBefore).Hours() > yearHours {
		hwlog.RunLog.Warn("the certificate valid period is more than 10 years")
	}
	gapHours := cert.NotAfter.Sub(now).Hours()
	overdueDays := gapHours / dayHours
	if overdueDays > math.MaxInt64 {
		overdueDays = math.MaxInt64
	}

	return overdueDays, nil
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
	if err := certificateChain[0][1].CheckCRLSignature(crlcerList); err != nil {
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
	return ValidateX509PairV2(certBytes, keyBytes, InvalidNum)
}

// ValidateX509PairV2 validate the x509pair version 2
func ValidateX509PairV2(certBytes []byte, keyBytes []byte, overdueTime int) (*tls.Certificate, error) {
	c, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return nil, errors.New("failed to load X509KeyPair")
	}
	cc, err := x509.ParseCertificate(c.Certificate[0])
	if err != nil {
		return nil, errors.New("parse certificate failed")
	}
	if err = checkExtension(cc); err != nil {
		return nil, err
	}
	if err = CheckSignatureAlgorithm(cc); err != nil {
		return nil, err
	}
	switch overdueTime {
	case InvalidNum:
		err = CheckValidityPeriod(cc)
	default:
		err = CheckValidityPeriodWithError(cc, overdueTime)
	}
	if err != nil {
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
	keyBytes, err := ReadLimitBytes(keyFile, Size10M)
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
	mainPd, err := ReadLimitBytes(mainPath, Size10M)
	if err != nil {
		hwlog.RunLog.Warn("there is no main passwd,start to find backup files")
		backPd, err := ReadLimitBytes(backPath, Size10M)
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
			standbyKey = "/etc/mindx-dl/.config/backup.ks"
		}
		checkRootMaterial(primaryKey)
		checkRootMaterial(standbyKey)
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
	cryptoAPI.UpdateLifetimeDays(TenDays * TenDays)
	if err != nil {
		hwlog.RunLog.Fatal("initial kmc failed,please make sure the LD_LIBRARY_PATH include the kmc-ext.so ")
	}
}

func checkRootMaterial(primaryKey string) {
	if IsExists(primaryKey) {
		_, err := CheckPath(primaryKey)
		if err != nil {
			hwlog.RunLog.Fatal("kmc root material file is symlinks")
		}
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
	return EncryptPrivateKeyAgainWithMode(key, psFile, psBkFile, encrypt, FileMode)
}

// EncryptPrivateKeyAgainWithMode encrypt privatekey again with mode
func EncryptPrivateKeyAgainWithMode(key *pem.Block, psFile, psBkFile string, encrypt int, mode os.FileMode) (*pem.Block,
	error) {
	// generate new passwd for private key
	pd := GetRandomPass()
	KmcInit(encrypt, "", "")
	encryptedPd, err := Encrypt(0, pd)
	if err != nil {
		hwlog.RunLog.Fatal("encrypt passwd failed")
	}
	hwlog.RunLog.Info("encrypt new passwd successfully")
	if err := OverridePassWdFile(psFile, encryptedPd, mode); err != nil {
		hwlog.RunLog.Fatal("write encrypted passwd to file failed")
	}
	hwlog.RunLog.Info("create or update  passwd file successfully")
	if err = OverridePassWdFile(psBkFile, encryptedPd, mode); err != nil {
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
func PeriodCheck() {
	ticker := time.NewTicker(time.Duration(CheckInterval) * dayHours * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			now := time.Now()
			for _, v := range CertificateMap {
				checkCertStatus(now, v)
			}
		}
	}
}

func checkCertStatus(now time.Time, v *CertStatus) {
	if now.After(v.NotAfter) || now.Before(v.NotBefore) {
		hwlog.RunLog.Warnf("the certificate: %s is already overdue", v.FingerprintSHA256)
	}
	gapHours := v.NotAfter.Sub(now).Hours()
	overdueDays := gapHours / dayHours
	if overdueDays > math.MaxInt64 {
		overdueDays = math.MaxInt64
	}
	if overdueDays < float64(WarningDays) && overdueDays >= 0 {
		hwlog.RunLog.Warnf("the certificate: %s will overdue after %d days later",
			v.FingerprintSHA256, int64(overdueDays))
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
		return errors.New("get random words failed")
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
	return CheckCaCertV2(caFile, InvalidNum)
}

// CheckCaCertV2 check the import ca cert version 2
func CheckCaCertV2(caFile string, overdueTime int) ([]byte, error) {
	caBytes, err := LoadFile(caFile)
	if err != nil {
		return nil, err
	}
	if caBytes == nil {
		return nil, nil
	}
	caCrt, err := LoadCertsFromPEM(caBytes)
	if err != nil {
		return nil, errors.New("convert ca certificate failed")
	}
	if !caCrt.IsCA {
		return nil, errors.New("this is not ca certificate")
	}
	if err = checkExtension(caCrt); err != nil {
		return nil, err
	}
	switch overdueTime {
	case InvalidNum:
		err = CheckValidityPeriod(caCrt)
	default:
		err = CheckValidityPeriodWithError(caCrt, overdueTime)
	}
	if err != nil {
		return nil, err
	}
	if err = caCrt.CheckSignature(caCrt.SignatureAlgorithm, caCrt.RawTBSCertificate, caCrt.Signature); err != nil {
		return nil, errors.New("check ca certificate signature failed")
	}
	if err = AddToCertStatusTrace(caCrt); err != nil {
		hwlog.RunLog.Fatal(err)
	}
	hwlog.RunLog.Infof("ca certificate signature check pass")
	return caBytes, nil
}

// LoadCertPair load and valid encrypted certificate and private key
func LoadCertPair(cert, key, psFile, psFileBk string, encryptAlgorithm int) (*tls.Certificate, error) {
	pathMap := map[string]string{
		CertStorePath:      cert,
		KeyStorePath:       key,
		PassFilePath:       psFile,
		PassFileBackUpPath: psFileBk,
	}
	certBytes, keyPem, err := LoadCertPairByte(pathMap, encryptAlgorithm, FileMode)
	if err != nil {
		return nil, err
	}

	return ValidateCertPair(certBytes, keyPem, true, InvalidNum)
}

// CheckCertFiles CheckCertFiles
func CheckCertFiles(pathMap map[string]string) error {
	cert, ok := pathMap[CertStorePath]
	if !ok {
		return fmt.Errorf("%s is empty", CertStorePath)
	}
	key, ok := pathMap[KeyStorePath]
	if !ok {
		return fmt.Errorf("%s is empty", KeyStorePath)
	}
	psFile, ok := pathMap[PassFilePath]
	if !ok {
		return fmt.Errorf("%s is empty", PassFilePath)
	}
	psFileBk, ok := pathMap[PassFileBackUpPath]
	if !ok {
		return fmt.Errorf("%s is empty", PassFileBackUpPath)
	}

	// if password file not exists, remove privateKey and regenerate
	if !IsExists(psFile) && !IsExists(psFileBk) {
		hwlog.RunLog.Error("psFile or psFileBk is empty")
		return os.ErrNotExist
	}
	if !IsExists(key) {
		hwlog.RunLog.Error("keyFile is empty")
		return os.ErrNotExist
	}
	if !IsExists(cert) {
		hwlog.RunLog.Error("certFile is empty")
		return os.ErrNotExist
	}
	for _, v := range pathMap {
		if _, err := CheckPath(v); err != nil {
			return err
		}
	}

	return nil
}

// LoadCertPairByte load and valid encrypted certificate and private key
func LoadCertPairByte(pathMap map[string]string, encryptAlgorithm int, mode os.FileMode) ([]byte, []byte, error) {
	if err := CheckCertFiles(pathMap); err != nil {
		return nil, nil, err
	}
	cert := pathMap[CertStorePath]
	key := pathMap[KeyStorePath]
	psFile := pathMap[PassFilePath]
	psFileBk := pathMap[PassFileBackUpPath]
	certBytes, err := ReadLimitBytes(cert, Size10M)
	if err != nil {
		return nil, nil, errors.New("there is no certFile provided")
	}
	encodedPd := ReadOrUpdatePd(psFile, psFileBk, mode)
	KmcInit(encryptAlgorithm, "", "")
	pd, err := Decrypt(0, encodedPd)
	if err != nil {
		return nil, nil, errors.New("decrypt passwd failed")
	}
	hwlog.RunLog.Info("decrypt passwd successfully")
	keyBlock, err := DecryptPrivateKeyWithPd(key, pd)
	if err != nil {
		return nil, nil, err
	}
	hwlog.RunLog.Info("decrypt success")
	if Bootstrap != nil {
		Bootstrap.Shutdown()
	}
	PaddingAndCleanSlice(pd)
	return certBytes, pem.EncodeToMemory(keyBlock), nil
}

// ValidateCertPair ValidateCertPair
func ValidateCertPair(certBytes, keyPem []byte, periodCheck bool, overdueTime int) (*tls.Certificate, error) {
	var err error
	var tlsCert *tls.Certificate
	// preload cert and key files
	switch overdueTime {
	case InvalidNum:
		tlsCert, err = ValidateX509Pair(certBytes, keyPem)
	default:
		tlsCert, err = ValidateX509PairV2(certBytes, keyPem, overdueTime)
	}
	if err != nil || tlsCert == nil {
		return nil, err
	}
	x509Cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return nil, errors.New("parse certificate failed")
	}
	if err = AddToCertStatusTrace(x509Cert); err != nil {
		return nil, err
	}
	if periodCheck {
		go PeriodCheck()
	}
	return tlsCert, nil
}

// NewTLSConfig  create the tls config struct
func NewTLSConfig(caBytes []byte, certificate tls.Certificate, cipherSuites uint16) (*tls.Config, error) {
	return NewTLSConfigV2(caBytes, certificate, []uint16{cipherSuites})
}

// NewTLSConfigV2  create the tls config struct version 2
func NewTLSConfigV2(caBytes []byte, certificate tls.Certificate, cipherSuites []uint16) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: cipherSuites,
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

func write(path string, overrideByte []byte, mode os.FileMode) error {
	if err := ioutil.WriteFile(path, overrideByte, mode); err != nil {
		return errors.New("write encrypted key to config failed")
	}
	return nil
}

// check the certificate extensions, the cert version must be x509v3 and if the cert is ca, need check keyUsage,
// the keyUsage must include keyCertSign.
// detail information refer to https://datatracker.ietf.org/doc/html/rfc5280#section-4.2.1.3
func checkExtension(cert *x509.Certificate) error {
	if cert.Version != x509v3 {
		return errors.New("the certificate must be x509v3")
	}
	if !cert.IsCA {
		return nil
	}
	// ca cert need check whether the keyUsage include CertSign
	if (cert.KeyUsage & x509.KeyUsageCertSign) != x509.KeyUsageCertSign {
		msg := "CA certificate keyUsage didn't include keyCertSign"
		return errors.New(msg)
	}
	return nil
}

var dirPrefix = "/etc/mindx-dl/npu-exporter/"

// GetTLSConfigForClient get the tls config for client
func GetTLSConfigForClient(componentType string, encryptAlgorithm int) (*tls.Config, error) {
	if componentType != "npu-exporter" {
		dirPrefix = strings.Replace(dirPrefix, "npu-exporter", componentType, -1)
	}
	keyStore := dirPrefix + KeyStore
	certStore := dirPrefix + CertStore
	caStore := dirPrefix + CaStore
	passFile := dirPrefix + PassFileBackUp
	passFileBackUp := dirPrefix + PassFileBackUp
	tlsCert, err := LoadCertPair(certStore, keyStore, passFile, passFileBackUp, encryptAlgorithm)
	if err != nil {
		return nil, err
	}
	caBytes, err := CheckCaCert(caStore)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(caBytes); !ok {
		return nil, errors.New("append the CA file failed")
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*tlsCert},
		RootCAs:      pool,
	}
	return tlsConfig, nil
}

// AddToCertStatusTrace  add certstatus to trace map
func AddToCertStatusTrace(cert *x509.Certificate) error {
	if cert == nil {
		return errors.New("cert is nil")
	}

	sh256 := sha256.New()
	_, err := sh256.Write(cert.Raw)
	if err != nil {
		return err
	}
	fpsha256 := hex.EncodeToString(sh256.Sum(nil))

	cs := &CertStatus{
		NotBefore:         cert.NotBefore,
		NotAfter:          cert.NotAfter,
		IsCA:              cert.IsCA,
		FingerprintSHA256: fpsha256,
	}
	CertificateMap[fpsha256] = cs
	return nil
}

// CheckPath  validate path
func CheckPath(path string) (string, error) {
	return hwlog.CheckPath(path)
}

// ClientIP try to get the clientIP
func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

// LoadFile load file content
func LoadFile(filePath string) ([]byte, error) {
	if filePath == "" {
		return nil, nil
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, errors.New("the filePath is invalid")
	}
	if !IsExists(absPath) {
		return nil, nil
	}
	contentBytes, err := ReadLimitBytes(absPath, Size10M)
	if err != nil {
		return nil, errors.New("read file failed")
	}

	return contentBytes, nil
}

// Interceptor Interceptor
func Interceptor(h http.Handler, crlCertList *pkix.CertificateList) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if crlCertList != nil && CheckRevokedCert(r, crlCertList) {
			return
		}
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		h.ServeHTTP(w, r)
	})
}

// ReplacePrefix replate string with prefix
func ReplacePrefix(source, prefix string) string {
	if prefix == "" {
		prefix = "****"
	}
	if len(source) <= maskLen {
		return prefix
	}
	end := string([]rune(source)[maskLen:len(source)])
	return prefix + end
}

// MaskPrefix mask string prefix with ***
func MaskPrefix(source string) string {
	return ReplacePrefix(source, "")
}
