//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package utils offer the some utils for certificate handling
package utils

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// TestBuildConfigFromFlags test function for BuildConfigFromFlags
func TestBuildConfigFromFlags(t *testing.T) {
	Convey("relative path", t, func() {
		config, err := BuildConfigFromFlags("", "./testdata/test.conf")
		So(err, ShouldEqual, nil)
		So(config.Host, ShouldEqual, "https://127.0.0.1:6443")
	})
}
