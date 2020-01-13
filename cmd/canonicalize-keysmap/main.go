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
	keysmap, groups := readKeysMap(bufio.NewReader(os.Stdin))

	for _, groupID := range groups {
		// groupFingerprint := allArtifactsVersionsSame(keysmap, groupID)
		// expectFingerprintSet(groupFingerprint)
		// if groupFingerprint != nil {
		// 	writeKeysMapLine(groupID+":*", "*", groupFingerprint)
		// 	continue
		// }
		for identifier, artifact := range keysmap {
			if !strings.HasPrefix(identifier, groupID+":") {
				continue
			}
			ranges, order := artifactVersionRanges(artifact)
			for _, versionrange := range order {
				fingerprint := ranges[versionrange]
				writeKeysMapLine(identifier, versionrange, fingerprint[:])
			}
		}
	}
	// TODO it would be possible to do a second pass and combine artifactIDs to '*' in case all artifacts are using the same version range specifier.
}

func writeKeysMapLine(groupIDartifactID, version string, fingerprint []byte) {
	if bytes.Equal(fingerprint, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) {
		fmt.Printf("%s:%s =\n", groupIDartifactID, version)
	} else {
		fmt.Printf("%s:%s = 0x%040X\n", groupIDartifactID, version, fingerprint)
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
		rangekey := "*"
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
			compA = append(compA, "0")
		}
		for len(compA) > len(compB) {
			compB = append(compB, "0")
		}
		for k := 0; k < len(compA) || k < len(compB); k++ {
			classA := classify(byte(compA[k][0]))
			classB := classify(byte(compB[k][0]))
			if classA < classB {
				return true
			} else if classA > classB {
				return false
			}
			if classA == numeric && classB == numeric {
				numA, err := strconv.ParseInt(compA[k], 10, 64)
				expectSuccess(err, "BUG: numeric version component is not parseable: %v")
				numB, err := strconv.ParseInt(compB[k], 10, 64)
				expectSuccess(err, "BUG: numeric version component is not parseable: %v")
				if numA < numB {
					return true
				} else if numA > numB {
					return false
				}
				continue
			}
			if classA == alpha && classB == alpha {
				// FIXME need to give priority to predefined qualifiers (alpha, beta, etc.)
				cmp := strings.Compare(compA[k], compB[k])
				if cmp < 0 {
					return true
				} else if cmp > 0 {
					return false
				}
				continue
			}
			panic("BUG: should not reach this! There seems to be a third class of version components?")
		}
		return true
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
		if len(cmp) > 0 && classify(cmp[len(cmp)-1]) != classify(c) {
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

func classify(b byte) tokenclass {
	if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') {
		return alpha
	}
	if b >= '0' && b <= '9' {
		return numeric
	}
	panic(fmt.Sprintf("BUG: Unknown token type: %v", b))
}

func allArtifactsVersionsSame(group map[string]map[string][20]byte) []byte {
	expect(len(group) > 0, "Invalid input: no artifacts in group map.")
	var previous = [20]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	for _, artifact := range group {
		for _, fpr := range artifact {
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

func readKeysMap(reader *bufio.Reader) (map[string]map[string][20]byte, []string) {
	// groupID:artifactID -> version -> fingerprint
	keysmap := make(map[string]map[string][20]byte, 0)
	groupmap := make(map[string]struct{}, 0)
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
		groupmap[matches[1]] = struct{}{}
		key := matches[1] + ":" + matches[2]
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

	groups := make([]string, 0, len(groupmap))
	for k := range groupmap {
		groups = append(groups, k)
	}
	sort.Strings(groups)
	return keysmap, groups
}

func expectSuccess(err error, msg string) {
	expect(err == nil, fmt.Sprintf(msg, err))
}

func expect(result bool, msg string) {
	if !result {
		panic(msg)
	}
}