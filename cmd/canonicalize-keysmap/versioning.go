package main

import (
	"fmt"
	"strings"

	"github.com/cobratbq/goutils/std/builtin"
	"github.com/cobratbq/goutils/std/strconv"
)

// extraordinaryLabelValue offsets any numeric version > 2 with +2, such that
// we can represent an order if alpha and numeric components are mixed.
// Offsetting +2 allows us to fit in alpha components "sp" and any undefined
// labels, while still respecting priority of numeric versions.
const extraordinaryLabelOffset = 2

// versionsorter produces a function that sorts according to Maven's rules on
// version ordering:
// (https://maven.apache.org/ref/3.6.2/maven-artifact/apidocs/org/apache/maven/artifact/versioning/ComparableVersion.html,
//  https://maven.apache.org/ref/3.6.3/maven-artifact/xref/org/apache/maven/artifact/versioning/ComparableVersion.html)
//
// 1. component:
//    all-alpha / all-numeric
// 2. separators:
//    '-' / '.' / alpha-numeric-transition
// [..]
func versionsorter(components []version) func(i, j int) bool {
	return func(i, j int) bool {
		versionA := components[i].components
		versionB := components[j].components
		for len(versionA) < len(versionB) {
			depth := versionA[len(versionA)-1].sub
			versionA = append(versionA, component{sub: depth, value: ""})
		}
		for len(versionA) > len(versionB) {
			depth := versionB[len(versionB)-1].sub
			versionB = append(versionB, component{sub: depth, value: ""})
		}
		for k := 0; k < len(versionA) || k < len(versionB); k++ {
			compA := versionA[k]
			valueCompA := valuate(compA.value)
			compB := versionB[k]
			valueCompB := valuate(compB.value)
			if compA.sub > compB.sub {
				return valueCompA < 0 || valueCompB > 0
			}
			if compA.sub < compB.sub {
				return valueCompA < 0 || (valueCompA == 0 && valueCompB > 0)
			}
			if valueCompA < valueCompB {
				return true
			}
			if valueCompA > valueCompB {
				return false
			}
			if valueCompA == extraordinaryLabelOffset {
				valueLabelCmp := strings.Compare(compA.value, compB.value)
				if valueLabelCmp != 0 {
					return valueLabelCmp < 0
				}
			}
		}
		return false
	}
}

// valuate determines a symbolic value for the version component for use in mixed numeric/alpha comparison.
//
// version ordering:
// (https://maven.apache.org/ref/3.6.2/maven-artifact/apidocs/org/apache/maven/artifact/versioning/ComparableVersion.html,
//  https://maven.apache.org/ref/3.6.3/maven-artifact/xref/org/apache/maven/artifact/versioning/ComparableVersion.html)
//
// [..]
// 3. qualifiers: strings are checked for well-known qualifiers and the
//    qualifier ordering is used for version ordering. Well-known qualifiers
//    (case insensitive) are:
//    - "alpha" or "a"
//    - "beta" or "b"
//    - "milestone" or "m"
//    - "rc" or "cr"
//    - "snapshot"
//    - (the empty string) or "ga" or "final" [or "release"]
//    - "sp"
//    Unknown qualifiers are considered after known qualifiers, with lexical
//    order (always case insensitive),
// 4. (a dash usually precedes a qualifier, and) is always less important than
//    something preceded with a dot.
func valuate(v string) int64 {
	if len(v) == 0 {
		return 0
	}
	if classify([]byte(v)) == numeric {
		num := strconv.MustParseInt(v, 10, 64)
		if num == 0 {
			return 0
		}
		// offset numeric value, such that numerics "> 0" are always preferred
		// over alpha components.
		return num + extraordinaryLabelOffset
	}
	switch strings.ToLower(v) {
	case "a", "alpha":
		return -5
	case "b", "beta":
		return -4
	case "m", "milestone":
		return -3
	case "rc", "cr":
		return -2
	case "snapshot":
		return -1
	case "", "ga", "final", "release":
		return 0
	case "sp":
		return 1
	default:
		return extraordinaryLabelOffset
	}
}

func componentize(versionstring string) version {
	components := []component{}
	depth := uint(0)
	cmp := ""
	for _, c := range []byte(versionstring) {
		if c == '.' || c == '-' {
			builtin.Require(len(cmp) > 0,
				"BUG? expected separator to separate either an alpha or numeric component.")
			components = append(components, component{sub: depth, value: strings.ToLower(cmp)})
			cmp = ""
			if c == '-' {
				depth++
			}
			continue
		}
		if len(cmp) > 0 && classify([]byte(cmp)) != classify([]byte{c}) {
			// in case of implicit separation
			components = append(components, component{sub: depth, value: strings.ToLower(cmp)})
			cmp = ""
		}
		cmp += string(c)
	}
	if len(cmp) > 0 {
		components = append(components, component{sub: depth, value: strings.ToLower(cmp)})
	}
	return version{source: versionstring, components: components}
}

type version struct {
	source     string
	components []component
}

type component struct {
	sub   uint
	value string
}

func classify(component []byte) tokenclass {
	b := component[0]
	if b >= '0' && b <= '9' {
		return numeric
	}
	if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '-' || b == '_' {
		return alpha
	}
	panic(fmt.Sprintf("BUG: Unknown token type: %c", b))
}

type tokenclass uint

const (
	_ tokenclass = iota
	alpha
	numeric
)
