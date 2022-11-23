/* SPDX-License-Identifier: GPL-3.0-only */

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/cobratbq/goutils/assert"
	io_ "github.com/cobratbq/goutils/std/io"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func main() {
	block, err := armor.Decode(os.Stdin)
	if err == io.EOF {
		return
	}
	assert.Success(err, "failed to read signature")
	pkt, err := packet.NewReader(block.Body).Next()
	assert.Success(err, "failed to extract signature body")
	switch sig := pkt.(type) {
	case *packet.Signature:
		os.Stdout.WriteString(fmt.Sprintf("%016X\n", *sig.IssuerKeyId))
	case *packet.SignatureV3:
		os.Stdout.WriteString(fmt.Sprintf("%016X\n", sig.IssuerKeyId))
	default:
		panic(fmt.Sprintf("Unsupported type: %#v", sig))
	}
	io_.Discard(block.Body)
}
