/* SPDX-License-Identifier: GPL-3.0-only */

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/cobratbq/goutils/std/builtin"
	io_ "github.com/cobratbq/goutils/std/io"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func main() {
	block, err := armor.Decode(os.Stdin)
	if err == io.EOF {
		return
	}
	builtin.RequireSuccess(err, "failed to decode public key")
	pkt, err := packet.NewReader(block.Body).Next()
	builtin.RequireSuccess(err, "failed to read signature body")
	switch key := pkt.(type) {
	case *packet.PublicKey:
		os.Stdout.WriteString(fmt.Sprintf("0x%040X", key.Fingerprint))
	case *packet.PublicKeyV3:
		os.Stdout.WriteString(fmt.Sprintf("0x%040X", key.Fingerprint))
	default:
		panic(fmt.Sprintf("Unsupported type: %#v", key))
	}
	io_.Discard(block.Body)
}
