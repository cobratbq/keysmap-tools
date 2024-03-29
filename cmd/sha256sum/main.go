/* SPDX-License-Identifier: GPL-3.0-only */

package main

import (
	"bufio"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/cobratbq/goutils/assert"
	io_ "github.com/cobratbq/goutils/std/io"
)

// TODO consider if we should change error handling tactics, as we silence the original error now - in favor of our own error.
func main() {
	config := initConfig()
	if err := verifyConfig(config); err != nil {
		os.Exit(1)
	}

	exitCode := 0
	if config.checkMode {
		// check existing checksum files
		failures := uint(0)
		for _, source := range config.sources {
			sourceFails, err := verifySource(config, source)
			if err != nil {
				if err == os.ErrNotExist {
					os.Stderr.WriteString("sha256sum: " + source + ": No such file or directory\n")
				} else if err == os.ErrInvalid {
					os.Stderr.WriteString("sha256sum: " + source + ": read error\n")
				} else if err == io.ErrNoProgress {
					os.Stderr.WriteString("sha256sum: " + source + ": no properly formatted SHA256 checksum lines found\n")
				} else {
					panic("Unexpected failure verifying source '" + source + "': " + err.Error())
				}
				exitCode = 1
				continue
			}
			failures += sourceFails
		}
		if failures > 0 {
			exitCode = 1
			os.Stderr.WriteString(fmt.Sprintf("sha256sum: WARNING: %d computed checksum(s) did NOT match\n", failures))
		}
	} else {
		// generate checksum content, given provided inputs
		var err error
		var sum []byte
		for _, source := range config.sources {
			if sum, err = checksumSource(os.Stdout, source); err != nil {
				if err == os.ErrNotExist {
					os.Stderr.WriteString("sha256sum: " + source + ": No such file or directory\n")
				} else if err == os.ErrInvalid {
					os.Stderr.WriteString("sha256sum: " + source + ": Is a directory\n")
				} else {
					os.Stderr.WriteString("sha256sum: " + err.Error() + "\n")
				}
				exitCode = 1
				continue
			}
			assert.Success(err, "Unexpected failure reading content from '"+source+"': %v")
			writeChecksum(os.Stdout, sum, source)
		}
	}
	os.Exit(exitCode)
}

type config struct {
	checkMode bool
	sources   []string
	quiet     bool
}

func initConfig() *config {
	_ = flag.Bool("b", true, "binary mode (no-op)")
	c1 := flag.Bool("c", false, "verify existing checksums")
	c2 := flag.Bool("check", false, "verify existing checksums")
	_ = flag.Bool("t", false, "text mode (no-op)")
	quiet := flag.Bool("quiet", false, "Silence reporting output.")
	flag.Parse()

	var c config
	c.checkMode = *c1 || *c2
	if flag.NArg() == 0 {
		c.sources = []string{"-"}
	} else {
		c.sources = append(c.sources, flag.Args()...)
	}
	c.quiet = *quiet
	return &c
}

func verifyConfig(config *config) error {
	if !config.checkMode && config.quiet {
		os.Stderr.WriteString("sha256sum: the --quiet option is meaningful only when verifying checksums\n")
		return os.ErrInvalid
	}
	return nil
}

var checksumLineFormat = regexp.MustCompile(`^([0-9a-f]{64}) \*(.+)$`)

func verifySource(c *config, source string) (uint, error) {
	var reader *bufio.Reader
	if source == "-" {
		reader = bufio.NewReader(os.Stdin)
	} else {
		f, err := os.Open(source)
		if err != nil {
			return 0, os.ErrNotExist
		}
		defer io_.CloseLogged(f, "Failed to close source: %+v")
		stat, err := f.Stat()
		assert.Success(err, "Failed to acquire source metadata: %v")
		if stat.IsDir() {
			return 0, os.ErrInvalid
		}
		reader = bufio.NewReader(f)
	}
	found := uint(0)
	failures := uint(0)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		assert.Success(err, "Unexpected interruption while reading content: %v")
		matches := checksumLineFormat.FindStringSubmatch(strings.TrimRight(line, "\n\r"))
		if matches == nil {
			continue
		}
		found++
		fileName := strings.TrimSpace(matches[2])
		var actual []byte
		if fileName == "-" {
			actual, err = checksum(os.Stdin)
		} else {
			f, err := os.Open(fileName)
			if err != nil {
				return 0, os.ErrNotExist
			}
			actual, err = checksum(f)
			assert.Success(err, "sha256sum: "+fileName+": failed to read all content")
			f.Close()
		}
		if fmt.Sprintf("%064x", actual) != matches[1] {
			failures++
			writeResult(c.quiet, fileName, false)
			continue
		}
		writeResult(c.quiet, fileName, true)
	}
	if found == 0 {
		return 0, io.ErrNoProgress
	}
	return failures, nil
}

func writeResult(quiet bool, source string, success bool) {
	if success && !quiet {
		os.Stdout.WriteString(source + ": OK\n")
	} else if !success {
		os.Stdout.WriteString(source + ": FAILED\n")
	}
}

func checksumSource(out io.Writer, source string) ([]byte, error) {
	var in io.Reader
	if source == "-" {
		in = os.Stdin
	} else {
		var err error
		var f *os.File
		if f, err = os.Open(source); err != nil {
			return nil, os.ErrNotExist
		}
		defer io_.CloseLogged(f, "Failed to close source: %+v")
		stat, err := f.Stat()
		assert.Success(err, "Failed to acquire file metadata: %v")
		if stat.IsDir() {
			return nil, os.ErrInvalid
		}
		in = f
	}
	return checksum(in)
}

func checksum(in io.Reader) ([]byte, error) {
	checksum := sha256.New()
	if _, err := io.Copy(checksum, in); err != nil {
		return nil, err
	}
	return checksum.Sum(nil), nil
}

func writeChecksum(out io.Writer, checksum []byte, name string) {
	_, err := out.Write([]byte(fmt.Sprintf("%064x *%s\n", checksum, name)))
	assert.Success(err, "Failed to write checksum line to stdout: %v")
}
