/* SPDX-License-Identifier: GPL-3.0-only */

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/cobratbq/goutils/assert"
	io_ "github.com/cobratbq/goutils/std/io"
)

func main() {
	block, err := armor.Decode(os.Stdin)
	if err == io.EOF {
		// do not silently accept that public key data is non-existent
		os.Exit(1)
	}
	assert.Success(err, "failed to decode public key")
	pkt, err := packet.NewReader(block.Body).Next()
	assert.Success(err, "failed to read signature body")
	switch key := pkt.(type) {
	case *packet.PublicKey:
		os.Stdout.WriteString(fmt.Sprintf("0x%040X", key.Fingerprint))
	default:
		panic(fmt.Sprintf("Unsupported type: %#v", key))
	}
	io_.Discard(block.Body)
}
