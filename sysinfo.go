// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

// Package sysinfo is a Go library providing Linux and macOS / kernel / hardware system information.
package sysinfo

// SysInfo struct encapsulates all other information structs.
type SysInfo struct {
	Runtime RuntimeInfo     `json:"runtime"`
	Meta    Meta            `json:"sysinfo"`
	Node    Node            `json:"node"`
	OS      OS              `json:"os"`
	Kernel  Kernel          `json:"kernel"`
	Product Product         `json:"product"`
	Board   Board           `json:"board"`
	Chassis Chassis         `json:"chassis"`
	BIOS    BIOS            `json:"bios"`
	CPU     CPU             `json:"cpu"`
	Memory  Memory          `json:"memory"`
	Storage []StorageDevice `json:"storage,omitempty"`
	Network []NetworkDevice `json:"network,omitempty"`
}

// GetSysInfo gathers all available system information.
func (si *SysInfo) GetSysInfo() {
	*si = SysInfo{}
	si.getRuntimeInfo()

	// Meta info
	si.getMetaInfo()

	si.getPlatformInfo()
}
