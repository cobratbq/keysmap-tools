package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func main() {
	block, err := armor.Decode(os.Stdin)
	if err == io.EOF {
		return
	} else if err != nil {
		panic("Unexpected error: " + err.Error())
	}
	pkt, err := packet.NewReader(block.Body).Next()
	expectSuccess(err)
	switch sig := pkt.(type) {
	case *packet.Signature:
		os.Stdout.WriteString(fmt.Sprintf("%016X\n", *sig.IssuerKeyId))
	case *packet.SignatureV3:
		os.Stdout.WriteString(fmt.Sprintf("%016X\n", sig.IssuerKeyId))
	default:
		panic(fmt.Sprintf("Unsupported type: %#v", sig))
	}
	io.Copy(ioutil.Discard, block.Body)
}

func expectSuccess(err error) {
	if err != nil {
		panic(err.Error())
	}
}
