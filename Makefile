# SPDX-License-Identifier: GPL-3.0-only

.PHONY: all
all: download-metadata download-signatures extract-keyid extract-fingerprint sha256sum canonicalize-keysmap

download-metadata: cmd/download-metadata/*.go
	go build ./cmd/download-metadata

download-signatures: cmd/download-signatures/*.go
	go build ./cmd/download-signatures

extract-keyid: cmd/extract-keyid/*.go
	go build ./cmd/extract-keyid

extract-fingerprint: cmd/extract-fingerprint/*.go
	go build ./cmd/extract-fingerprint

sha256sum: cmd/sha256sum/*.go
	go build ./cmd/sha256sum

canonicalize-keysmap: cmd/canonicalize-keysmap/*.go
	go build ./cmd/canonicalize-keysmap

.PHONY: clean
clean:
	rm -f download-metadata download-signatures extract-keyid extract-fingerprint sha256sum canonicalize-keysmap
