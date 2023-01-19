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

// Package utils provides the util func
package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestIsDir(t *testing.T) {
	convey.Convey("test logger", t, func() {
		convey.Convey("test IsDir func", func() {
			res := IsDir("/tmp/")
			convey.So(res, convey.ShouldBeTrue)
			res = IsDir("/utils/")
			convey.So(res, convey.ShouldBeTrue)
			res = IsDir("")
			convey.So(res, convey.ShouldBeFalse)
		})
	})
}

func TestIsFile(t *testing.T) {
	convey.Convey("test IsFile func", t, func() {
		res := IsFile("/tmp/")
		convey.So(res, convey.ShouldBeFalse)
		res = IsFile("")
		convey.So(res, convey.ShouldBeFalse)
	})
}

func TestIsExist(t *testing.T) {
	convey.Convey("test IsExist func", t, func() {
		res := IsExist("/xxxx/")
		convey.So(res, convey.ShouldBeFalse)
	})
}

func TestIsLexist(t *testing.T) {
	convey.Convey("test IsLexist func", t, func() {
		res := IsLexist("/xxxx/")
		convey.So(res, convey.ShouldBeFalse)
	})
}

func TestCheckPath(t *testing.T) {
	convey.Convey("test CheckPath func", t, func() {
		convey.Convey("should return itself given empty string", func() {
			res, err := CheckPath("")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return error given not exist path", func() {
			res, err := CheckPath("xxxxxxx")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "file does not exist")
		})

		convey.Convey("should return resolve path given normal path", func() {
			res, err := CheckPath("../../go.mod")
			convey.So(res, convey.ShouldNotBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return err when get abs path failed", func() {
			absStub := gomonkey.ApplyFunc(filepath.Abs, func(path string) (string, error) {
				return "", errors.New("abs failed")
			})
			defer absStub.Reset()
			res, err := CheckPath("../../go.mod")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "get the absolute path failed")
		})

		convey.Convey("should return err when get eval symbol link failed", func() {
			symStub := gomonkey.ApplyFunc(filepath.EvalSymlinks, func(path string) (string, error) {
				return "", errors.New("symlinks path failed")
			})
			defer symStub.Reset()
			res, err := CheckPath("../../go.mod")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "get the symlinks path failed")
		})

		convey.Convey("should return err given symbol link", func() {
			symStub := gomonkey.ApplyFunc(filepath.EvalSymlinks, func(path string) (string, error) {
				return "xxx", nil
			})
			defer symStub.Reset()
			res, err := CheckPath("../../go.mod")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "can't support symlinks")
		})

	})
}

func TestMakeSureDir(t *testing.T) {
	convey.Convey("test MakeSureDir func", t, func() {
		convey.Convey("normal situation, no err returned", func() {
			err := MakeSureDir("./testdata/tmp/test")
			convey.So(err, convey.ShouldEqual, nil)
		})
		convey.Convey("abnormal situation,err returned", func() {
			mock := gomonkey.ApplyFunc(os.MkdirAll, func(name string, perm os.FileMode) error {
				return fmt.Errorf("error")
			})
			defer mock.Reset()
			err := MakeSureDir("./xxxx/xxx")
			convey.So(err.Error(), convey.ShouldEqual, "create directory failed")
		})
	})
}

func TestGetDriverLibPath(t *testing.T) {
	convey.Convey("test GetDriverLibPath func", t, func() {
		convey.Convey("should return itself given empty string", func() {
			err := os.Setenv(ldLibPath, "")
			convey.So(err, convey.ShouldBeNil)
			res, err := GetDriverLibPath("")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("should return path when getLibFromEnv succeed", func() {
			envStub := gomonkey.ApplyFunc(getLibFromEnv, func(libraryName string) (string, error) {
				return "/test", nil
			})
			defer envStub.Reset()
			res, err := GetDriverLibPath("")
			convey.So(res, convey.ShouldEqual, "/test")
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return path when getLibFromEnv failed but getLibFromLdCmd succeed", func() {
			envStub := gomonkey.ApplyFunc(getLibFromEnv, func(libraryName string) (string, error) {
				return "", errors.New("failed")
			})
			defer envStub.Reset()
			cmdStub := gomonkey.ApplyFunc(getLibFromLdCmd, func(libraryName string) (string, error) {
				return "/test", nil
			})
			defer cmdStub.Reset()
			res, err := GetDriverLibPath("")
			convey.So(res, convey.ShouldEqual, "/test")
			convey.So(err, convey.ShouldBeNil)
		})

	})
}
