//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package utils offer the some utils for certificate handling
package utils

import (
	"bytes"
	"github.com/agiledragon/gomonkey/v2"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestBuildConfigFromFlags test function for BuildConfigFromFlags
func TestBuildConfigFromFlags(t *testing.T) {
	kubeconfigBytes, err := ReadLimitBytes("./testdata/test.conf", Size10M)
	if err != nil {
		return
	}
	initStub := gomonkey.ApplyFunc(bytes.Contains, func(s, prefix []byte) bool {
		return false
	})
	defer initStub.Reset()
	kmc := gomonkey.ApplyFunc(KmcInit, func(sdpAlgID int, primaryKey, standbyKey string) error {
		return nil
	})
	defer kmc.Reset()
	decrypt := gomonkey.ApplyFunc(Decrypt, func(domainID int, data []byte) ([]byte, error) {
		return kubeconfigBytes, nil
	})
	defer decrypt.Reset()
	Convey("relative path", t, func() {
		config, err := BuildConfigFromFlags("", "./testdata/test.conf")
		So(err, ShouldEqual, nil)
		So(config.Host, ShouldEqual, "https://127.0.0.1:6443")
	})
	Convey("init client", t, func() {
		cli, err := K8sClient("./testdata/test.conf")
		So(err, ShouldEqual, nil)
		So(cli, ShouldNotEqual, nil)
	})
}

func TestK8sClientFor(t *testing.T) {

	Convey("get from init client", t, func() {
		cli, err := K8sClientFor("", "")
		So(err, ShouldEqual, nil)
		So(cli, ShouldNotEqual, nil)
	})
}
