# SPDX-License-Identifier: GPL-3.0-only

MAKEFLAGS += --no-builtin-rules
.SUFFIXES:

.PHONY: all
all: download-metadata download-signatures extract-keyid extract-fingerprint sha256sum canonicalize-keysmap

download-metadata: go.mod cmd/download-metadata/*.go
	go build ./cmd/download-metadata

download-signatures: go.mod cmd/download-signatures/*.go
	go build ./cmd/download-signatures

extract-keyid: go.mod cmd/extract-keyid/*.go
	go build ./cmd/extract-keyid

extract-fingerprint: go.mod cmd/extract-fingerprint/*.go
	go build ./cmd/extract-fingerprint

sha256sum: go.mod cmd/sha256sum/*.go
	go build ./cmd/sha256sum

canonicalize-keysmap: go.mod cmd/canonicalize-keysmap/*.go
	go build ./cmd/canonicalize-keysmap

.PHONY: clean
clean:
	rm -f download-metadata download-signatures extract-keyid extract-fingerprint sha256sum canonicalize-keysmap
