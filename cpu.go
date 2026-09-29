// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// CPU information.
type CPU struct {
	Features []string `json:"features,omitempty"` // flags common to reported CPU feature sets
	Vendor   string   `json:"vendor,omitempty"`
	Model    string   `json:"model,omitempty"`
	Speed    uint     `json:"speed,omitempty"`   // CPU clock rate in MHz
	Cache    uint     `json:"cache,omitempty"`   // CPU cache size in KB
	Cpus     uint     `json:"cpus,omitempty"`    // number of physical CPUs
	Cores    uint     `json:"cores,omitempty"`   // number of physical CPU cores
	Threads  uint     `json:"threads,omitempty"` // number of logical (HT) CPU cores
}

var (
	reExtraSpace = regexp.MustCompile(" +")
	reCacheSize  = regexp.MustCompile(`^(\d+) KB$`)
)

func (si *SysInfo) getCPUInfo() {
	si.CPU.Threads = uint(runtime.NumCPU())

	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return
	}
	defer f.Close()

	cpu := make(map[string]bool)
	core := make(map[string]bool)

	var cpuID string

	s := bufio.NewScanner(f)
	for s.Scan() {
		if key, value, ok := strings.Cut(s.Text(), ":"); ok {
			sl := []string{strings.TrimSpace(key), strings.TrimSpace(value)}
			switch sl[0] {
			case "flags", "Features":
				si.CPU.Features = commonFeatures(si.CPU.Features, strings.Fields(sl[1]))
			case "CPU implementer":
				if si.CPU.Vendor == "" {
					si.CPU.Vendor = sl[1]
				}
			case "Hardware":
				if si.CPU.Model == "" {
					si.CPU.Model = sl[1]
				}
			case "physical id":
				cpuID = sl[1]
				cpu[cpuID] = true
			case "core id":
				coreID := fmt.Sprintf("%s/%s", cpuID, sl[1])
				core[coreID] = true
			case "vendor_id":
				if si.CPU.Vendor == "" {
					si.CPU.Vendor = sl[1]
				}
			case "model name":
				if si.CPU.Model == "" {
					// CPU model, as reported by /proc/cpuinfo, can be a bit ugly. Clean up...
					model := reExtraSpace.ReplaceAllLiteralString(sl[1], " ")
					si.CPU.Model = strings.Replace(model, "- ", "-", 1)
				}
			case "cache size":
				if si.CPU.Cache == 0 {
					if m := reCacheSize.FindStringSubmatch(sl[1]); m != nil {
						if cache, err := strconv.ParseUint(m[1], 10, 64); err == nil {
							si.CPU.Cache = uint(cache)
						}
					}
				}
			}
		}
	}
	if s.Err() != nil {
		return
	}

	si.CPU.Cpus = uint(len(cpu))
	si.CPU.Cores = uint(len(core))
}
