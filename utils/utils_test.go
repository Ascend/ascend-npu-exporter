//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package utils offer the some utils for certificate handling
package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"

	"huawei.com/kmc/pkg/adaptor/inbound/api/kmc"
	"huawei.com/kmc/pkg/adaptor/inbound/api/kmc/vo"
	"huawei.com/kmc/pkg/application/gateway"
	"huawei.com/kmc/pkg/application/gateway/loglevel"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/rand"
)

const testMode = 0660

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&config, nil)
}

// TestLoadCertsFromPEM test LoadCertsFromPEM
func TestLoadCertsFromPEM(t *testing.T) {
	Convey("test for DecryptPrivateKey", t, func() {
		Convey("normal cert", func() {
			caByte, err := ReadLimitBytes("./testdata/cert/ca.crt", Size10M)
			So(err, ShouldEqual, nil)
			ca, err := LoadCertsFromPEM(caByte)
			So(err, ShouldEqual, nil)
			So(ca.IsCA, ShouldBeTrue)
		})
		Convey("abnormal cert", func() {
			caByte, err := ReadLimitBytes("./testdata/cert/ca_err.crt", Size10M)
			So(err, ShouldEqual, nil)
			ca, err := LoadCertsFromPEM(caByte)
			So(ca, ShouldEqual, nil)
			So(err, ShouldNotBeEmpty)
		})
	})
}

//  TestCheckSignatureAlgorithm test CheckSignatureAlgorithm
func TestCheckSignatureAlgorithm(t *testing.T) {

	Convey("test for CheckSignatureAlgorithm", t, func() {
		Convey("normal cert", func() {
			caByte, err := ReadLimitBytes("./testdata/cert/ca.crt", Size10M)
			So(err, ShouldEqual, nil)
			ca, err := LoadCertsFromPEM(caByte)
			err = CheckSignatureAlgorithm(ca)
			So(err, ShouldEqual, nil)
		})
	})
}

// TestCheckRevokedCert test RevokedCert
func TestCheckRevokedCert(t *testing.T) {
	Convey("test for CheckRevokedCert", t, func() {
		Convey("cert revoked", func() {
			certByte, err := ReadLimitBytes("./testdata/checkcrl_testdata/certificate.crt", Size10M)
			So(err, ShouldEqual, nil)
			cert, _ := LoadCertsFromPEM(certByte)
			cacert, err := ReadLimitBytes("./testdata/checkcrl_testdata/ca.crt", Size10M)
			So(err, ShouldEqual, nil)
			ca, _ := LoadCertsFromPEM(cacert)
			r := &http.Request{
				TLS: &tls.ConnectionState{
					VerifiedChains: [][]*x509.Certificate{{cert, ca}},
					PeerCertificates: []*x509.Certificate{
						{SerialNumber: big.NewInt(1)}, cert},
				},
			}
			crlByte, err := ReadLimitBytes("./testdata/checkcrl_testdata/certificate_revokelist.crl", Size10M)
			So(err, ShouldEqual, nil)
			crl, err := x509.ParseCRL(crlByte)
			if err == nil {
				So(err, ShouldEqual, nil)
			}
			res := CheckRevokedCert(r, crl)
			So(res, ShouldBeTrue)
		})
		Convey("cert not revoked", func() {
			r := &http.Request{
				TLS: &tls.ConnectionState{},
			}
			crlcerList := &pkix.CertificateList{
				TBSCertList: pkix.TBSCertificateList{
					RevokedCertificates: []pkix.RevokedCertificate{{
						SerialNumber:   big.NewInt(1),
						RevocationTime: time.Time{},
						Extensions:     nil,
					}},
				},
			}
			res := CheckRevokedCert(r, nil)
			So(res, ShouldBeFalse)
			res = CheckRevokedCert(r, crlcerList)
			So(res, ShouldBeFalse)
		})
	})
}

// TestCheckPath test for file path check
func TestCheckPath(t *testing.T) {
	Convey("test for file path check", t, func() {
		Convey("file standardize", func() {
			conf, err := CheckPath("./testdata/cert/ca.crt")
			So(err, ShouldEqual, nil)
			So(conf, ShouldNotEqual, "./testdata/cert/ca.crt")
		})
		Convey("symlinks check", func() {
			err := os.Symlink("./testdata/cert/ca.crt", "./testdata/cert/ca.crtlnk")
			if err != nil {
				t.Error(err)
			}
			_, err = CheckPath("./testdata/cert/ca.crtlnk")
			So(err, ShouldNotBeEmpty)
			err = os.Remove("./testdata/cert/ca.crtlnk")
			fmt.Printf("remove file faild:%s", err)
		})
	})
}

