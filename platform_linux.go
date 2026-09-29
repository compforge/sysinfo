package sysinfo

func (si *SysInfo) getPlatformInfo() {
	// DMI info
	si.getProductInfo()
	si.getBoardInfo()
	si.getChassisInfo()
	si.getBIOSInfo()

	// SMBIOS info
	si.getMemoryInfo()

	// Node info
	si.getNodeInfo() // depends on BIOS info

	// Hardware info
	si.getCPUInfo() // depends on Node info
	si.getStorageInfo()
	si.getNetworkInfo()

	// Software info
	si.getOSInfo()
	si.getKernelInfo()
}
