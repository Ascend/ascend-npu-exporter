//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package utils offer the some utils for certificate handling
package utils

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/prashantv/gostub"
	. "github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/hwlog"
	"math/big"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"
)

const testMode = 0660

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	stopCh := make(chan struct{})
	hwlog.InitRunLogger(&config, stopCh)
}

// TestCheckCRL test CheckCRL
func TestCheckCRL(t *testing.T) {
	Convey("CheckCRL test", t, func() {
		Convey("crl update time not match,return error", func() {
			_, err := CheckCRL("./testdata/cert/client.crl")
			So(err, ShouldNotBeEmpty)
		})
		Convey("directory no exist,no err returned", func() {
			_, err := CheckCRL("./testdata/cert/xxx.crl")
			So(err, ShouldEqual, nil)
		})
		Convey("crl file content wrong,err returned", func() {
			_, err := CheckCRL("./testdata/cert/client_err.crl")
			So(err, ShouldNotBeEmpty)
		})

	})
}

// TestMakeSureDir test MakeSureDir
func TestMakeSureDir(t *testing.T) {
	Convey("MakeSureDir test", t, func() {
		Convey("normal situation, no err returned", func() {
			err := MakeSureDir("./testdata/tmp/test")
			So(err, ShouldEqual, nil)
		})
		Convey("abnormal situation,err returned", func() {
			mock := gostub.Stub(&osMkdirAll, func(name string, perm os.FileMode) error {
				return fmt.Errorf("error")
			})
			defer mock.Reset()
			err := MakeSureDir("./xxxx/xxx")
			So(err, ShouldNotBeEmpty)
		})
	})
}

// TestOverridePassWdFile test OverridePassWdFile
func TestOverridePassWdFile(t *testing.T) {
	Convey("override padding test", t, func() {
		var path = "./testdata/test.key"
		data, err := ReadBytes("./testdata/cert/client.key")
		So(err, ShouldBeEmpty)
		err = OverridePassWdFile(path, data, testMode)
		So(err, ShouldBeEmpty)
		data2, err := ReadBytes(path)
		So(err, ShouldBeEmpty)
		So(reflect.DeepEqual(data, data2), ShouldBeTrue)
	})
}

// TestReadOrUpdatePd test ReadOrUpdatePd
func TestReadOrUpdatePd(t *testing.T) {
	var mainks = "./testdata/mainks"
	var backupks = "./testdata/backupks"
	Convey("Password back function test", t, func() {
		Convey("read from main file", func() {
			data := ReadOrUpdatePd(mainks, backupks, testMode)
			So(string(data), ShouldEqual, "111111")
			back, err := ReadBytes(backupks)
			So(err, ShouldEqual, nil)
			So(reflect.DeepEqual(back, data), ShouldBeTrue)
		})
		Convey("read from back file", func() {
			os.Remove(mainks)
			data := ReadOrUpdatePd(mainks, backupks, testMode)
			So(string(data), ShouldEqual, "111111")
			back, err := ReadBytes(mainks)
			So(err, ShouldEqual, nil)
			So(reflect.DeepEqual(back, data), ShouldBeTrue)
			// recover status before testing
			os.Remove(backupks)
		})
	})
}

// TestEncryptPrivateKeyAgain test EncryptPrivateKeyAgain
func TestEncryptPrivateKeyAgain(t *testing.T) {
	var mainks = "./testdata/mainPd"
	var backupks = "./testdata/backupPd"
	Convey("test for EncryptPrivateKey", t, func() {
		// mock kmcInit
		initStub := gostub.Stub(&KmcInit, func(sdpAlgID int, primaryKey, standbyKey string) {})
		encryptStub := gostub.Stub(&Encrypt, func(domainID int, data []byte) ([]byte, error) {
			return []byte("test"), nil
		})
		defer initStub.Reset()
		defer encryptStub.Reset()
		keyBytes, err := ReadBytes("./testdata/cert/client.key")
		So(err, ShouldEqual, nil)
		block, _ := pem.Decode(keyBytes)
		Convey("read from main file", func() {
			encryptedBlock, err := EncryptPrivateKeyAgain(block, mainks, backupks, 0)
			So(err, ShouldEqual, nil)
			_, ok := encryptedBlock.Headers["DEK-Info"]
			So(ok, ShouldBeTrue)
			pd, err := ReadBytes(mainks)
			So(err, ShouldEqual, nil)
			So(pd, ShouldNotBeEmpty)
		})

	})

}

