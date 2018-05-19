package xbpspkgdb

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"sort"

	"github.com/groob/plist"
)

var (
	_ = plist.NewDecoder
)

type Package struct {
	Alternatives map[string][]string `plist:"alternatives"`
	Architecture string              `plist:"architecture"`
	//Archive_compression_type string `plist:"archive-compression-type"` // unused?
	Automatic_install bool     `plist:"automatic-install"`
	Build_date        string   `plist:"build-date"`
	Build_options     string   `plist:"build-options"`
	Conf_files        []string `plist:"conf_files"`
	Conflicts         []string `plist:"conflicts"`
	Filename_sha256   string   `plist:"filename-sha256"`
	Filename_size     int      `plist:"filename-size"`
	Homepage          string   `plist:"homepage"`
	Install_date      string   `plist:"install-date"`
	Install_msg       []byte   `plist:"install-msg"`
	Install_script    []string `plist:"install-script"`
	Installed_size    int      `plist:"installed_size"`
	License           string   `plist:"license"`
	Maintainer        string   `plist:"maintainer"`
	Metafile_sha256   string   `plist:"metafile-sha256"`
	Pkgver            string   `plist:"pkgver"`
	Preserve          bool     `plist:"preserve"`
	Provides          []string `plist:"provides"`
	Remove_msg        []byte   `plist:"remove-msg"`
	Remove_script     []byte   `plist:"remove-script"`
	Replaces          []string `plist:"replaces"`
	Repolock          bool     `plist:"repolock"`
	Repository        string   `plist:"repository"`
	Reverts           []string `plist:"reverts"`
	Run_depends       []string `plist:"run_depends"`
	Shlib_provides    []string `plist:"shlib-provides"`
	Shlib_requires    []string `plist:"shlib-requires"`
	Short_desc        string   `plist:"short_desc"`
	Source_revisions  string   `plist:"source-revisions"`
	State             string   `plist:"state"`
}

type Pkgdb map[string]Package

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
func (p Pkgdb) Filter(f FilterFunc) Pkgdb {
	ret := make(Pkgdb)
	for k, v := range p {
		if f(v) {
			ret[k] = v
		}
	}
	return ret
}

type FilterFunc func(Package) bool

func Not(f FilterFunc) FilterFunc {
	return func(p Package) bool {
		return !f(p)
	}
}
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

func IsManual(p Package) bool {
	return !p.Automatic_install
}
func IsAuto(p Package) bool {
	return p.Automatic_install
}

func Decode(r io.Reader) (Pkgdb, error) {
	ret := make(Pkgdb)
	err := plist.NewDecoder(r).Decode(&ret)
	return ret, err

}
func DecodeFile(file string) (Pkgdb, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Decode(f)
}
func DecodeRepoData(r io.Reader) (Pkgdb, error) {
	var err error

	rgz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer rgz.Close()

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
func DecodeRepoDataFile(file string) (Pkgdb, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return DecodeRepoData(f)
}
