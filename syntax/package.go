// Package syntax provides package manifest parsing for Typst packages.
package syntax

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// PackageManifest represents a parsed package manifest.
// The UnknownFields contains fields which were found but not expected.
type PackageManifest struct {
	// Package contains details about the package itself.
	Package PackageInfo `toml:"package"`
	// Template contains details about the template, if the package is one.
	Template *TemplateInfo `toml:"template,omitempty"`
	// Tool is the tools section for third-party configuration.
	Tool ToolInfo `toml:"tool"`
	// UnknownFields contains all parsed but unknown fields for validation.
	UnknownFields map[string]any `toml:"-"`
}

// NewPackageManifest creates a new package manifest with the given package info.
func NewPackageManifest(pkg PackageInfo) *PackageManifest {
	return &PackageManifest{
		Package:       pkg,
		Template:      nil,
		Tool:          ToolInfo{Sections: make(map[string]map[string]any)},
		UnknownFields: make(map[string]any),
	}
}

// Validate ensures that this manifest is indeed for the specified package.
func (m *PackageManifest) Validate(spec *PackageSpec) error {
	if m.Package.Name != spec.Name {
		return fmt.Errorf("package manifest contains mismatched name `%s`", m.Package.Name)
	}

	if m.Package.Version != spec.Version {
		return fmt.Errorf("package manifest contains mismatched version %s", m.Package.Version)
	}

	if m.Package.Compiler != nil {
		current := CompilerVersion()
		if !current.MatchesGE(m.Package.Compiler) {
			return fmt.Errorf("package requires Typst %s or newer (current version is %s)",
				m.Package.Compiler, current)
		}
	}

	return nil
}

// ToolInfo represents the [tool] key in the manifest.
// This field can be used to retrieve 3rd-party tool configuration.
type ToolInfo struct {
	// Sections contains any fields parsed in the tool section.
	Sections map[string]map[string]any `toml:",inline"`
}

// TemplateInfo represents the [template] key in the manifest.
// The UnknownFields contains fields which were found but not expected.
type TemplateInfo struct {
	// Path is the directory within the package that contains the files
	// that should be copied into the user's new project directory.
	Path string `toml:"path"`
	// Entrypoint is a path relative to the template's path that points
	// to the file serving as the compilation target.
	Entrypoint string `toml:"entrypoint"`
	// Thumbnail is a path relative to the package's root that points
	// to a PNG or lossless WebP thumbnail for the template.
	Thumbnail *string `toml:"thumbnail,omitempty"`
	// UnknownFields contains all parsed but unknown fields for validation.
	UnknownFields map[string]any `toml:"-"`
}

// NewTemplateInfo creates a new template info with only required fields.
func NewTemplateInfo(path, entrypoint string) *TemplateInfo {
	return &TemplateInfo{
		Path:          path,
		Entrypoint:    entrypoint,
		Thumbnail:     nil,
		UnknownFields: make(map[string]any),
	}
}

// PackageInfo represents the [package] key in the manifest.
// The UnknownFields contains fields which were found but not expected.
type PackageInfo struct {
	// Name is the name of the package within its namespace.
	Name string `toml:"name"`
	// Version is the package's version.
	Version PackageVersion `toml:"version"`
	// Entrypoint is the path of the entrypoint into the package.
	Entrypoint string `toml:"entrypoint"`
	// Authors is a list of the package's authors.
	Authors []string `toml:"authors,omitempty"`
	// License is the package's license.
	License *string `toml:"license,omitempty"`
	// Description is a short description of the package.
	Description *string `toml:"description,omitempty"`
	// Homepage is a link to the package's web presence.
	Homepage *string `toml:"homepage,omitempty"`
	// Repository is a link to the repository where this package is developed.
	Repository *string `toml:"repository,omitempty"`
	// Keywords is an array of search keywords for the package.
	Keywords []string `toml:"keywords,omitempty"`
	// Categories is an array with up to three of the predefined categories
	// to help users discover the package.
	Categories []string `toml:"categories,omitempty"`
	// Disciplines is an array of disciplines defining the target audience
	// for which the package is useful.
	Disciplines []string `toml:"disciplines,omitempty"`
	// Compiler is the minimum required compiler version for the package.
	Compiler *VersionBound `toml:"compiler,omitempty"`
	// Exclude is an array of globs specifying files that should not be
	// part of the published bundle.
	Exclude []string `toml:"exclude,omitempty"`
	// UnknownFields contains all parsed but unknown fields for validation.
	UnknownFields map[string]any `toml:"-"`
}

