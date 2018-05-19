package xbpspkgdb

import (
	"reflect"
	"testing"
)

func Must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestPkgdbRepodataEqual(t *testing.T) {
	p1, err := DecodeRepoDataFile("test/x86_64-repodata")
	Must(t, err)
	p2, err := DecodeFile("test/index.plist")
	Must(t, err)

	if !reflect.DeepEqual(p1, p2) {
		t.Fatalf("Parsed output differs")
	}
}

func TestFilterInstalled(t *testing.T) {
	p, err := DecodeFile("test/index.plist")
	Must(t, err)

	k1 := p.FilterKeys(Not(IsAuto))
	k2 := p.FilterKeys(IsManual)

	if !reflect.DeepEqual(k1, k2) {
		t.Fatal("Filtered names differ")
	}
}
