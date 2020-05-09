package main

import (
	"testing"
)

func TestVersionSorting(t *testing.T) {
	testvalues := []struct {
		ver1   string
		ver2   string
		result bool
	}{
		{"1", "1", false},
		{"1", "2", true},
		{"1.5", "2", true},
		{"1", "2.5", true},
		{"1", "1.0", false},
		{"1", "1.0.0", false},
		{"1.0", "1.0", false},
		{"1.0", "1", false},
		{"1.0", "1.1", true},
		{"1.0.0", "1.1", true},
		{"1.1", "1.2.0", true},
		{"1.0-alpha-1", "1.0", true},
		{"1.0-alpha-1", "1.0-alpha-2", true},
		{"1.0-alpha-1", "1.0-beta-1", true},
		{"1.0alpha", "1.0.alpha", false},
		{"1.0a2", "1.0.alpha.2", false},
		{"1.milestone.2", "1m2", false},
		{"1.0", "1.0-1", true},
		{"1.0-1", "1.0-2", true},
		{"2.0", "2-0", false},
		{"2.0", "2.0-0", false},
		{"2.0", "2.0-1", true},
		{"2.0", "2-1", true},
		{"2.0.1-klm", "2.0.1-lmn", true},
		{"2.0.1-xyz", "2.0.1", false},
		{"2.0.1", "2.0.1-123", true},
		{"2.0.1-xyz", "2.0.1-123", true},
		{"3.0-beta", "3.0", true},
		{"1-rc", "1", true},
		{"1-ga", "1", false},
		{"1.any", "1.1", true},
		{"1.any", "1-1", false},
		{"1.0", "1-sp", true},
		{"1-1", "1-sp", false},
		{"1.0-alpha-1-SNAPSHOT", "1.0-SNAPSHOT", true},
		{"2.0-1-SNAPSHOT", "2.0.1-SNAPSHOT", true},
		{"1.1.2", "1.1.2-3", true},
		{"1.1.2-3", "1.1.2", false},
	}
	for _, v := range testvalues {
		input := []version{componentize(v.ver1), componentize(v.ver2)}
		sorter := versionsorter(input)
		if sorter(0, 1) != v.result {
			t.Errorf("Expected %s < %s == %v", v.ver1, v.ver2, v.result)
		}
	}
}
