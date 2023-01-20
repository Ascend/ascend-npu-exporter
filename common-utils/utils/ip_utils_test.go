/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package utils offer the some utils for certificate handling
package utils

import (
	"net/http"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const (
	localhost     = "127.0.0.1"
	localhostLoop = "0.0.0.0"
)

func TestClientIP(t *testing.T) {
	convey.Convey("test ClientIP func", t, func() {
		convey.Convey("get IP from X-Forwarded-For", func() {
			ip := ClientIP(mockRequest(map[string][]string{"X-Forwarded-For": {localhost, localhostLoop}}))
			convey.So(ip, convey.ShouldEqual, localhost)
		})
		convey.Convey("get IP from X-Real-Ip", func() {
			ip := ClientIP(mockRequest(map[string][]string{"X-Forwarded-For": {},
				"X-Real-Ip": {localhost}}))
			convey.So(ip, convey.ShouldEqual, localhost)
		})
		convey.Convey("get IP from RemoteAddr", func() {
			ip := ClientIP(mockRequest(map[string][]string{"X-Forwarded-For": {},
				"X-Real-Ip": {}}))
			convey.So(ip, convey.ShouldEqual, localhost)
		})
		convey.Convey("get IP from RemoteAddr failed", func() {
			ip := ClientIP(&http.Request{RemoteAddr: localhost})
			convey.So(ip, convey.ShouldEqual, "")
		})
		convey.Convey("get IP failed", func() {
			ip := ClientIP(&http.Request{})
			convey.So(ip, convey.ShouldEqual, "")
		})
	})
}

func mockRequest(header map[string][]string) *http.Request {
	return &http.Request{
		Method:        "GET",
		URL:           nil,
		Proto:         "HTTP",
		ProtoMajor:    0,
		ProtoMinor:    0,
		Header:        header,
		ContentLength: 0,
		Close:         false,
		Host:          "www.test.com",
		RemoteAddr:    "127.0.0.1:8080",
	}
}
