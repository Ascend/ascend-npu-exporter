//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package utils offer the some utils for certificate handling
package utils

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"huawei.com/kmc/pkg/adaptor/inbound/api"
	"huawei.com/kmc/pkg/adaptor/inbound/api/kmc"
	"huawei.com/kmc/pkg/adaptor/inbound/api/kmc/vo"
	"huawei.com/kmc/pkg/application/gateway"
	"huawei.com/kmc/pkg/application/gateway/loglevel"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/npu-exporter/kmclog"
)

const (

	// FileMode file privilege
	FileMode = 0400

	// Aes256gcm AES256-GCM
	Aes256gcm = 9

	dayHours = 24

	yearHours = 87600

	tenDays = 10
	// Size10M  bytes of 10M
	Size10M            = 10 * 1024 * 1024
	maxSize            = 1024 * 1024 * 1024
	byteToBit          = 8
	defaultWarningDays = 100
	initSize           = 4
)

var (
	cryptoAPI api.CryptoApi
	// bootstrap kmc bootstrap
	bootstrap *kmc.ManualBootstrap
	// certificateMap  using certificate information
	certificateMap = make(map[string]*CertStatus, initSize)
	// warningDays cert warning day ,unit days
	warningDays = defaultWarningDays
	// checkInterval  cert period check interval,unit days
	checkInterval = 1
)

// CertStatus  the certificate valid period
type CertStatus struct {
	NotBefore         time.Time `json:"not_before"`
	NotAfter          time.Time `json:"not_after"`
	IsCA              bool      `json:"is_ca"`
	FingerprintSHA256 string    `json:"fingerprint_sha256,omitempty"`
}

// ReadLimitBytes read limit length of contents from file path
func ReadLimitBytes(path string, limitLength int) ([]byte, error) {
	key, err := CheckPath(path)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(key, os.O_RDONLY, FileMode)
	if err != nil {
		return nil, err
	}
	defer file.Close()
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

// CheckValidityPeriodWithError if the time expires, an error is reported
func CheckValidityPeriodWithError(cert *x509.Certificate, overdueTime int) error {
	overdueDays, err := GetValidityPeriod(cert)
	if err != nil {
		return err
	}
	if overdueDays <= float64(overdueTime) {
		return fmt.Errorf("overdueDayes is (%#v) need to update certification", overdueDays)
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
		return len(priv.X.Bytes()) * byteToBit, "ECC", nil
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

// KmcInit init kmc component
func KmcInit(sdpAlgID int, primaryKey, standbyKey string) error {
	if bootstrap == nil {
		defaultLogLevel := loglevel.Info
		var defaultLogger gateway.CryptoLogger = &kmclog.LoggerAdaptor{}
		defaultInitConfig := vo.NewKmcInitConfigVO()
		if primaryKey == "" {
			primaryKey = "/etc/mindx-dl/kmc_primary_store/master.ks"
		}
		if standbyKey == "" {
			standbyKey = "/etc/mindx-dl/.config/backup.ks"
		}
		if err := checkRootMaterial(primaryKey); err != nil {
			return err
		}
		if err := checkRootMaterial(standbyKey); err != nil {
			return err
		}
		defaultInitConfig.PrimaryKeyStoreFile = primaryKey
		defaultInitConfig.StandbyKeyStoreFile = standbyKey
		if sdpAlgID == 0 {
			sdpAlgID = Aes256gcm
		}
		defaultInitConfig.SdpAlgId = sdpAlgID
		bootstrap = kmc.NewManualBootstrap(0, defaultLogLevel, &defaultLogger, defaultInitConfig)
	}
	var err error
	cryptoAPI, err = bootstrap.Start()
	if err != nil {
		return errors.New("initial kmc failed,please make sure the LD_LIBRARY_PATH include the kmc-ext.so ")
	}
	if updateErr := cryptoAPI.UpdateLifetimeDays(tenDays * tenDays); updateErr != nil {
		hwlog.RunLog.Warn("update crypto lifetime failed ")
	}
	return nil
}

func checkRootMaterial(primaryKey string) error {
	if IsExists(primaryKey) {
		_, err := CheckPath(primaryKey)
		if err != nil {
			return errors.New("kmc root material file is symlinks")
		}
	}
	return nil
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
	if overdueDays < float64(warningDays) && overdueDays >= 0 {
		hwlog.RunLog.Warnf("the certificate: %s will overdue after %d days later",
			v.FingerprintSHA256, int64(overdueDays))
	}
}

var dirPrefix = "/etc/mindx-dl/npu-exporter/"

// CheckPath  validate path
func CheckPath(path string) (string, error) {
	if path == "" {
		return path, nil
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", errors.New("get the absolute path failed")
	}
	resoledPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return "", os.ErrNotExist
		}
		return "", errors.New("get the symlinks path failed")
	}
	if absPath != resoledPath {
		return "", errors.New("can't support symlinks")
	}
	return resoledPath, nil
}