// TestCertStatus test for checkCertStatus
func TestCertStatus(t *testing.T) {
	t1, err := time.Parse(time.RFC3339, "2022-03-18T00:00:00Z")
	if err != nil {
		fmt.Printf("Parse time failed %#v\n", err)
	}
	t2, err := time.Parse(time.RFC3339, "2022-03-20T00:00:00Z")
	if err != nil {
		fmt.Printf("Parse time failed %#v\n", err)
	}
	cs := &CertStatus{
		NotBefore: t1,
		NotAfter:  t2,
		IsCA:      true,
	}
	Convey("overdue 1day", t, func() {
		x, err := time.Parse(time.RFC3339, "2022-03-19T00:00:00Z")
		if err != nil {
			fmt.Printf("Parse time failed %#v\n", err)
		}
		checkCertStatus(x, cs)
	})
}

func createGetPrivateKeyLengthTestData(curve elliptic.Curve) (*x509.Certificate, *tls.Certificate) {
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Printf("create ecdsa private key failed: %#v\n", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"This is Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 1),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		fmt.Printf("Failed to create certificate: %s\n", err)
	}
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		fmt.Printf("Parse certificate failed: %s\n", err)
	}
	c := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		fmt.Printf("x509.MarshalECPrivateKey failed: %s\n", err)
	}
	k := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	keypair, err := tls.X509KeyPair(c, k)
	if err != nil {
		fmt.Printf("tls.X509KeyPair failed: %s\n", err)
	}
	return cert, &keypair
}

// TestGetPrivateKeyLength
func TestGetPrivateKeyLength(t *testing.T) {
	Convey("get key length of Curve 384", t, func() {
		// P384 curve key length is 388
		const bitLengthP384 = 384
		cert, keypair := createGetPrivateKeyLengthTestData(elliptic.P384())
		keyLen, keyType, err := GetPrivateKeyLength(cert, keypair)
		if err != nil {
			fmt.Printf("GetPrivateKeyLength failed %#v\n", err)
		}
		fmt.Printf("private key length is %#v, key type is %#v\n", keyLen, keyType)
		So(keyLen, ShouldEqual, bitLengthP384)
	})
	Convey("get key length of Curve 256", t, func() {
		// P521 curve key length is 256. the byte lengh is in 256
		const bitLengthP256 = 256
		cert, keypair := createGetPrivateKeyLengthTestData(elliptic.P256())
		keyLen, keyType, err := GetPrivateKeyLength(cert, keypair)
		if err != nil {
			fmt.Printf("GetPrivateKeyLength failed %#v\n", err)
		}
		fmt.Printf("private key length is %#v, key type is %#v\n", keyLen, keyType)
		So(keyLen, ShouldEqual, bitLengthP256)
	})
}

// TestGetPrivateKeyLength
func TestCheckValidityPeriodWithError(t *testing.T) {
	Convey("normal", t, func() {
		now := time.Now()
		cert := &x509.Certificate{
			NotBefore: now,
			NotAfter:  now.AddDate(1, 0, 0),
		}
		err := CheckValidityPeriodWithError(cert, 1)
		So(err, ShouldEqual, nil)
	})
	Convey("need update", t, func() {
		now := time.Now()
		cert := &x509.Certificate{
			NotBefore: now,
			NotAfter:  now.AddDate(0, 0, 1),
		}
		err := CheckValidityPeriodWithError(cert, 1)
		So(err.Error(), ShouldContainSubstring, "need to update certification")
	})
	Convey("overdue", t, func() {
		now := time.Now()
		cert := &x509.Certificate{
			NotBefore: now,
			NotAfter:  now,
		}
		err := CheckValidityPeriodWithError(cert, 1)
		So(err.Error(), ShouldContainSubstring, "the certificate overdue")
	})
}

// TestKmcInit
func TestKmcInit(t *testing.T) {
	Convey("success", t, func() {
		mock := gomonkey.ApplyFunc(kmc.NewManualBootstrap, func(defaultAppId int, logLevel loglevel.CryptoLogLevel,
			logger *gateway.CryptoLogger, kmcInitConfig *vo.KmcInitConfigVO) *kmc.ManualBootstrap {
			return nil
		})
		defer mock.Reset()
		defer func() {
			if r := recover(); r != nil {
				So(fmt.Sprintf("%#v", r), ShouldContainSubstring, "invalid memory address")
			}
		}()
		KmcInit(0, "./primary.key", "standby.key")
	})
}
