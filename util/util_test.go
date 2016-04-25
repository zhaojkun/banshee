// Copyright 2015 Eleme Inc. All rights reserved.

package util

import (
	"os"
	"testing"
)

func TestToFixed(t *testing.T) {
	Must(t, ToFixed(1.2345, 2) == "1.23")
	Must(t, ToFixed(10000.12121121, 5) == "10000.12121")
	Must(t, ToFixed(102, 3) == "102")
	Must(t, ToFixed(102.22, 3) == "102.22")
	Must(t, ToFixed(100, 3) == "100")
}

func TestIsFileExist(t *testing.T) {
	Must(t, IsFileExist(os.Args[0]))
	Must(t, !IsFileExist("file-not-exist"))
}
