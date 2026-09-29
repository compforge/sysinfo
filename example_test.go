// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/compforge/sysinfo"
)

func TestGetSysInfo(t *testing.T) {
	var si sysinfo.SysInfo

	si.GetSysInfo()

	data, err := json.MarshalIndent(&si, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) || si.Meta.Version != sysinfo.Version {
		t.Fatal("invalid report or version")
	}
	if si.Runtime.PageSizeBytes != os.Getpagesize() {
		t.Fatalf("page size: %d", si.Runtime.PageSizeBytes)
	}
}