// TestDecryptPrivateKeyWithPd test DecryptPrivateKeyWithPd
func TestDecryptPrivateKeyWithPd(t *testing.T) {
	Convey("test for DecryptPrivateKey", t, func() {
		Convey("private key is not encrypt", func() {
			block, err := DecryptPrivateKeyWithPd("./testdata/cert/client.key", nil)
			So(err, ShouldEqual, nil)
			_, ok := block.Headers["DEK-Info"]
			So(ok, ShouldBeFalse)
		})
		Convey("private key is  encrypted", func() {
			block, err := DecryptPrivateKeyWithPd("./testdata/cert/server-aes.key", []byte("111111"))
			So(err, ShouldEqual, nil)
			_, ok := block.Headers["DEK-Info"]
			So(ok, ShouldBeFalse)
		})
	})
}

// TestLoadCertsFromPEM test LoadCertsFromPEM
func TestLoadCertsFromPEM(t *testing.T) {
	Convey("test for DecryptPrivateKey", t, func() {
		Convey("normal cert", func() {
			caByte, err := ReadBytes("./testdata/cert/ca.crt")
			So(err, ShouldEqual, nil)
			ca, err := LoadCertsFromPEM(caByte)
			So(err, ShouldEqual, nil)
			So(ca.IsCA, ShouldBeTrue)
		})
		Convey("abnormal cert", func() {
			caByte, err := ReadBytes("./testdata/cert/ca_err.crt")
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
			caByte, err := ReadBytes("./testdata/cert/ca.crt")
			So(err, ShouldEqual, nil)
			ca, err := LoadCertsFromPEM(caByte)
			err = CheckSignatureAlgorithm(ca)
			So(err, ShouldEqual, nil)
		})
	})
}

// TestValidateX509Pair test ValidateX509Pair
func TestValidateX509Pair(t *testing.T) {
	Convey("test for ValidateX509Pair", t, func() {
		Convey("normal v1 cert", func() {
			certByte, err := ReadBytes("./testdata/cert/client-v1.crt")
			So(err, ShouldEqual, nil)
			keyByte, err := ReadBytes("./testdata/cert/client.key")
			So(err, ShouldEqual, nil)
			// validate period is 10 years, after that this case maybe failed
			c, err := ValidateX509Pair(certByte, keyByte)
			So(err, ShouldNotBeEmpty)
			So(c, ShouldEqual, nil)
		})
		Convey("normal cert", func() {
			certByte, err := ReadBytes("./testdata/cert/client-v3.crt")
			So(err, ShouldEqual, nil)
			keyByte, err := ReadBytes("./testdata/cert/client.key")
			So(err, ShouldEqual, nil)
			// validate period is 10 years, after that this case maybe failed
			c, err := ValidateX509Pair(certByte, keyByte)
			So(err, ShouldEqual, nil)
			So(c, ShouldNotBeEmpty)
		})
		Convey("not match cert", func() {
			certByte, err := ReadBytes("./testdata/cert/server.crt")
			So(err, ShouldEqual, nil)
			keyByte, err := ReadBytes("./testdata/cert/client.key")
			So(err, ShouldEqual, nil)
			c, err := ValidateX509Pair(certByte, keyByte)
			So(err, ShouldNotBeEmpty)
			So(c, ShouldEqual, nil)
		})
	})
}

