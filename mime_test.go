package wail

import (
	"strings"
	"testing"
)

var emails = []string{
	"example1@example.com",
	"example2@example.com",
	"example3@example.com",
	"example4@example.com",
}

const subjectExample = `=?UTF-8?B?U29tZSB2ZXJ5IGxvbmcgdGV4dCB3aXRob3V0IG1lYW5pbmc=?= 
=?UTF-8?B?U29tZSB2ZXJ5IGxvbmcgdGV4dCB3aXRob3V0IG1lYW5pbmc=?= 
=?UTF-8?B?U29tZSB2ZXJ5IGxvbmcgdGV4dCB3aXRob3V0IG1lYW5pbmc=?=`

func TestMakeAddrString(t *testing.T) {
	if str := makeAddrString(emails[:1]); str != "<example1@example.com>" {
		t.Errorf("Invalid adress string, expect %s, got %s", "<example1@example.com>", str)
	}

	if str := makeAddrString(emails[:2]); str != "<example1@example.com>,<example2@example.com>" {
		t.Errorf("Invalid adress string, expect %s, got %s",
			"<example1@example.com>,<example2@example.com>", str)
	}

	if str := makeAddrString(emails); str != "<example1@example.com>,<example2@example.com>,<example3@example.com>,\r\n<example4@example.com>" {
		t.Errorf("Invalid adress string, expect %s, got %s",
			"<example1@example.com>,<example2@example.com>,<example3@example.com>,\r\n<example4@example.com>", str)
	}
}

func TestSplitHeader(t *testing.T) {
	str := ""

	if s := splitHeader(str); s != "" {
		t.Error("Trying to split an empty header")
	}

	if s := splitHeader("=?UTF-8?B?SGVsbG8gd29ybGQ=?="); s != "=?UTF-8?B?SGVsbG8gd29ybGQ=?=" {
		t.Errorf("Invalid split result, expect %s, got %s", "=?UTF-8?B?SGVsbG8gd29ybGQ=?=", s)
	}

	expect := "=?UTF-8?B?U29tZSB2ZXJ5IGxvbmcgdGV4dCB3aXRob3V0IG1lYW5pbmc=?=\r\n=?UTF-8?B?U29tZSB2ZXJ5IGxvbmcgdGV4dCB3aXRob3V0IG1lYW5pbmc=?=\r\n=?UTF-8?B?U29tZSB2ZXJ5IGxvbmcgdGV4dCB3aXRob3V0IG1lYW5pbmc=?="

	if s := splitHeader(subjectExample); s != expect {
		t.Errorf("Invalid split result, expect %s, got %s", expect, s)
	}

	expect = "=?UTF-8?B?VmVyeSB2ZXJ5IHZlcnkgdmVyeSB2ZXJ5IHZlcnkgdmVyeSB2ZXJ5IHZlcnkgdmVyeS\r\nB2ZXJ5IGxvbmcgc3RyaW5n?="

	if s := splitHeader("=?UTF-8?B?VmVyeSB2ZXJ5IHZlcnkgdmVyeSB2ZXJ5IHZlcnkgdmVyeSB2ZXJ5IHZlcnkgdmVyeSB2ZXJ5IGxvbmcgc3RyaW5n?="); s != expect {
		t.Errorf("Invalid split result, expect %s, got %s", expect, s)
	}
}

func TestSplit(t *testing.T) {
	s := "VmVyeSB2ZXJ5IHZlcnkgdmVyeSB2ZXJ5IHZlcnkgdmVyeSB2ZXJ5IHZlcnkgdmVyeSB2ZXJ5IGxvbmcgc3RyaW5n"
	str := split(s)

	expect := "VmVyeSB2ZXJ5IHZlcnkgdmVyeSB2ZXJ5IHZlcnkgdmVyeSB2ZXJ5IHZlcnkgdmVyeSB2ZXJ5IGxv\r\nbmcgc3RyaW5n"

	if s := strings.Join(str, "\r\n"); s != expect {
		t.Errorf("Invalid split result, expect %s, got %s", expect, s)
	}
}
