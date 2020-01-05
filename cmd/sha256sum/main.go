package main

import (
	"bufio"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	config := initConfig()

	exitCode := 0
	if config.checkMode {
		failures := uint(0)
		for _, source := range config.sources {
			sourceFails, err := verifySource(source)
			if err != nil {
				panic("Unexpected failure verifying source '" + source + "': " + err.Error())
			}
			failures += sourceFails
		}
		if failures > 0 {
			exitCode = 1
			os.Stderr.WriteString(fmt.Sprintf("sha256sum: WARNING: %d computed checksum(s) did NOT match\n", failures))
		}
	} else {
		var err error
		var sum []byte
		for _, source := range config.sources {
			if sum, err = checksumSource(os.Stdout, source); err == os.ErrInvalid {
				exitCode = 1
				continue
			}
			expectSuccess(err, "Unexpected failure reading content from '"+source+"': %v")
			writeChecksum(os.Stdout, sum, source)
		}
	}
	os.Exit(exitCode)
}

type config struct {
	checkMode bool
	sources   []string
}

func initConfig() *config {
	_ = flag.Bool("b", true, "binary mode")
	c1 := flag.Bool("c", false, "verify existing checksums")
	c2 := flag.Bool("check", false, "")
	_ = flag.Bool("t", false, "text mode")
	_ = flag.Bool("quiet", false, "Silence reporting output.")
	// _ = flag.Bool("z", false, "end each output line with NUL")
	flag.Parse()

	var c config
	c.checkMode = *c1 || *c2
	if flag.NArg() == 0 {
		c.sources = []string{"-"}
	} else {
		c.sources = append(c.sources, flag.Args()...)
	}
	return &c
}

func verifySource(source string) (uint, error) {
	var reader *bufio.Reader
	if source == "-" {
		reader = bufio.NewReader(os.Stdin)
	} else {
		f, err := os.Open(source)
		if err != nil {
			return 0, os.ErrNotExist
		}
		defer closeLogged(f)
		stat, err := f.Stat()
		expectSuccess(err, "Failed to acquire source metadata: %v")
		if stat.IsDir() {
			return 0, os.ErrInvalid
		}
		reader = bufio.NewReader(f)
	}
	failures := uint(0)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			// FIXME how to handle error
			panic(err.Error())
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			failures++
			// FIXME check correct response
			panic("To be implemented.")
			continue
		}
		expected := strings.TrimSpace(parts[0])
		fileName := strings.TrimLeft(strings.TrimSpace(parts[1]), "*")
		var actual []byte
		if fileName == "-" {
			actual, err = checksum(os.Stdin)
		} else {
			f, err := os.Open(fileName)
			if err != nil {
				// FIXME check correct response
				panic("To be implemented.")
			}
			actual, err = checksum(f)
			if err != nil {
				// FIXME check correct response
				panic("To be implemented.")
			}
			f.Close()
		}
		if fmt.Sprintf("%064x", actual) != expected {
			failures++
			writeResult(fileName, "FAILED")
			continue
		}
		writeResult(fileName, "OK")
	}
	return failures, nil
}

func writeResult(source, result string) {
	os.Stdout.WriteString(source + ": " + result + "\n")
}

func checksumSource(out io.Writer, source string) ([]byte, error) {
	var in io.Reader
	if source == "-" {
		in = os.Stdin
	} else {
		var err error
		var f *os.File
		if f, err = os.Open(source); err != nil {
			// FIXME how to handle error here, exit code?
			return nil, os.ErrNotExist
		}
		defer closeLogged(f)
		stat, err := f.Stat()
		expectSuccess(err, "Failed to acquire file metadata: %v")
		if stat.IsDir() {
			os.Stderr.WriteString("sha256sum: " + source + ": Is a directory\n")
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
	expectSuccess(err, "Failed to write checksum line to stdout: %v")
}

func expectSuccess(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf(msg, err))
	}
}

func closeLogged(c io.Closer) {
	if err := c.Close(); err != nil {
		panic("Failed to close: " + err.Error())
	}
}