// TestCheckRevokedCert test RevokedCert
func TestCheckRevokedCert(t *testing.T) {
	Convey("test for CheckRevokedCert", t, func() {
		Convey("cert revoked", func() {
			certByte, err := ReadBytes("./testdata/checkcrl_testdata/certificate.crt")
			So(err, ShouldEqual, nil)
			cert, _ := LoadCertsFromPEM(certByte)
			cacert, err := ReadBytes("./testdata/checkcrl_testdata/ca.crt")
			So(err, ShouldEqual, nil)
			ca, _ := LoadCertsFromPEM(cacert)
			r := &http.Request{
				TLS: &tls.ConnectionState{
					VerifiedChains: [][]*x509.Certificate{{cert, ca}},
					PeerCertificates: []*x509.Certificate{
						{SerialNumber: big.NewInt(1)}, cert},
				},
			}
			crlByte, err := ReadBytes("./testdata/checkcrl_testdata/certificate_revokelist.crl")
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

// TestCheckCaCert test for CheckCaCert
func TestCheckCaCert(t *testing.T) {
	Convey("test for CheckCaCert", t, func() {
		Convey("normal cert", func() {
			_, err := CheckCaCert("./testdata/cert/ca.crt")
			So(err, ShouldEqual, nil)
		})
		Convey("cert is nil", func() {
			_, err := CheckCaCert("")
			So(err, ShouldEqual, nil)
		})
		Convey("cert file is not exsit", func() {
			_, err := CheckCaCert("/djdsk.../dsd")
			So(err, ShouldEqual, nil)
		})
		Convey("ca file not right", func() {
			_, err := CheckCaCert("./testdata/cert/ca_err.crt")
			So(err, ShouldNotBeEmpty)
		})
		Convey("cert file is not ca", func() {
			_, err := CheckCaCert("./testdata/cert/server.crt")
			So(err, ShouldNotBeEmpty)
		})
	})
}

// TestLoadEncryptedCertPair  test load function
func TestLoadEncryptedCertPair(t *testing.T) {
	Convey("test for LoadCertPair", t, func() {
		var mainks = "./testdata/mainks"
		var backupks = "./testdata/backupks"
		// mock kmcInit
		initStub := gostub.Stub(&KmcInit, func(sdpAlgID int, primaryKey, standbyKey string) {})
		defer initStub.Reset()
		Convey("normal cert", func() {
			encryptStub := gostub.Stub(&Decrypt, func(domainID int, data []byte) ([]byte, error) {
				return []byte("111111"), nil
			})
			defer encryptStub.Reset()
			c, err := LoadCertPair("./testdata/cert/client-v3.crt",
				"./testdata/cert/client.key", mainks, backupks, 0)
			So(err, ShouldEqual, nil)
			So(c, ShouldNotBeEmpty)
		})
		Convey("cert not match", func() {
			encryptStub := gostub.Stub(&Decrypt, func(domainID int, data []byte) ([]byte, error) {
				return []byte("111111"), nil
			})
			defer encryptStub.Reset()
			c, err := LoadCertPair("./testdata/cert/server.crt",
				"./testdata/cert/client.key", mainks, backupks, 0)
			So(c, ShouldEqual, nil)
			So(err, ShouldNotBeEmpty)
		})
		Convey("cert not exist", func() {
			c, err := LoadCertPair("./testdata/xxx.crt",
				"./testdata/xxx/client.key", mainks, backupks, 0)
			So(c, ShouldEqual, nil)
			So(err, ShouldNotBeEmpty)
		})
		Convey("decrypt failed", func() {
			encryptStub := gostub.Stub(&Decrypt, func(domainID int, data []byte) ([]byte, error) {
				return nil, errors.New("mock err")
			})
			defer encryptStub.Reset()
			c, err := LoadCertPair("./testdata/cert/client-v1.crt",
				"./testdata/cert/client.key", mainks, backupks, 0)
			So(c, ShouldEqual, nil)
			So(err.Error(), ShouldEqual, "decrypt passwd failed")
		})

	})
}

// TestNewTLSConfig test for new tls
func TestNewTLSConfig(t *testing.T) {
	Convey("test for NewTLSConfig", t, func() {
		c := tls.Certificate{}
		Convey("One-way HTTPS", func() {
			conf, err := NewTLSConfig([]byte{}, c, tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256)
			So(err, ShouldEqual, nil)
			So(conf, ShouldNotBeEmpty)
		})
		Convey("Two-way HTTPS,but ca check failed", func() {
			conf, err := NewTLSConfig([]byte("sdsddd"), c, tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256)
			So(conf, ShouldEqual, nil)
			So(err, ShouldNotBeEmpty)
		})
		Convey("Two-way HTTPS", func() {
			ca, err := CheckCaCert("./testdata/cert/ca.crt")
			conf, err := NewTLSConfig(ca, c, tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256)
			So(err, ShouldEqual, nil)
			So(conf, ShouldNotBeEmpty)
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
