package sysinfo

import (
	"os"
	"strings"
)

func (si *SysInfo) getPlatformRuntime() {
	r := &si.Runtime
	read := func(path string) string {
		data, err := os.ReadFile(path)
		if err != nil {
			r.Errors[path] = err.Error()
			return ""
		}
		value := strings.TrimSpace(string(data))
		if value == "" {
			r.Errors[path] = "file is empty"
		}
		return value
	}
	if data := read("/proc/self/limits"); data != "" {
		var err error
		r.Limits, err = parseLimits(data)
		if err != nil {
			r.Errors["/proc/self/limits"] = err.Error()
		}
	}
	for _, line := range strings.Split(read("/proc/meminfo"), "\n") {
		if key, value, ok := strings.Cut(line, ":"); ok {
			r.Memory[key] = strings.TrimSpace(value)
		}
	}
	for _, name := range []string{"overcommit_memory", "overcommit_ratio", "overcommit_kbytes", "max_map_count", "mmap_min_addr"} {
		if value := read("/proc/sys/vm/" + name); value != "" {
			r.VM[name] = value
		}
	}
	r.Cgroup = read("/proc/self/cgroup")
	r.Libc, r.Errors["libc"] = findLibc()
	if r.Errors["libc"] == "" {
		delete(r.Errors, "libc")
	}
}
