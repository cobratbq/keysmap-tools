/* SPDX-License-Identifier: GPL-3.0-only */

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/cobratbq/goutils/assert"
	io_ "github.com/cobratbq/goutils/std/io"
	gocryptoarmor "golang.org/x/crypto/openpgp/armor"
	gocryptopacket "golang.org/x/crypto/openpgp/packet"
)

func main() {
	content, err := ioutil.ReadAll(os.Stdin)
	assert.Success(err, "Failed to read content from signature file")
	err = readSignaturePacket(bytes.NewBuffer(content))
	if err != io.ErrUnexpectedEOF {
		return
	}
	err = readLegacyPacket(bytes.NewBuffer(content))
	assert.Success(err, "Failed to read signature.")
}

// readSignaturePacket reads a signature packet and extracts the issuer key-id. ProtonMail/go-crypto
// cannot work with SignatureV3 packets (legacy format).
func readSignaturePacket(in io.Reader) error {
	block, err := armor.Decode(in)
	if err == io.EOF {
		return err
	}
	defer io_.Discard(block.Body)
	assert.Success(err, "failed to read signature")
	pkt, err := packet.NewReader(block.Body).Next()
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	assert.Success(err, "failed to extract signature body")
	switch sig := pkt.(type) {
	case *packet.Signature:
		os.Stdout.WriteString(fmt.Sprintf("%016X\n", *sig.IssuerKeyId))
	default:
		panic(fmt.Sprintf("Unsupported type: %#v", sig))
	}
	return nil
}

// readLegacyPacket reads an openpgp signature. readLegacyPacket exists to handle SignatureV3, the
// old signature format that ProtonMail/go-crypto does not support.
func readLegacyPacket(in io.Reader) error {
	block, err := gocryptoarmor.Decode(in)
	if err == io.EOF {
		return err
	}
	defer io_.Discard(block.Body)
	assert.Success(err, "failed to read signature")
	pkt, err := gocryptopacket.NewReader(block.Body).Next()
	assert.Success(err, "failed to extract signature body")
	switch sig := pkt.(type) {
	case *gocryptopacket.Signature:
		panic("BUG: we would expect ProtonMail openpgp to have processed this type of signature.")
	case *gocryptopacket.SignatureV3:
		os.Stdout.WriteString(fmt.Sprintf("%016X\n", sig.IssuerKeyId))
	default:
		panic(fmt.Sprintf("Unsupported type: %#v", sig))
	}
	return nil
}
