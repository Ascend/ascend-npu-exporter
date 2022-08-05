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
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"reflect"
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
			mock := gomonkey.ApplyFunc(os.MkdirAll, func(name string, perm os.FileMode) error {
				return fmt.Errorf("error")
			})
			defer mock.Reset()
			err := MakeSureDir("./xxxx/xxx")
			So(err.Error(), ShouldEqual, "create config directory failed")
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

	Convey("read from main file", t, func() {
		data := ReadOrUpdatePd(mainks, backupks, testMode)
		So(string(data), ShouldEqual, "111111")
		back, err := ReadBytes(backupks)
		So(err, ShouldEqual, nil)
		So(reflect.DeepEqual(back, data), ShouldBeTrue)
	})
	Convey("read from back file", t, func() {
		err := os.Remove(mainks)
		if err != nil {
			fmt.Println("clean source failed")
		}
		data := ReadOrUpdatePd(mainks, backupks, testMode)
		So(string(data), ShouldEqual, "111111")
		back, err := ReadBytes(mainks)
		So(err, ShouldEqual, nil)
		So(reflect.DeepEqual(back, data), ShouldBeTrue)
		// recover status before testing
		err = os.Remove(backupks)
		if err != nil {
			fmt.Println("clean source failed")
		}
	})
}

// TestEncryptPrivateKeyAgain test EncryptPrivateKeyAgain
func TestEncryptPrivateKeyAgain(t *testing.T) {
	var mainks = "./testdata/mainPd"
	var backupks = "./testdata/backupPd"
	Convey("test for EncryptPrivateKey", t, func() {
		// mock kmcInit
		initStub := gomonkey.ApplyFunc(KmcInit, func(sdpAlgID int, primaryKey, standbyKey string) error {
			return nil
		})
		defer initStub.Reset()
		encryptStub := gomonkey.ApplyFunc(Encrypt, func(domainID int, data []byte) ([]byte, error) {
			return []byte("test"), nil
		})
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
		Convey("normal situation,no err returned", func() {
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
		var backupks = "./testdata/mainks"
		// mock kmcInit
		initStub := gomonkey.ApplyFunc(KmcInit, func(sdpAlgID int, primaryKey, standbyKey string) error {
			return nil
		})
		defer initStub.Reset()
		Convey("normal cert", func() {
			encryptStub := gomonkey.ApplyFunc(Decrypt, func(domainID int, data []byte) ([]byte, error) {
				return []byte("111111"), nil
			})
			defer encryptStub.Reset()
			isEncryptedStub := gomonkey.ApplyFunc(isEncryptedKey, func(keyFile string) (bool, error) {
				return true, nil
			})
			defer isEncryptedStub.Reset()
			c, err := LoadCertPair("./testdata/cert/client-v3.crt",
				"./testdata/cert/client.key", mainks, backupks, 0)
			So(err, ShouldEqual, nil)
			So(c, ShouldNotBeEmpty)
		})
		Convey("cert not match", func() {
			encryptStub := gomonkey.ApplyFunc(Decrypt, func(domainID int, data []byte) ([]byte, error) {
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
			encryptStub := gomonkey.ApplyFunc(Decrypt, func(domainID int, data []byte) ([]byte, error) {
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

// TestReplacePrefix test for ReplacePrefix
func TestReplacePrefix(t *testing.T) {
	Convey("relative path", t, func() {
		path := ReplacePrefix("./testdata/cert/ca.crt", "****")
		So(path, ShouldEqual, "****testdata/cert/ca.crt")
	})
	Convey("absolute path", t, func() {
		path := ReplacePrefix("/testdata/cert/ca.crt", "****")
		So(path, ShouldEqual, "****estdata/cert/ca.crt")
	})
	Convey("path length less than 2", t, func() {
		path := ReplacePrefix("/", "****")
		So(path, ShouldEqual, "****")
	})
	Convey("empty string", t, func() {
		path := ReplacePrefix("", "****")
		So(path, ShouldEqual, "****")
	})

}

// TestClientIP test for ClientIP
func TestClientIP(t *testing.T) {
	Convey("get ip from RemoteAddr", t, func() {
		req, err := http.NewRequest("GET", "http://127.0.0.1", nil)
		req.RemoteAddr = "127.0.0.1:80"
		ip := ClientIP(req)
		So(err, ShouldEqual, nil)
		So(ip, ShouldEqual, "127.0.0.1")
	})
	Convey("get ip from X-Real-Ip", t, func() {
		req, err := http.NewRequest("GET", "http://127.0.0.1", nil)
		req.Header.Set("X-Real-IP", "127.0.0.2")
		ip := ClientIP(req)
		So(err, ShouldEqual, nil)
		So(ip, ShouldEqual, "127.0.0.2")
	})
	Convey("get ip from X-Forwarded-For", t, func() {
		req, err := http.NewRequest("GET", "http://127.0.0.1", nil)
		req.Header.Set("X-Forwarded-For", "127.0.0.3")
		ip := ClientIP(req)
		So(err, ShouldEqual, nil)
		So(ip, ShouldEqual, "127.0.0.3")
	})
}

// TestGetTLSConfigForClient test for GetTLSConfigForClient
func TestGetTLSConfigForClient(t *testing.T) {
	Convey("get tlsconfig", t, func() {
		gomonkey.ApplyFunc(LoadCertPairByte, func(pathMap map[string]string, encryptAlgorithm int,
			mode os.FileMode) ([]byte, []byte, error) {
			return nil, nil, errors.New("error")
		})
		cfg, err := GetTLSConfigForClient("npu-exporter", 1)
		So(err, ShouldNotBeEmpty)
		So(cfg, ShouldNotBeEmpty)
		So(cfg, ShouldEqual, nil)
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

func TestGetRandomPass(t *testing.T) {
	Convey("normal situation", t, func() {
		r1 := gomonkey.ApplyFunc(rand.Read, func(b []byte) (int, error) {
			for i := range b {
				b[i] = byte(i)
			}
			return len(b), nil
		})
		defer r1.Reset()
		res, err := GetRandomPass()
		So(err, ShouldEqual, nil)
		So(len(res), ShouldNotEqual, 0)
	})
	Convey("simple passwd situation", t, func() {
		r2 := gomonkey.ApplyFunc(rand.Read, func(b []byte) (int, error) {
			for i := range b {
				b[i] = 1
			}
			return len(b), nil
		})
		defer r2.Reset()
		_, err := GetRandomPass()
		So(err.Error(), ShouldEqual, "the password is to simple,please retry")
	})
}
