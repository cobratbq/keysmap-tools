.PHONY: all
all: download-metadata download-signatures extract-keyid extract-fingerprint

download-metadata:
	go build ./cmd/download-metadata

download-signatures:
	go build ./cmd/download-signatures

extract-keyid:
	go build ./cmd/extract-keyid

extract-fingerprint:
	go build ./cmd/extract-fingerprint

.PHONY: clean
clean:
	rm -f download-metadata download-signatures extract-keyid extract-fingerprint