// NewPackageInfo creates a new package info with only required fields.
func NewPackageInfo(name string, version PackageVersion, entrypoint string) PackageInfo {
	return PackageInfo{
		Name:          name,
		Version:       version,
		Entrypoint:    entrypoint,
		Authors:       []string{},
		Categories:    []string{},
		Compiler:      nil,
		Description:   nil,
		Disciplines:   []string{},
		Exclude:       []string{},
		Homepage:      nil,
		Keywords:      []string{},
		License:       nil,
		Repository:    nil,
		UnknownFields: make(map[string]any),
	}
}

// PackageSpec identifies a package.
type PackageSpec struct {
	// Namespace is the namespace the package lives in.
	Namespace string
	// Name is the name of the package within its namespace.
	Name string
	// Version is the package's version.
	Version PackageVersion
}

// Versionless returns a VersionlessPackageSpec from this PackageSpec.
func (s *PackageSpec) Versionless() VersionlessPackageSpec {
	return VersionlessPackageSpec{
		Namespace: s.Namespace,
		Name:      s.Name,
	}
}

// ParsePackageSpec parses a package specification from a string.
// Format: @namespace/name:version
func ParsePackageSpec(s string) (*PackageSpec, error) {
	scanner := newScanner(s)

	namespace, err := parseNamespace(scanner)
	if err != nil {
		return nil, err
	}

	name, err := parseName(scanner)
	if err != nil {
		return nil, err
	}

	version, err := parseVersion(scanner)
	if err != nil {
		return nil, err
	}

	return &PackageSpec{
		Namespace: namespace,
		Name:      name,
		Version:   version,
	}, nil
}

// String returns the string representation of a PackageSpec.
func (s *PackageSpec) String() string {
	return fmt.Sprintf("@%s/%s:%s", s.Namespace, s.Name, s.Version.String())
}

// VersionlessPackageSpec identifies a package, but not a specific version of it.
type VersionlessPackageSpec struct {
	// Namespace is the namespace the package lives in.
	Namespace string
	// Name is the name of the package within its namespace.
	Name string
}

// At fills in the version to get a complete PackageSpec.
func (s *VersionlessPackageSpec) At(version PackageVersion) *PackageSpec {
	return &PackageSpec{
		Namespace: s.Namespace,
		Name:      s.Name,
		Version:   version,
	}
}

// ParseVersionlessPackageSpec parses a versionless package specification from a string.
// Format: @namespace/name
func ParseVersionlessPackageSpec(s string) (*VersionlessPackageSpec, error) {
	scanner := newScanner(s)

	namespace, err := parseNamespace(scanner)
	if err != nil {
		return nil, err
	}

	name, err := parseName(scanner)
	if err != nil {
		return nil, err
	}

	if !scanner.Done() {
		return nil, errors.New("unexpected version in versionless package specification")
	}

	return &VersionlessPackageSpec{
		Namespace: namespace,
		Name:      name,
	}, nil
}

// String returns the string representation of a VersionlessPackageSpec.
func (s *VersionlessPackageSpec) String() string {
	return fmt.Sprintf("@%s/%s", s.Namespace, s.Name)
}

// scanner is a simple string scanner for parsing.
type scanner struct {
	s   string
	pos int
}

