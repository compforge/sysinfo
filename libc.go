package sysinfo

import (
	"debug/elf"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Libc describes an installed library, not the libc linked into this static CLI.
// Multiple installations can coexist. Exported ABI versions are deliberately
// separate from the release version: GLIBC_2.x is not a libc release detector.
type Libc struct {
	Path        string   `json:"path"`
	Family      string   `json:"family,omitempty"`
	Version     string   `json:"version,omitempty"`
	Machine     string   `json:"elf_machine,omitempty"`
	ABIVersions []string `json:"abi_versions,omitempty"`
	Error       string   `json:"error,omitempty"`
}

var glibcRelease = regexp.MustCompile(`release version ([0-9]+\.[0-9]+(?:\.[0-9]+)?)`)
var muslRelease = regexp.MustCompile(`musl libc[^\x00]*\nVersion ([0-9]+\.[0-9]+\.[0-9]+)`)

func findLibc() ([]Libc, string) {
	var paths []string
	for _, dir := range []string{"/lib", "/lib64", "/usr/lib", "/usr/lib64"} {
		for _, suffix := range []string{"libc.so.6", "*/libc.so.6", "ld-musl-*.so.1"} {
			matches, _ := filepath.Glob(filepath.Join(dir, suffix))
			paths = append(paths, matches...)
		}
	}
	sort.Strings(paths)
	result := []Libc{}
	seen := map[string]bool{}
	for _, path := range paths {
		real, err := filepath.EvalSymlinks(path)
		if err != nil {
			result = append(result, Libc{Path: path, Error: err.Error()})
			continue
		}
		if seen[real] {
			continue
		}
		seen[real] = true
		result = append(result, inspectLibc(real))
	}
	if len(result) == 0 {
		return result, "no libc found in standard /lib{,64} or /usr/lib{,64} locations"
	}
	return result, ""
}

func inspectLibc(path string) Libc {
	result := Libc{Path: path}
	f, err := openRegularELF(path)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer f.Close()
	result.Machine = f.Machine.String()
	versions, err := f.DynamicVersions()
	if err != nil {
		result.Error = "read ABI versions: " + err.Error()
	}
	for _, version := range versions {
		if strings.HasPrefix(version.Name, "GLIBC_") {
			result.Family = "glibc"
			result.ABIVersions = append(result.ABIVersions, version.Name)
		}
	}
	sort.Strings(result.ABIVersions)
	// Never execute libc or a loader to discover its version. Read the embedded
	// release banner; if absent, preserve unknown rather than infer from ABI tags.
	if section := f.Section(".rodata"); section != nil {
		data, err := io.ReadAll(io.LimitReader(section.Open(), 8<<20))
		if err != nil {
			result.Error = "read release banner: " + err.Error()
		} else {
			family, version := libcBanner(data)
			if family != "" {
				result.Family, result.Version = family, version
			}
		}
	}
	if result.Version == "" && result.Error == "" {
		result.Error = "release version not found in ELF banner"
	}
	return result
}

func libcBanner(data []byte) (string, string) {
	text := string(data)
	if strings.Contains(text, "GNU C Library") {
		if m := glibcRelease.FindStringSubmatch(text); m != nil {
			return "glibc", m[1]
		}
		return "glibc", ""
	}
	if strings.Contains(text, "musl libc") {
		if m := muslRelease.FindStringSubmatch(text); m != nil {
			return "musl", m[1]
		}
		return "musl", ""
	}
	return "", ""
}

func openRegularELF(path string) (*elf.File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", path)
	}
	return elf.Open(path)
}

func userlandArchitecture(path string) string {
	f, err := openRegularELF(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	switch f.Machine {
	case elf.EM_X86_64:
		return "amd64"
	case elf.EM_386:
		return "386"
	case elf.EM_AARCH64:
		return "arm64"
	case elf.EM_ARM:
		return "arm"
	default:
		return f.Machine.String()
	}
}
