# SPDX-License-Identifier: GPL-3.0-only

.PHONY: all
all: download-metadata download-signatures extract-keyid extract-fingerprint

download-metadata: cmd/download-metadata/*.go
	go build ./cmd/download-metadata

download-signatures: cmd/download-signatures/*.go
	go build ./cmd/download-signatures

extract-keyid: cmd/extract-keyid/*.go
	go build ./cmd/extract-keyid

extract-fingerprint: cmd/extract-fingerprint/*.go
	go build ./cmd/extract-fingerprint

.PHONY: clean
clean:
	rm -f download-metadata download-signatures extract-keyid extract-fingerprint
