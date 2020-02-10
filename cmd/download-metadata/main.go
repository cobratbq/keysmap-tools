/* SPDX-License-Identifier: GPL-3.0-only */

package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/cobratbq/goutils/std/errors"
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
		err := cmd("curl", "-f", "-z", destFile, "-o", destFile, url)
		errors.RequireSuccess(err, "Failed to download metadata for artifact "+groupID+":"+artifactID+": %+v")
	}
	if err != io.EOF {
		panic(err.Error())
	}
}

func generateMetadataURL(groupID, artifactID string) string {
	groupIDPath := path.Join(strings.Split(groupID, ".")...)
	return repositoryBaseURL + path.Join(groupIDPath, artifactID, "maven-metadata.xml")
}

func cmd(command ...string) error {
	cmd := exec.Command(command[0], command[1:]...)
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		io.Copy(os.Stderr, stderrPipe)
		wg.Done()
	}()
	err = cmd.Run()
	wg.Wait()
	os.Stderr.Write([]byte{'\n'})
	return err
}
