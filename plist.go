package xbpspkgdb

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"sort"

	"github.com/groob/plist"
	"github.com/lemmi/closer"
)

// Package hold all metadata
type Package struct {
	Alternatives     map[string][]string `plist:"alternatives"`
	Architecture     string              `plist:"architecture"`
	AutomaticInstall bool                `plist:"automatic-install"`
	BuildDate        string              `plist:"build-date"`
	BuildOptions     string              `plist:"build-options"`
	ConfFiles        []string            `plist:"conf_files"`
	Conflicts        []string            `plist:"conflicts"`
	FilenameSha256   string              `plist:"filename-sha256"`
	FilenameSize     int                 `plist:"filename-size"`
	Homepage         string              `plist:"homepage"`
	InstallDate      string              `plist:"install-date"`
	InstallMsg       []byte              `plist:"install-msg"`
	InstallScript    []string            `plist:"install-script"`
	InstalledSize    int                 `plist:"installed_size"`
	License          string              `plist:"license"`
	Maintainer       string              `plist:"maintainer"`
	MetafileSha256   string              `plist:"metafile-sha256"`
	Pkgver           string              `plist:"pkgver"`
	Preserve         bool                `plist:"preserve"`
	Provides         []string            `plist:"provides"`
	RemoveMsg        []byte              `plist:"remove-msg"`
	RemoveScript     []byte              `plist:"remove-script"`
	Replaces         []string            `plist:"replaces"`
	Repolock         bool                `plist:"repolock"`
	Repository       string              `plist:"repository"`
	Reverts          []string            `plist:"reverts"`
	RunDepends       []string            `plist:"run_depends"`
	ShlibProvides    []string            `plist:"shlib-provides"`
	ShlibRequires    []string            `plist:"shlib-requires"`
	ShortDesc        string              `plist:"short_desc"`
	SourceRevisions  string              `plist:"source-revisions"`
	State            string              `plist:"state"`

	//Archive_compression_type string `plist:"archive-compression-type"` // unused?
}

// Pkgdb maps a package name to its metadata
type Pkgdb map[string]Package

// FilterKeys returns a slice of package names that match a `FilterFunc`
func (p Pkgdb) FilterKeys(f FilterFunc) []string {
	var ret []string
	for k, v := range p {
		if f(v) {
			ret = append(ret, k)
		}
	}
	sort.Strings(ret)
	return ret
}

// Filter returns a new `Pkgdb` with filtered packages
func (p Pkgdb) Filter(f FilterFunc) Pkgdb {
	ret := make(Pkgdb)
	for k, v := range p {
		if f(v) {
			ret[k] = v
		}
	}
	return ret
}

// FilterFunc is the type of the function called by the Pkgdb.Filter* methods
type FilterFunc func(Package) bool

// Not negates another FilterFunc
func Not(f FilterFunc) FilterFunc {
	return func(p Package) bool {
		return !f(p)
	}
}

// And returns true if all `FilterFunc` return true
func And(fs ...FilterFunc) FilterFunc {
	return func(p Package) bool {
		for _, f := range fs {
			if !f(p) {
				return false
			}
		}
		return true
	}
}

// Or returns true if any `FilterFunc` returns true
func Or(fs ...FilterFunc) FilterFunc {
	return func(p Package) bool {
		for _, f := range fs {
			if f(p) {
				return true
			}
		}
		return false
	}
}

// IsManual return true if the package is not automatically installed
func IsManual(p Package) bool {
	return !p.AutomaticInstall
}

// IsAuto return true if the package is automatically installed
func IsAuto(p Package) bool {
	return p.AutomaticInstall
}

// Decode parses a xbps plist stream into a `Pkgdb`
func Decode(r io.Reader) (Pkgdb, error) {
	ret := make(Pkgdb)
	err := plist.NewDecoder(r).Decode(&ret)
	return ret, err

}

// DecodeFile is a convenience function that parses xbps plist files
func DecodeFile(file string) (Pkgdb, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer closer.Do(f)

	return Decode(f)
}

// DecodeRepoData parses a stream of repodata
func DecodeRepoData(r io.Reader) (Pkgdb, error) {
	var err error

	rgz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer closer.Do(rgz)

	rtar := tar.NewReader(rgz)

	for {
		var h *tar.Header
		h, err = rtar.Next()
		if err == nil && h.Name == "index.plist" {
			return Decode(rtar)
		}
		if err != nil {
			return nil, err
		}
	}
	return nil, err
}

// DecodeRepoDataFile is a convenience function to parse xbps repodata files
func DecodeRepoDataFile(file string) (Pkgdb, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer closer.Do(f)

	return DecodeRepoData(f)
}
