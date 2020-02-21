/* SPDX-License-Identifier: GPL-3.0-only */

package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cobratbq/goutils/std/builtin"
	"github.com/cobratbq/goutils/std/net/http"
)

const repositoryBaseURL = "https://repo1.maven.org/maven2/"

var artifactPattern = regexp.MustCompile(`([a-zA-Z0-9\._]+):([a-zA-Z0-9\.\-_]+)`)

func main() {
	destination := flag.String("d", "artifact-metadata", "Destination directory for artifact metadata.")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)
	var line string
	var err error
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		matches := artifactPattern.FindStringSubmatch(line)
		if matches == nil {
			os.Stderr.WriteString("no match: " + line + "\n")
			continue
		}
		groupID := matches[1]
		artifactID := matches[2]
		url := generateMetadataURL(groupID, artifactID)
		destFile := filepath.Join(*destination, strings.Join([]string{groupID, ":", artifactID, ".xml"}, ""))
		if _, err := os.Stat(destFile); err == nil {
			os.Stderr.WriteString("Skipping " + destFile + "\n")
			continue
		}
		os.Stderr.WriteString("Downloading " + url + " ...\n")
		err := http.DownloadToFilePath(destFile, url)
		builtin.RequireSuccess(err, "Failed to download metadata for artifact "+groupID+":"+artifactID+": %+v")
	}
	if err != io.EOF {
		panic(err.Error())
	}
}

func generateMetadataURL(groupID, artifactID string) string {
	groupIDPath := path.Join(strings.Split(groupID, ".")...)
	return repositoryBaseURL + path.Join(groupIDPath, artifactID, "maven-metadata.xml")
}