func newScanner(s string) *scanner {
	return &scanner{s: s, pos: 0}
}

func (sc *scanner) Done() bool {
	return sc.pos >= len(sc.s)
}

func (sc *scanner) EatIf(b byte) bool {
	if sc.pos < len(sc.s) && sc.s[sc.pos] == b {
		sc.pos++
		return true
	}
	return false
}

func (sc *scanner) EatUntil(b byte) string {
	start := sc.pos
	for sc.pos < len(sc.s) && sc.s[sc.pos] != b {
		sc.pos++
	}
	return sc.s[start:sc.pos]
}

func (sc *scanner) After() string {
	return sc.s[sc.pos:]
}

func parseNamespace(sc *scanner) (string, error) {
	if !sc.EatIf('@') {
		return "", errors.New("package specification must start with '@'")
	}

	namespace := sc.EatUntil('/')
	if namespace == "" {
		return "", errors.New("package specification is missing namespace")
	}
	if !IsNamespaceIdent(namespace) {
		return "", fmt.Errorf("`%s` is not a valid package namespace", namespace)
	}

	return namespace, nil
}

func parseName(sc *scanner) (string, error) {
	sc.EatIf('/')

	name := sc.EatUntil(':')
	if name == "" {
		return "", errors.New("package specification is missing name")
	}
	if !IsNamespaceIdent(name) {
		return "", fmt.Errorf("`%s` is not a valid package name", name)
	}

	return name, nil
}

func parseVersion(sc *scanner) (PackageVersion, error) {
	sc.EatIf(':')

	versionStr := sc.After()
	if versionStr == "" {
		return PackageVersion{}, errors.New("package specification is missing version")
	}

	return ParsePackageVersion(versionStr)
}

// IsNamespaceIdent checks if a string is a valid namespace identifier.
// A namespace identifier consists of lowercase letters, digits, and hyphens,
// must start with a letter, and cannot end with a hyphen.
func IsNamespaceIdent(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Must start with a lowercase letter
	if s[0] < 'a' || s[0] > 'z' {
		return false
	}

	// Cannot end with a hyphen
	if s[len(s)-1] == '-' {
		return false
	}

	// Check all characters
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}

	return true
}

// PackageVersion represents a package's version using semantic versioning.
type PackageVersion struct {
	// Major is the package's major version.
	Major uint32
	// Minor is the package's minor version.
	Minor uint32
	// Patch is the package's patch version.
	Patch uint32
}

// CompilerVersion returns the current compiler version.
// This is a placeholder - in a real implementation, this would return
// the actual Typst compiler version.
func CompilerVersion() PackageVersion {
	// Placeholder: return a default version
	return PackageVersion{Major: 0, Minor: 12, Patch: 0}
}

// MatchesEQ performs an == match with the given version bound.
// Version elements missing in the bound are ignored.
func (v PackageVersion) MatchesEQ(bound *VersionBound) bool {
	if v.Major != bound.Major {
		return false
	}
	if bound.Minor != nil && v.Minor != *bound.Minor {
		return false
	}
	if bound.Patch != nil && v.Patch != *bound.Patch {
		return false
	}
	return true
}

// MatchesGT performs a > match with the given version bound.
// The match only succeeds if some version element in the bound
// is actually greater than that of the version.
func (v PackageVersion) MatchesGT(bound *VersionBound) bool {
	if v.Major != bound.Major {
		return v.Major > bound.Major
	}
	if bound.Minor == nil {
		return false
	}
	if v.Minor != *bound.Minor {
		return v.Minor > *bound.Minor
	}
	if bound.Patch == nil {
		return false
	}
	if v.Patch != *bound.Patch {
		return v.Patch > *bound.Patch
	}
	return false
}

