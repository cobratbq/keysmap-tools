/* SPDX-License-Identifier: GPL-3.0-only */

package main

import (
	"encoding/xml"
	"flag"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/cobratbq/goutils/std/builtin"
	io_ "github.com/cobratbq/goutils/std/io"
	http_ "github.com/cobratbq/goutils/std/net/http"
	os_ "github.com/cobratbq/goutils/std/os"
)

func main() {
	destination := flag.String("d", "artifact-signatures", "The destination location for downloaded artifact signatures.")
	flag.Parse()

	// TODO read metadata in separate goroutine and dump URLs on buffered channel.
	data := io_.MustReadAll(os.Stdin)
	var metadata metadata
	xml.Unmarshal(data, &metadata)
	for _, version := range metadata.Versions {
		destinationPath := path.Join(*destination, generateName(metadata.GroupID, metadata.ArtifactID, version))
		if _, err := os.Stat(destinationPath); err == nil {
			os.Stderr.WriteString("Skipping " + destinationPath + "\n")
			continue
		}
		url := generateURL(metadata.GroupID, metadata.ArtifactID, version)
		os.Stderr.WriteString("Downloading " + url + " ...\n")
		if err := http_.DownloadToFilePath(destinationPath, url); err != nil {
			if code, ok := err.(http_.ErrStatusCode); !ok || code != http.StatusNotFound {
				panic("Failed to download " + destinationPath + ": " + err.Error())
			}
			// no need to panic if document is simply not found (404)
			builtin.RequireSuccess(os_.CreateEmptyFile(destinationPath),
				"Failed to create empty file "+destinationPath+": %+v")
			os.Stderr.WriteString("  not found: " + url + "\n")
		}
	}
}

func generateName(groupID, artifactID, version string) string {
	return strings.Join([]string{groupID, ":", artifactID, ":", version, ".asc"}, "")
}

type metadata struct {
	GroupID     string   `xml:"groupId"`
	ArtifactID  string   `xml:"artifactId"`
	Latest      string   `xml:"versioning>latest"`
	Release     string   `xml:"versioning>release"`
	Versions    []string `xml:"versioning>versions>version"`
	LastUpdated string   `xml:"versioning>lastUpdated"`
}

const repositoryBaseURL = "https://repo1.maven.org/maven2/"

func generateURL(groupID, artifactID, version string) string {
	groupIDPath := path.Join(strings.Split(groupID, ".")...)
	fileName := artifactID + "-" + version + ".jar.asc"
	return repositoryBaseURL + path.Join(groupIDPath, artifactID, version, fileName)
}
