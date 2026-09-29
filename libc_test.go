package sysinfo

import (
	"debug/elf"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestLibcBanner(t *testing.T) {
	for _, tt := range []struct{ banner, family, version string }{
		{"GNU C Library (Debian GLIBC 2.28-10) stable release version 2.28.\x00GLIBC_2.17", "glibc", "2.28"},
		{"GLIBC_2.34\x00GLIBC_2.17", "", ""},
		{"GNU C Library (custom build)", "glibc", ""},
		{"musl libc (aarch64)\nVersion 1.2.5\nDynamic Program Loader", "musl", "1.2.5"},
	} {
		family, version := libcBanner([]byte(tt.banner))
		if family != tt.family || version != tt.version {
			t.Fatalf("%q: %s %s", tt.banner, family, version)
		}
	}
}

func TestUserlandArchitecture(t *testing.T) {
	// Tiny ELF headers let us test ARM64 detection even on an amd64/macOS host.
	for _, tt := range []struct {
		machine elf.Machine
		arch    string
	}{
		{elf.EM_AARCH64, "arm64"}, {elf.EM_X86_64, "amd64"},
	} {
		data := make([]byte, 64)
		copy(data, "\x7fELF\x02\x01\x01")
		binary.LittleEndian.PutUint16(data[16:], uint16(elf.ET_EXEC))
		binary.LittleEndian.PutUint16(data[18:], uint16(tt.machine))
		binary.LittleEndian.PutUint32(data[20:], 1)
		binary.LittleEndian.PutUint16(data[52:], 64)
		path := filepath.Join(t.TempDir(), "shell")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		if got := userlandArchitecture(path); got != tt.arch {
			t.Fatalf("got %q, want %q", got, tt.arch)
		}
		if got := inspectLibc(path); got.Version != "" || got.Error == "" {
			t.Fatalf("invented libc: %+v", got)
		}
	}
}

func TestInvalidLibc(t *testing.T) {
	dir := t.TempDir()
	if got := inspectLibc(dir); got.Error == "" {
		t.Fatal("directory accepted")
	}
	path := filepath.Join(dir, "not-elf")
	if err := os.WriteFile(path, []byte("GLIBC_2.28"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := inspectLibc(path); got.Error == "" || got.Version != "" {
		t.Fatalf("%+v", got)
	}
}
