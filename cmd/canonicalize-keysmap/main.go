/* SPDX-License-Identifier: GPL-3.0-only */

package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// TODO investigate what the exact rules are for groupID, artifactID and version strings.
var keysmapLineFormat = regexp.MustCompile(`^([a-zA-Z0-9\.\-_]+):([a-zA-Z0-9\.\-_]+):([0-9][0-9a-zA-Z\.\-_]*)\s*=\s*(?:0x([0-9A-F]{40}))?$`)

func main() {
	// "<groupID>:<artifactID>" -> version -> key fingerprint
	keysmap, groups, identifiers := readKeysMap(bufio.NewReader(os.Stdin))

	for _, groupID := range groups {
		groupFingerprint := allArtifactsVersionsSame(keysmap, groupID)
		expectFingerprintSet(groupFingerprint)
		if groupFingerprint != nil {
			writeKeysMapLine(groupID, groupFingerprint)
			continue
		}
		for _, identifier := range identifiers {
			artifact := keysmap[identifier]
			if !strings.HasPrefix(identifier, groupID+":") {
				continue
			}
			ranges, order := artifactVersionRanges(artifact)
			for _, versionrange := range order {
				fingerprint := ranges[versionrange]
				key := identifier
				if versionrange != "" {
					key += ":" + versionrange
				}
				writeKeysMapLine(key, fingerprint[:])
			}
		}
	}
	// TODO it would be possible to do a second pass and combine artifactIDs to '*' in case all artifacts are using the same version range specifier.
}

func writeKeysMapLine(identifier string, fingerprint []byte) {
	if bytes.Equal(fingerprint, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) {
		fmt.Printf("%s =\n", identifier)
	} else {
		fmt.Printf("%s = 0x%040X\n", identifier, fingerprint)
	}
}

var fingerprintUnset = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

func expectFingerprintSet(fpr []byte) {
	if bytes.Equal(fpr, fingerprintUnset) {
		panic("BUG: fingerprint should not be 'fingerprintUnset' sentinel value!")
	}
}

func artifactVersionRanges(artifact map[string][20]byte) (map[string][20]byte, []string) {
	versions := make([]string, 0, len(artifact))
	for v := range artifact {
		versions = append(versions, v)
	}
	versions = orderVersions(versions)

	ranges := make(map[string][20]byte, 1)
	rangeorder := make([]string, 0, 1)
	rangeStart := 0
	for i := 1; i < len(versions); i++ {
		if artifact[versions[i]] == artifact[versions[rangeStart]] {
			continue
		}
		if rangeStart == i-1 {
			// exactly 1 version in range, use version as-is
			rangekey := versions[rangeStart]
			ranges[rangekey] = artifact[versions[rangeStart]]
			rangeorder = append(rangeorder, rangekey)
		} else {
			// more than 1 version in range
			rangekey := "[" + versions[rangeStart] + "," + versions[i-1] + "]"
			ranges[rangekey] = artifact[versions[rangeStart]]
			rangeorder = append(rangeorder, rangekey)
		}
		rangeStart = i
	}
	if rangeStart == 0 {
		rangekey := ""
		ranges[rangekey] = artifact[versions[rangeStart]]
		rangeorder = append(rangeorder, rangekey)
	} else if rangeStart < len(versions) {
		rangekey := "[" + versions[rangeStart] + ",)"
		ranges[rangekey] = artifact[versions[rangeStart]]
		rangeorder = append(rangeorder, rangekey)
	}
	return ranges, rangeorder
}

func orderVersions(versions []string) []string {
	components := make([]component, 0, len(versions))
	for _, v := range versions {
		components = append(components, componentize(v))
	}
	sort.Slice(components, versionsorter(components))
	sorted := make([]string, 0)
	for _, v := range components {
		sorted = append(sorted, v.version)
	}
	return sorted
}

// versionsorter produces a function that sorts according to Maven's rules on
// version ordering:
// (https://maven.apache.org/ref/3.6.2/maven-artifact/apidocs/org/apache/maven/artifact/versioning/ComparableVersion.html)
//
// 1. component:
//    all-alpha / all-numeric
// 2. separators:
//    '-' / '.' / alpha-numeric-transition
// 3. qualifiers: strings are checked for well-known qualifiers and the
//    qualifier ordering is used for version ordering. Well-known qualifiers
//    (case insensitive) are:
//    - "alpha" or "a"
//    - "beta" or "b"
//    - "milestone" or "m"
//    - "rc" or "cr"
//    - "snapshot"
//    - (the empty string) or "ga" or "final"
//    - "sp"
//    Unknown qualifiers are considered after known qualifiers, with lexical
//    order (always case insensitive),
// 4. (a dash usually precedes a qualifier, and) is always less important than
//    something preceded with a dot.
func versionsorter(components []component) func(i, j int) bool {
	return func(i, j int) bool {
		compA := components[i].components[:]
		compB := components[j].components[:]
		for len(compA) < len(compB) {
			compA = append(compA, "")
		}
		for len(compA) > len(compB) {
			compB = append(compB, "")
		}
		for k := 0; k < len(compA) || k < len(compB); k++ {
			valueA := valuate(compA[k])
			valueB := valuate(compB[k])
			if valueA < valueB {
				return true
			}
			if valueA > valueB {
				return false
			}
			if valueA == extraordinaryLabelValue {
				return strings.Compare(compA[k], compB[k]) <= 0
			}
			continue
		}
		return true
	}
}

