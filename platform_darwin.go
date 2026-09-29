package sysinfo

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func (si *SysInfo) sysctlString(name string) string {
	value, err := syscall.Sysctl(name)
	if err != nil {
		si.Runtime.Errors[name] = err.Error()
	}
	return strings.TrimSpace(value)
}

func (si *SysInfo) sysctlUint32(name string) uint32 {
	value, err := syscall.SysctlUint32(name)
	if err != nil {
		si.Runtime.Errors[name] = err.Error()
	}
	return value
}

func (si *SysInfo) getPlatformInfo() {
	si.getKernelInfo()
	si.OS = OS{Name: "macOS", Vendor: "Apple", Version: si.sysctlString("kern.osproductversion"),
		Release: si.sysctlString("kern.osversion"), Architecture: si.Kernel.Architecture}
	var err error
	si.Node.Hostname, err = os.Hostname()
	if err != nil {
		si.Runtime.Errors["hostname"] = err.Error()
	}
	si.Node.Timezone, _ = time.Now().Zone()
	si.Product.Vendor = "Apple"
	si.Product.Name = si.sysctlString("hw.model")
	si.CPU.Model = si.sysctlString("machdep.cpu.brand_string")
	si.CPU.Cpus = uint(si.sysctlUint32("hw.packages"))
	si.CPU.Cores = uint(si.sysctlUint32("hw.physicalcpu"))
	si.CPU.Threads = uint(si.sysctlUint32("hw.logicalcpu"))
	// x86 feature strings are optional on Apple Silicon. Do not translate
	// unrelated ARM capability keys into x86 instruction-set claims.
	if value, err := syscall.Sysctl("machdep.cpu.vendor"); err == nil {
		si.CPU.Vendor = value
	}
	for _, name := range []string{"machdep.cpu.features", "machdep.cpu.leaf7_features", "machdep.cpu.extfeatures"} {
		if value, err := syscall.Sysctl(name); err == nil {
			si.CPU.Features = append(si.CPU.Features, strings.Fields(strings.ToLower(value))...)
		}
	}
	if value, err := sysctlUint64("hw.memsize"); err == nil {
		si.Memory.Size = uint(value / (1024 * 1024))
		si.Runtime.Memory["physical_bytes"] = strconv.FormatUint(value, 10)
	} else {
		si.Runtime.Errors["hw.memsize"] = err.Error()
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		si.Runtime.Errors["network_interfaces"] = err.Error()
		return
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		si.Network = append(si.Network, NetworkDevice{Name: iface.Name, MACAddress: iface.HardwareAddr.String()})
	}
}

func sysctlUint64(name string) (uint64, error) {
	value, err := syscall.Sysctl(name)
	if err != nil {
		return 0, err
	}
	return decodeSysctlUint64(value)
}

func decodeSysctlUint64(value string) (uint64, error) {
	// syscall.Sysctl removes one trailing NUL even for binary data. Restore
	// that byte for uint64 values. Both supported Darwin architectures are LE.
	if len(value) != 7 && len(value) != 8 {
		return 0, fmt.Errorf("expected 8-byte sysctl value, got %d bytes", len(value))
	}
	var data [8]byte
	copy(data[:], value)
	return binary.LittleEndian.Uint64(data[:]), nil
}

func (si *SysInfo) getPlatformRuntime() {
	if translated, err := syscall.SysctlUint32("sysctl.proc_translated"); err == nil && translated == 1 {
		si.Runtime.Translation = "rosetta2"
	}
	si.Runtime.Notes = []string{
		"macOS uses the system libSystem runtime; Linux libc, overcommit and cgroup fields do not apply.",
		"macOS firmware, storage inventory and ARM CPU feature flags are not collected.",
	}
	for _, limit := range []struct {
		name     string
		resource int
		unit     string
	}{
		{"Max cpu time", syscall.RLIMIT_CPU, "seconds"},
		{"Max file size", syscall.RLIMIT_FSIZE, "bytes"},
		{"Max data size", syscall.RLIMIT_DATA, "bytes"},
		{"Max stack size", syscall.RLIMIT_STACK, "bytes"},
		{"Max core file size", syscall.RLIMIT_CORE, "bytes"},
		{"Max address space", syscall.RLIMIT_AS, "bytes"},
		{"Max open files", syscall.RLIMIT_NOFILE, "files"},
	} {
		var value syscall.Rlimit
		if err := syscall.Getrlimit(limit.resource, &value); err != nil {
			si.Runtime.Errors[limit.name] = err.Error()
			continue
		}
		si.Runtime.Limits = append(si.Runtime.Limits, ResourceLimit{Name: limit.name,
			Soft: darwinLimit(value.Cur), Hard: darwinLimit(value.Max), Unit: limit.unit})
	}
}

func darwinLimit(value uint64) string {
	if value == uint64(syscall.RLIM_INFINITY) {
		return "unlimited"
	}
	return strconv.FormatUint(value, 10)
}
