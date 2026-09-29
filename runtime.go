package sysinfo

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// RuntimeInfo records facts visible to this process, not a compatibility verdict.
// Limits are inherited from the caller; another service may have different limits.
type RuntimeInfo struct {
	PageSizeBytes int               `json:"page_size_bytes"`
	ProcessArch   string            `json:"process_architecture"`
	GoVersion     string            `json:"go_version"`
	Libc          []Libc            `json:"libc"`
	Limits        []ResourceLimit   `json:"resource_limits"`
	Memory        map[string]string `json:"memory"`
	VM            map[string]string `json:"vm"`
	Cgroup        string            `json:"cgroup_membership,omitempty"`
	Errors        map[string]string `json:"errors,omitempty"`
}

type ResourceLimit struct {
	Name string `json:"name"`
	Soft string `json:"soft"`
	Hard string `json:"hard"`
	Unit string `json:"unit,omitempty"`
}

func (si *SysInfo) getRuntimeInfo() {
	r := &si.Runtime
	*r = RuntimeInfo{
		// Getpagesize uses the runtime's actual OS page size, not Hugepagesize
		// or a compile-time architecture default. ARM64 can use 4K, 16K or 64K.
		PageSizeBytes: os.Getpagesize(),
		ProcessArch:   runtime.GOARCH,
		GoVersion:     runtime.Version(),
		Libc:          []Libc{}, Limits: []ResourceLimit{},
		Memory: map[string]string{}, VM: map[string]string{},
		Errors: map[string]string{},
	}
	if runtime.GOOS != "linux" {
		r.Errors["platform"] = "Linux is required"
		return
	}
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

func parseLimits(data string) ([]ResourceLimit, error) {
	limits := []ResourceLimit{}
	for _, line := range strings.Split(data, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] == "Limit" {
			continue
		}
		// The limit name contains spaces; the first numeric/unlimited column
		// marks its end. Some rows (nice/realtime priority) have no unit.
		i := 0
		for ; i < len(fields); i++ {
			if _, err := strconv.ParseUint(fields[i], 10, 64); err == nil || fields[i] == "unlimited" {
				break
			}
		}
		if i == 0 || i+2 > len(fields) || i+3 < len(fields) {
			return limits, fmt.Errorf("cannot parse resource limit: %q", line)
		}
		if _, err := strconv.ParseUint(fields[i+1], 10, 64); err != nil && fields[i+1] != "unlimited" {
			return limits, fmt.Errorf("invalid hard limit: %q", fields[i+1])
		}
		limit := ResourceLimit{Name: strings.Join(fields[:i], " "), Soft: fields[i], Hard: fields[i+1]}
		if i+2 < len(fields) {
			limit.Unit = fields[i+2]
		}
		limits = append(limits, limit)
	}
	if len(limits) == 0 {
		return limits, fmt.Errorf("no resource limits found")
	}
	return limits, nil
}

func commonFeatures(previous, current []string) []string {
	result := []string{}
	seen := map[string]bool{}
	for _, feature := range current {
		seen[feature] = true
	}
	if previous == nil {
		for feature := range seen {
			result = append(result, feature)
		}
	} else {
		for _, feature := range previous {
			if seen[feature] {
				result = append(result, feature)
			}
		}
	}
	sort.Strings(result)
	return result
}
