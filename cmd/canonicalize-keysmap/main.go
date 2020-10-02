/* SPDX-License-Identifier: GPL-3.0-only */

package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/cobratbq/goutils/std/builtin"
	sort_ "github.com/cobratbq/goutils/std/sort"
)

// TODO investigate what the exact rules are for groupID, artifactID and version strings.
var keysmapLineFormat = regexp.MustCompile(`^([a-zA-Z0-9\.\-_]+):([a-zA-Z0-9\.\-_]+):([0-9a-zA-Z][0-9a-zA-Z\.\-\+_]*)\s*=\s*(0x[0-9A-F]{40}|noKey)?$`)

type fingerprint [20]byte

var fingerprintUnset = fingerprint{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
var fingerprintZero = fingerprint{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var fingerprintNoKey = fingerprint{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

func main() {
	// "<groupID>:<artifactID>" -> version -> key fingerprint
	keysmap, groups, identifiers := readKeysMap(bufio.NewReader(os.Stdin))

	for _, groupID := range groups {
		groupFingerprint := allArtifactsVersionsSame(keysmap, groupID)
		if groupFingerprint != fingerprintUnset {
			writeKeysMapLine(groupID, []fingerprint{groupFingerprint})
			continue
		}
		for _, identifier := range identifiers {
			artifact := keysmap[identifier]
			if !strings.HasPrefix(identifier, groupID+":") {
				continue
			}
			ranges, order := artifactVersionRanges(artifact)
			fingerprints := make([]fingerprint, 0)
			for _, versionrange := range order {
				fpr := ranges[versionrange]
				if fpr == fingerprintZero || fpr == fingerprintNoKey {
					key := identifier
					if versionrange != "" {
						key += ":" + versionrange
					}
					writeKeysMapLine(key, []fingerprint{fpr})
					continue
				}
				fingerprints = append(fingerprints, fpr)
			}
			writeKeysMapLine(identifier, fingerprints)
		}
	}
	// TODO it would be possible to do a second pass and combine artifactIDs to '*' in case all artifacts are using the same version range specifier.
}

func writeKeysMapLine(identifier string, fingerprints []fingerprint) {
	if len(fingerprints) <= 0 {
		return
	}
	if len(fingerprints) == 1 && fingerprints[0] == fingerprintZero {
		fmt.Printf("%s =\n", identifier)
	} else if len(fingerprints) == 1 && fingerprints[0] == fingerprintNoKey {
		fmt.Printf("%s = noKey\n", identifier)
	} else {
		fmt.Printf("%s = 0x%040X", identifier, fingerprints[0])
		for i := 1; i < len(fingerprints); i++ {
			fmt.Printf(", 0x%040X", fingerprints[i])
		}
		fmt.Printf("\n")
	}
}

func artifactVersionRanges(artifact map[string]fingerprint) (map[string]fingerprint, []string) {
	versions := artifactVersionOrder(artifact)

	ranges := make(map[string]fingerprint, 1)
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

func artifactVersionOrder(artifact map[string]fingerprint) []string {
	versions := make([]string, 0, len(artifact))
	for v := range artifact {
		versions = append(versions, v)
	}
	return orderVersions(versions)
}

func orderVersions(versionstrings []string) []string {
	versions := make([]version, 0, len(versionstrings))
	for _, v := range versionstrings {
		versions = append(versions, componentize(v))
	}
	sort.Slice(versions, versionsorter(versions))
	sorted := make([]string, 0)
	for _, v := range versions {
		sorted = append(sorted, v.source)
	}
	return sorted
}

func allArtifactsVersionsSame(keysmap map[string]map[string]fingerprint, groupID string) fingerprint {
	builtin.Require(len(keysmap) > 0, "Empty keysmap.")
	var previous = fingerprintUnset
	for key, version := range keysmap {
		if !strings.HasPrefix(key, groupID+":") {
			continue
		}
		for _, fpr := range version {
			if previous == fingerprintUnset {
				copy(previous[:], fpr[:])
			}
			if previous != fpr {
				return fingerprintUnset
			}
		}
	}
	return previous
}

func readKeysMap(reader *bufio.Reader) (map[string]map[string]fingerprint, []string, []string) {
	// groupID:artifactID -> version -> fingerprint
	keysmap := make(map[string]map[string]fingerprint, 0)
	groupset := make(map[string]struct{}, 0)
	artifactset := make(map[string]struct{}, 0)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		builtin.RequireSuccess(err, "Unexpected failure reading line: %v")
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
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
			artifact = make(map[string]fingerprint, 1)
			keysmap[key] = artifact
		}
		var v fingerprint
		var n int
		if matches[4] == "" {
			n, v = 0, fingerprintZero
		} else if matches[4] == "noKey" {
			n, v = len(fingerprintNoKey), fingerprintNoKey
		} else {
			n, err = hex.Decode(v[:], []byte(matches[4][2:]))
			builtin.RequireSuccess(err, "Failed to decode key fingerprint: %v")
		}
		if n != 0 && n != 20 {
			os.Stderr.WriteString(fmt.Sprintf("Incorrect length for public key fingerprint: %d\n", n))
			continue
		}
		artifact[matches[3]] = v
	}

	groups := sort_.StringSet(groupset)
	identifiers := sort_.StringSet(artifactset)
	return keysmap, groups, identifiers
}
