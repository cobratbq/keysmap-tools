# SPDX-License-Identifier: GPL-3.0-only

.PHONY: all
all: download-metadata download-signatures extract-keyid extract-fingerprint sha256sum canonicalize-keysmap

download-metadata: base cmd/download-metadata/*.go
	go build ./cmd/download-metadata

download-signatures: base cmd/download-signatures/*.go
	go build ./cmd/download-signatures

extract-keyid: base cmd/extract-keyid/*.go
	go build ./cmd/extract-keyid

extract-fingerprint: base cmd/extract-fingerprint/*.go
	go build ./cmd/extract-fingerprint

sha256sum: base cmd/sha256sum/*.go
	go build ./cmd/sha256sum

canonicalize-keysmap: base cmd/canonicalize-keysmap/*.go
	go build ./cmd/canonicalize-keysmap

base: go.mod go.sum

.PHONY: clean
clean:
	rm -f download-metadata download-signatures extract-keyid extract-fingerprint sha256sum canonicalize-keysmap
