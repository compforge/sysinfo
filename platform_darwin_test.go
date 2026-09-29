package sysinfo

import (
	"encoding/binary"
	"os"
	"strings"
	"syscall"
	"testing"
)

func TestDarwinReport(t *testing.T) {
	var si SysInfo
	si.GetSysInfo()
	if si.OS.Name != "macOS" || si.OS.Version == "" || si.Kernel.Release == "" || si.Kernel.Architecture == "" {
		t.Fatalf("missing platform information: %+v %+v", si.OS, si.Kernel)
	}
	if si.CPU.Cores == 0 || si.CPU.Threads == 0 || si.Memory.Size == 0 || si.Runtime.PageSizeBytes != os.Getpagesize() {
		t.Fatalf("missing hardware information: %+v %+v", si.CPU, si.Memory)
	}
	var limits syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limits); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, limit := range si.Runtime.Limits {
		if limit.Name == "Max open files" {
			found = true
			if limit.Soft != darwinLimit(limits.Cur) || limit.Hard != darwinLimit(limits.Max) {
				t.Fatal(limit)
			}
		}
	}
	if !found {
		t.Fatal("missing open file limit")
	}
	for key := range si.Runtime.Errors {
		if key == "platform" || strings.HasPrefix(key, "/proc/") {
			t.Fatalf("Linux collector ran: %s", key)
		}
	}
	if len(si.Runtime.Libc) != 0 || len(si.Runtime.VM) != 0 || len(si.Runtime.Notes) == 0 {
		t.Fatal("missing platform boundary")
	}
	if si.Runtime.Translation == "rosetta2" && (si.Kernel.Architecture != "arm64" || si.Runtime.ProcessArch != "amd64") {
		t.Fatalf("incorrect translated architecture: %+v %+v", si.Kernel, si.Runtime)
	}
}

func TestDecodeSysctlUint64(t *testing.T) {
	for _, value := range []uint64{0, 64 << 30, 0xffffffffffffffff} {
		var data [8]byte
		binary.LittleEndian.PutUint64(data[:], value)
		raw := strings.TrimSuffix(string(data[:]), "\x00")
		got, err := decodeSysctlUint64(raw)
		if err != nil || got != value {
			t.Fatalf("%d: got %d, %v", value, got, err)
		}
	}
	if _, err := decodeSysctlUint64("bad"); err == nil {
		t.Fatal("accepted invalid width")
	}
	if darwinLimit(uint64(syscall.RLIM_INFINITY)) != "unlimited" || darwinLimit(0) != "0" {
		t.Fatal("invalid limit rendering")
	}
}
