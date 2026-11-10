package main

import (
	"reflect"
	"testing"
)

func TestDiff(t *testing.T) {
	add, remove := diff([]string{"1.1.1.1", "2.2.2.2"}, []string{"2.2.2.2", "3.3.3.3"})
	if !reflect.DeepEqual(add, []string{"3.3.3.3"}) || !reflect.DeepEqual(remove, []string{"1.1.1.1"}) {
		t.Fatalf("add=%v remove=%v", add, remove)
	}
	if a, r := diff([]string{"1.1.1.1"}, []string{"1.1.1.1"}); len(a)+len(r) != 0 {
		t.Fatalf("expected in sync, got add=%v remove=%v", a, r)
	}
}

func TestNormalize(t *testing.T) {
	got := normalize([]string{"2.2.2.2", " 1.1.1.1", "2.2.2.2", ""})
	if !reflect.DeepEqual(got, []string{"1.1.1.1", "2.2.2.2"}) {
		t.Fatal(got)
	}
}

func TestFQDN(t *testing.T) {
	if fqdn("Apps.Example.com") != "apps.example.com." || fqdn("apps.example.com.") != "apps.example.com." {
		t.Fatal("fqdn")
	}
}

func TestPickZone(t *testing.T) {
	zones := []zoneRef{{"Z1", "example.com."}, {"Z2", "apps.example.com."}, {"Z3", "badexample.com."}}
	for name, want := range map[string]string{"x.apps.example.com": "Z2", "www.example.com.": "Z1", "apps.example.com": "Z2"} {
		if got, err := pickZone(zones, name); err != nil || got != want {
			t.Errorf("%s: got %q, %v; want %q", name, got, err, want)
		}
	}
	if _, err := pickZone(zones, "other.org"); err == nil {
		t.Error("expected no-zone error")
	}
	if _, err := pickZone(append(zones, zoneRef{"Z4", "example.com."}), "www.example.com"); err == nil {
		t.Error("expected ambiguity error")
	}
}

func TestVersion(t *testing.T) {
	// Under `go test` the build info is present; any non-empty answer is acceptable.
	if version() == "" {
		t.Fatal("version() is empty")
	}
}

func TestVersionPrefersLinkTimeValue(t *testing.T) {
	old := buildVersion
	t.Cleanup(func() { buildVersion = old })
	buildVersion = "v9.9.9"
	if got := version(); got != "v9.9.9" {
		t.Fatalf("version() = %q, want v9.9.9", got)
	}
}
