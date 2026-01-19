package syntax

import (
	"testing"
)

func TestVersionVersionMatch(t *testing.T) {
	v1_1_1 := PackageVersion{Major: 1, Minor: 1, Patch: 1}

	// Test MatchesEQ
	bound1, _ := ParseVersionBound("1")
	if !v1_1_1.MatchesEQ(bound1) {
		t.Error("v1.1.1 should match == 1")
	}

	bound1_1, _ := ParseVersionBound("1.1")
	if !v1_1_1.MatchesEQ(bound1_1) {
		t.Error("v1.1.1 should match == 1.1")
	}

	bound1_2, _ := ParseVersionBound("1.2")
	if v1_1_1.MatchesEQ(bound1_2) {
		t.Error("v1.1.1 should not match == 1.2")
	}

	// Test MatchesGT
	if v1_1_1.MatchesGT(bound1) {
		t.Error("v1.1.1 should not match > 1")
	}

	bound1_0, _ := ParseVersionBound("1.0")
	if !v1_1_1.MatchesGT(bound1_0) {
		t.Error("v1.1.1 should match > 1.0")
	}

	if v1_1_1.MatchesGT(bound1_1) {
		t.Error("v1.1.1 should not match > 1.1")
	}

	// Test MatchesLT
	if v1_1_1.MatchesLT(bound1) {
		t.Error("v1.1.1 should not match < 1")
	}

	if v1_1_1.MatchesLT(bound1_1) {
		t.Error("v1.1.1 should not match < 1.1")
	}

	if !v1_1_1.MatchesLT(bound1_2) {
		t.Error("v1.1.1 should match < 1.2")
	}
}

func TestParsePackageVersion(t *testing.T) {
	tests := []struct {
		input   string
		want    PackageVersion
		wantErr bool
	}{
		{"0.1.0", PackageVersion{0, 1, 0}, false},
		{"1.2.3", PackageVersion{1, 2, 3}, false},
		{"10.20.30", PackageVersion{10, 20, 30}, false},
		{"", PackageVersion{}, true},
		{"1", PackageVersion{}, true},
		{"1.2", PackageVersion{}, true},
		{"1.2.3.4", PackageVersion{}, true},
		{"a.b.c", PackageVersion{}, true},
	}

	for _, tt := range tests {
		got, err := ParsePackageVersion(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParsePackageVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("ParsePackageVersion(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParsePackageSpec(t *testing.T) {
	tests := []struct {
		input   string
		want    *PackageSpec
		wantErr bool
	}{
		{
			"@preview/example:0.1.0",
			&PackageSpec{Namespace: "preview", Name: "example", Version: PackageVersion{0, 1, 0}},
			false,
		},
		{
			"@namespace/my-package:1.2.3",
			&PackageSpec{Namespace: "namespace", Name: "my-package", Version: PackageVersion{1, 2, 3}},
			false,
		},
		{"invalid", nil, true},
		{"@/name:1.0.0", nil, true},
		{"@ns/:1.0.0", nil, true},
		{"@ns/name", nil, true},
	}

	for _, tt := range tests {
		got, err := ParsePackageSpec(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParsePackageSpec(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr {
			if got.Namespace != tt.want.Namespace || got.Name != tt.want.Name || got.Version != tt.want.Version {
				t.Errorf("ParsePackageSpec(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	}
}

func TestPackageSpecString(t *testing.T) {
	spec := &PackageSpec{
		Namespace: "preview",
		Name:      "example",
		Version:   PackageVersion{0, 1, 0},
	}
	want := "@preview/example:0.1.0"
	if got := spec.String(); got != want {
		t.Errorf("PackageSpec.String() = %q, want %q", got, want)
	}
}

func TestIsIdent(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"example", true},
		{"my-package", true},
		{"pkg123", true},
		{"a", true},
		{"", false},
		{"123pkg", false},
		{"-pkg", false},
		{"pkg-", false},
		{"Pkg", false},
		{"pkg_name", false},
	}

	for _, tt := range tests {
		if got := IsIdent(tt.input); got != tt.want {
			t.Errorf("IsIdent(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseVersionBound(t *testing.T) {
	tests := []struct {
		input   string
		want    *VersionBound
		wantErr bool
	}{
		{"1", &VersionBound{Major: 1, Minor: nil, Patch: nil}, false},
		{"1.2", &VersionBound{Major: 1, Minor: ptr(uint32(2)), Patch: nil}, false},
		{"1.2.3", &VersionBound{Major: 1, Minor: ptr(uint32(2)), Patch: ptr(uint32(3))}, false},
		{"", nil, true},
		{"a", nil, true},
	}

	for _, tt := range tests {
		got, err := ParseVersionBound(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseVersionBound(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr {
			if got.Major != tt.want.Major {
				t.Errorf("ParseVersionBound(%q).Major = %v, want %v", tt.input, got.Major, tt.want.Major)
			}
			if (got.Minor == nil) != (tt.want.Minor == nil) {
				t.Errorf("ParseVersionBound(%q).Minor nil mismatch", tt.input)
			}
			if got.Minor != nil && tt.want.Minor != nil && *got.Minor != *tt.want.Minor {
				t.Errorf("ParseVersionBound(%q).Minor = %v, want %v", tt.input, *got.Minor, *tt.want.Minor)
			}
		}
	}
}

func ptr(v uint32) *uint32 {
	return &v
}
