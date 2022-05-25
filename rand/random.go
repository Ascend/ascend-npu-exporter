// Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package rand implement the security io.Reader
package rand

import (
	"io"
)

// Reader rand reader to generate security random bytes
var Reader io.Reader

// Read is a helper function that calls Reader.Read using io.ReadFull.
func Read(b []byte) (int, error) {
	return io.ReadFull(Reader, b)
}
