package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// TODO investigate what the exact rules are for groupID, artifactID and version strings.
var keysmapLineFormat = regexp.MustCompile(`^([a-zA-Z0-9\.\-_]+):([a-zA-Z0-9\.\-_]+):([0-9][0-9a-zA-Z\.\-_]*)\s*=\s*(?:0x([0-9A-F]{40}))?$`)

func main() {
	// map: groupID -> artifactID -> version -> key fingerprint
	keysmap := readKeysMap(bufio.NewReader(os.Stdin))

	for groupID, group := range keysmap {
		groupFingerprint := allArtifactsVersionsSame(group)
		expectFingerprintSet(groupFingerprint)
		if groupFingerprint != nil {
			writeKeysMapLine(groupID, "*", "*", groupFingerprint)
			continue
		}
		for artifactID, artifact := range group {
			artifactFingerprint := allFingerprintsSame(artifact)
			expectFingerprintSet(artifactFingerprint)
			if artifactFingerprint != nil {
				writeKeysMapLine(groupID, artifactID, "*", artifactFingerprint)
				continue
			}
			for version, fingerprint := range artifact {
				writeKeysMapLine(groupID, artifactID, version, fingerprint[:])
			}
		}
	}
}

func writeKeysMapLine(groupID, artifactID, version string, fingerprint []byte) {
	if bytes.Equal(fingerprint, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) {
		fmt.Printf("%s:%s:%s =\n", groupID, artifactID, version)
	} else {
		fmt.Printf("%s:%s:%s = 0x%040X\n", groupID, artifactID, version, fingerprint)
	}
}

var fingerprintUnset = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

func expectFingerprintSet(fpr []byte) {
	if bytes.Equal(fpr, fingerprintUnset) {
		panic("BUG: fingerprint should not be 'fingerprintUnset' sentinel value!")
	}
}

// func artifactVersionRangesSame(artifact map[string][20]byte) map[string][20]byte {

// }

// func orderVersions(versions []string) []string {
// 	mapping := make(map[string][]string, 0)
// 	for _, v := range versions {
// 		versionTokens = append(versionTokens, tokenize(v))
// 	}
// 	// FIXME now start sorting based on information from the mapping
// 	panic("TODO")
// }

// 1. component:
//    all-alpha / all-numeric
// 2. separators:
//    '-' / '.' / alpha-numeric-transition
// 3. qualifiers: strings are checked for well-known qualifiers and the
//    qualifier ordering is used for version ordering. Well-known qualifiers
//    (case insensitive) are:
//    - alpha or a
//    - beta or b
//    - milestone or m
//    - rc or cr
//    - snapshot
//    - (the empty string) or ga or final
//    - sp
//    Unknown qualifiers are considered after known qualifiers, with lexical
//    order (always case insensitive),
// 4. a dash usually precedes a qualifier, and is always less important than
//    something preceded with a dot.

// func tokenize(version string) []string {
// 	tokens := []string{}
// 	token := "" // alpha chars, numeric chars
// 	for _, c := range version {

// 	}
// 	return tokens
// }

func allArtifactsVersionsSame(group map[string]map[string][20]byte) []byte {
	expect(len(group) > 0, "Invalid input: no versions in artifact map.")
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

func allFingerprintsSame(artifact map[string][20]byte) []byte {
	expect(len(artifact) > 0, "Invalid input: no versions in artifact map.")
	var previous = [20]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	for _, fpr := range artifact {
		if bytes.Equal(previous[:], fingerprintUnset) {
			copy(previous[:], fpr[:])
		}
		if !bytes.Equal(previous[:], fpr[:]) {
			return nil
		}
	}
	return previous[:]
}

func readKeysMap(reader *bufio.Reader) map[string]map[string]map[string][20]byte {
	keysmap := make(map[string]map[string]map[string][20]byte, 0)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		expectSuccess(err, "Unexpected failure reading line: %v")
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			os.Stderr.WriteString("Skipping empty line ...\n")
			continue
		}
		matches := keysmapLineFormat.FindStringSubmatch(line)
		if matches == nil {
			os.Stderr.WriteString("WARNING: Line did not match: " + line + "\n")
			continue
		}
		groupID := matches[1]
		group := keysmap[groupID]
		if group == nil {
			group = make(map[string]map[string][20]byte, 1)
			keysmap[groupID] = group
		}
		artifactID := matches[2]
		version := matches[3]
		artifact := group[artifactID]
		if artifact == nil {
			artifact = make(map[string][20]byte, 1)
			group[artifactID] = artifact
		}
		var v [20]byte
		n, err := hex.Decode(v[:], []byte(matches[4]))
		expectSuccess(err, "Failed to decode key fingerprint: %v")
		if n != 0 && n != 20 {
			os.Stderr.WriteString(fmt.Sprintf("Incorrect length: %d\n", n))
			continue
		}
		artifact[version] = v
	}
	return keysmap
}

func expectSuccess(err error, msg string) {
	expect(err == nil, fmt.Sprintf(msg, err))
}

func expect(result bool, msg string) {
	if !result {
		panic(msg)
	}
}