// MatchesLT performs a < match with the given version bound.
// The match only succeeds if some version element in the bound
// is actually less than that of the version.
func (v PackageVersion) MatchesLT(bound *VersionBound) bool {
	if v.Major != bound.Major {
		return v.Major < bound.Major
	}
	if bound.Minor == nil {
		return false
	}
	if v.Minor != *bound.Minor {
		return v.Minor < *bound.Minor
	}
	if bound.Patch == nil {
		return false
	}
	if v.Patch != *bound.Patch {
		return v.Patch < *bound.Patch
	}
	return false
}

// MatchesGE performs a >= match with the given version.
// The match succeeds when either a == or > match does.
func (v PackageVersion) MatchesGE(bound *VersionBound) bool {
	return v.MatchesEQ(bound) || v.MatchesGT(bound)
}

// MatchesLE performs a <= match with the given version.
// The match succeeds when either a == or < match does.
func (v PackageVersion) MatchesLE(bound *VersionBound) bool {
	return v.MatchesEQ(bound) || v.MatchesLT(bound)
}

// ParsePackageVersion parses a PackageVersion from a string.
// Format: major.minor.patch
func ParsePackageVersion(s string) (PackageVersion, error) {
	parts := strings.Split(s, ".")
	if len(parts) < 3 {
		missing := "major"
		if len(parts) >= 1 && parts[0] != "" {
			missing = "minor"
		}
		if len(parts) >= 2 && parts[1] != "" {
			missing = "patch"
		}
		return PackageVersion{}, fmt.Errorf("version number is missing %s version", missing)
	}
	if len(parts) > 3 {
		return PackageVersion{}, fmt.Errorf("version number has unexpected fourth component: `%s`", parts[3])
	}

	major, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil || parts[0] == "" {
		return PackageVersion{}, fmt.Errorf("`%s` is not a valid major version", parts[0])
	}

	minor, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil || parts[1] == "" {
		return PackageVersion{}, fmt.Errorf("`%s` is not a valid minor version", parts[1])
	}

	patch, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil || parts[2] == "" {
		return PackageVersion{}, fmt.Errorf("`%s` is not a valid patch version", parts[2])
	}

	return PackageVersion{
		Major: uint32(major),
		Minor: uint32(minor),
		Patch: uint32(patch),
	}, nil
}

// String returns the string representation of a PackageVersion.
func (v PackageVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// VersionBound represents a version bound for compatibility specification.
type VersionBound struct {
	// Major is the bound's major version.
	Major uint32
	// Minor is the bound's minor version (optional).
	Minor *uint32
	// Patch is the bound's patch version (optional, can only be present if minor is too).
	Patch *uint32
}

// ParseVersionBound parses a VersionBound from a string.
// Format: major[.minor[.patch]]
func ParseVersionBound(s string) (*VersionBound, error) {
	parts := strings.Split(s, ".")

	if len(parts) == 0 || parts[0] == "" {
		return nil, errors.New("version bound is missing major version")
	}

	major, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("`%s` is not a valid major version bound", parts[0])
	}

	bound := &VersionBound{
		Major: uint32(major),
	}

	if len(parts) > 1 && parts[1] != "" {
		minor, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("`%s` is not a valid minor version bound", parts[1])
		}
		minorVal := uint32(minor)
		bound.Minor = &minorVal
	}

	if len(parts) > 2 && parts[2] != "" {
		patch, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("`%s` is not a valid patch version bound", parts[2])
		}
		patchVal := uint32(patch)
		bound.Patch = &patchVal
	}

	if len(parts) > 3 {
		return nil, fmt.Errorf("version bound has unexpected fourth component: `%s`", parts[3])
	}

	return bound, nil
}

// String returns the string representation of a VersionBound.
func (b *VersionBound) String() string {
	s := fmt.Sprintf("%d", b.Major)
	if b.Minor != nil {
		s += fmt.Sprintf(".%d", *b.Minor)
	}
	if b.Patch != nil {
		s += fmt.Sprintf(".%d", *b.Patch)
	}
	return s
}
