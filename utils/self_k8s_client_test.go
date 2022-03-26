//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package utils offer the some utils for certificate handling
package utils

import (
	"bytes"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
)

// TestBuildConfigFromFlags test function for BuildConfigFromFlags
func TestBuildConfigFromFlags(t *testing.T) {
	Convey("relative path", t, func() {
		kubeconfigBytes, err := ReadLimitBytes("./testdata/test.conf", Size10M)
		initStub := gomonkey.ApplyFunc(bytes.HasPrefix, func(s, prefix []byte) bool {
			return false
		})
		defer initStub.Reset()
		kmc := gomonkey.ApplyFunc(KmcInit, func(sdpAlgID int, primaryKey, standbyKey string) {})
		defer kmc.Reset()
		decrypt := gomonkey.ApplyFunc(Decrypt, func(domainID int, data []byte) ([]byte, error) {
			return kubeconfigBytes, nil
		})
		defer decrypt.Reset()
		config, err := BuildConfigFromFlags("", "./testdata/test.conf")
		So(err, ShouldEqual, nil)
		So(config.Host, ShouldEqual, "https://127.0.0.1:6443")
	})
}
