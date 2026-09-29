//go:build darwin
// +build darwin

package sysinfo

import "syscall"

// Kernel information.
type Kernel struct {
	Release      string `json:"release,omitempty"`
	Version      string `json:"version,omitempty"`
	Architecture string `json:"architecture,omitempty"`
}

func (si *SysInfo) getKernelInfo() {
	si.Kernel.Release = si.sysctlString("kern.osrelease")
	si.Kernel.Version = si.sysctlString("kern.version")
	si.Kernel.Architecture = si.sysctlString("hw.machine")
	// Under Rosetta hw.machine may describe the translated process. ARM64
	// support or the translation flag identifies the native host;
	// runtime.process_architecture still describes the running executable.
	arm64, _ := syscall.SysctlUint32("hw.optional.arm64")
	if arm64 == 1 || si.Runtime.Translation == "rosetta2" {
		si.Kernel.Architecture = "arm64"
	}
}