const extraordinaryLabelValue = 2

// valuate determines a symbolic value for the version component for use in mixed numeric/alpha comparison.
//
// version ordering:
// (https://maven.apache.org/ref/3.6.2/maven-artifact/apidocs/org/apache/maven/artifact/versioning/ComparableVersion.html)
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
//    - (the empty string) or "ga" or "final" [or "release", addition Danny]
//    - "sp"
//    Unknown qualifiers are considered after known qualifiers, with lexical
//    order (always case insensitive),
// [..]
// FIXME is there any consequence for not paying special attention to distinctio between '.' and '-'?
func valuate(v string) int64 {
	if len(v) == 0 {
		return 0
	}
	if classify([]byte(v)) == numeric {
		num, err := strconv.ParseInt(v, 10, 64)
		expectSuccess(err, "BUG: numeric version component is not parseable: %v")
		if num == 0 {
			return 0
		}
		// offset numeric value, such that numerics "> 0" are always preferred
		// over alpha components.
		return num + extraordinaryLabelValue
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
		return extraordinaryLabelValue
	}
}

func componentize(version string) component {
	components := []string{}
	cmp := ""
	for _, c := range []byte(version) {
		// FIXME added '_' as separator, not sure if this is correct but found in artifact-version list. (or treat as alpha?)
		if c == '.' || c == '-' || c == '_' {
			// in case of explicit separators '.' and '-'
			expect(len(cmp) > 0, "BUG? expected separator to separate either an alpha or numeric component.")
			components = append(components, strings.ToLower(cmp))
			cmp = ""
			continue
		}
		if len(cmp) > 0 && classify([]byte(cmp)) != classify([]byte{c}) {
			// in case of implicit separation
			components = append(components, strings.ToLower(cmp))
			cmp = ""
		}
		cmp += string(c)
	}
	if len(cmp) > 0 {
		components = append(components, strings.ToLower(cmp))
	}
	return component{version: version, components: components}
}

type component struct {
	version    string
	components []string
}

type tokenclass uint

const (
	_ tokenclass = iota
	alpha
	numeric
)

func classify(component []byte) tokenclass {
	if len(component) == 0 {
		return numeric
	}
	b := component[len(component)-1]
	if b >= '0' && b <= '9' {
		return numeric
	}
	if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') {
		return alpha
	}
	panic(fmt.Sprintf("BUG: Unknown token type: %v", b))
}

func allArtifactsVersionsSame(keysmap map[string]map[string][20]byte, groupID string) []byte {
	expect(len(keysmap) > 0, "Empty keysmap.")
	var previous = [20]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	for key, version := range keysmap {
		if !strings.HasPrefix(key, groupID+":") {
			continue
		}
		for _, fpr := range version {
			if bytes.Equal(previous[:], fingerprintUnset) {
				copy(previous[:], fpr[:])
			}
			if !bytes.Equal(previous[:], fpr[:]) {
				return nil
			}
		}
	}
	return previous[:]
}

func readKeysMap(reader *bufio.Reader) (map[string]map[string][20]byte, []string, []string) {
	// groupID:artifactID -> version -> fingerprint
	keysmap := make(map[string]map[string][20]byte, 0)
	groupset := make(map[string]struct{}, 0)
	artifactset := make(map[string]struct{}, 0)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		expectSuccess(err, "Unexpected failure reading line: %v")
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		matches := keysmapLineFormat.FindStringSubmatch(line)
		if matches == nil {
			os.Stderr.WriteString("WARNING: Line does not match format: " + line + "\n")
			continue
		}
		groupset[matches[1]] = struct{}{}
		key := matches[1] + ":" + matches[2]
		artifactset[key] = struct{}{}
		artifact := keysmap[key]
		if artifact == nil {
			artifact = make(map[string][20]byte, 1)
			keysmap[key] = artifact
		}
		var v [20]byte
		n, err := hex.Decode(v[:], []byte(matches[4]))
		expectSuccess(err, "Failed to decode key fingerprint: %v")
		if n != 0 && n != 20 {
			os.Stderr.WriteString(fmt.Sprintf("Incorrect length for key fingerprint: %d\n", n))
			continue
		}
		artifact[matches[3]] = v
	}

	groups := make([]string, 0, len(groupset))
	for k := range groupset {
		groups = append(groups, k)
	}
	sort.Strings(groups)
	identifiers := make([]string, 0, len(artifactset))
	for key := range artifactset {
		identifiers = append(identifiers, key)
	}
	sort.Strings(identifiers)
	return keysmap, groups, identifiers
}

func expectSuccess(err error, msg string) {
	expect(err == nil, fmt.Sprintf(msg, err))
}

func expect(result bool, msg string) {
	if !result {
		panic(msg)
	}
}
