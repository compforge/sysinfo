package sysinfo

import (
	"os"
	"reflect"
	"testing"
)

func TestParseLimits(t *testing.T) {
	data := `Limit                     Soft Limit           Hard Limit           Units
Max cpu time              unlimited            unlimited            seconds
Max address space         1048576              unlimited            bytes
Max nice priority         0                    0
Max realtime priority     0                    0
`
	got, err := parseLimits(data)
	if err != nil {
		t.Fatal(err)
	}
	want := []ResourceLimit{
		{Name: "Max cpu time", Soft: "unlimited", Hard: "unlimited", Unit: "seconds"},
		{Name: "Max address space", Soft: "1048576", Hard: "unlimited", Unit: "bytes"},
		{Name: "Max nice priority", Soft: "0", Hard: "0"},
		{Name: "Max realtime priority", Soft: "0", Hard: "0"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v", got)
	}
	for _, data := range []string{"", "Max address space", "Max cpu time 0 broken seconds"} {
		if _, err := parseLimits(data); err == nil {
			t.Fatalf("expected error for %q", data)
		}
	}
}

func TestCommonFeatures(t *testing.T) {
	got := commonFeatures(nil, []string{"sse", "avx", "sse"})
	got = commonFeatures(got, []string{"sse", "avx2"})
	if !reflect.DeepEqual(got, []string{"sse"}) {
		t.Fatal(got)
	}
	got = commonFeatures(got, []string{"other"})
	got = commonFeatures(got, []string{"sse"})
	if got == nil || len(got) != 0 {
		t.Fatalf("empty intersection must stay empty: %v", got)
	}
}

func TestActualPageSize(t *testing.T) {
	var si SysInfo
	si.getRuntimeInfo()
	if si.Runtime.PageSizeBytes != os.Getpagesize() || si.Runtime.PageSizeBytes <= 0 {
		t.Fatalf("wrong page size: %d", si.Runtime.PageSizeBytes)
	}
}
