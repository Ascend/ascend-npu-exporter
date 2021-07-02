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
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"strings"
)

var maxLen = 2048

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
func ReadPassWd() string {
	fmt.Print("Enter Private Key Password: ")
	bytePassword, err := terminal.ReadPassword(0)
	if err != nil {
		klog.Fatal("program error")
	}
	if len(bytePassword) > maxLen {
		klog.Fatal("input too long")
	}
	password := string(bytePassword)

	return strings.TrimSpace(password)
}

// ParsePrivateKeyWithPassword  decode the private key
func ParsePrivateKeyWithPassword(keyBytes []byte) ([]byte, error) {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("decode key file failed")
	}
	buf := block.Bytes
	if x509.IsEncryptedPEMBlock(block) {
		pd := ReadPassWd()
		var err error
		buf, err = x509.DecryptPEMBlock(block, []byte(pd))
		if err != nil {
			if err == x509.IncorrectPasswordError {
				return nil, err
			}
			return nil, errors.New("cannot decode encrypted private keys")
		}
	} else {
		klog.Warning("detect that you provided private key is not encrypted")
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:    block.Type,
		Headers: nil,
		Bytes:   buf,
	}), nil

}

// CheckCRL validate crl file
func CheckCRL(crlFile string) []byte {
	if crlFile == "" {
		return nil
	}
	crl, err := filepath.Abs(crlFile)
	if err != nil {
		klog.Fatalf("the crlFile is invalid")
	}
	if !IsExists(crl) {
		return nil
	}
	crlBytes, err := ioutil.ReadFile(crl)
	if err != nil {
		klog.Fatal("read crlFile failed")
	}
	_, err = x509.ParseCRL(crlBytes)
	if err != nil {
		klog.Fatal("parse crlFile failed")
	}
	return crlBytes
}

// MakeSureDir make sure the directory was existed
func MakeSureDir(path string) {
	dir := filepath.Dir(path)
	if !IsExists(dir) {
		err := os.Mkdir(dir, 0700)
		if err != nil {
			klog.Fatal("create config directory failed")
		}
	}
}
