/* SPDX-License-Identifier: GPL-3.0-only */

package main

import (
	"encoding/xml"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
)

func main() {
	destination := flag.String("d", "artifact-signatures", "The destination location for downloaded artifact signatures.")
	flag.Parse()

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err.Error())
	}

	var metadata metadata
	xml.Unmarshal(data, &metadata)
	for _, version := range metadata.Versions {
		url := generateURL(metadata.GroupID, metadata.ArtifactID, version)
		name := generateName(metadata.GroupID, metadata.ArtifactID, version)
		destinationPath := path.Join(*destination, name)
		if err := cmd("curl", "-f", "-o", destinationPath, "-z", destinationPath, url); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok && exiterr.ProcessState.ExitCode() == 22 {
				// no need to panic if document is simply unavailable (404)
				f, err := os.Create(destinationPath)
				expectSuccess(err, "Failed to create empty file "+destinationPath)
				f.Close()
				continue
			}
			panic("Failed to download " + destinationPath + ": " + err.Error())
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

func expectSuccess(err error, msg string) {
	if err != nil {
		panic(msg + ": " + err.Error())
	}
}
