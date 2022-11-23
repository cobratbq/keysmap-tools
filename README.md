# keysmap-tools

Tools for automatically generating PGP keys map for pgpverify-maven-plugin.

## TODO

- ☐ (2022-11-23) issue w.r.t. lack of support for EdDSA public keys in [`golang.org/x/crypto`](<https://cs.opensource.google/go/x/crypto/+/master:openpgp/packet/packet.go;l=445;drc=0a44fdfbc16e146f50e5fb8823fcc5ac186049b2> "Current HEAD revision, public key algorithm 22 missing"):
  ```
  goroutine 1 [running]:
  github.com/cobratbq/goutils/std/builtin.Require(...)
  	/home/danny/.local/share/go/pkg/mod/github.com/cobratbq/goutils@v0.0.0-20200226134721-3d3adb53ada8/std/builtin/require.go:17
  github.com/cobratbq/goutils/std/builtin.RequireSuccess(...)
  	/home/danny/.local/share/go/pkg/mod/github.com/cobratbq/goutils@v0.0.0-20200226134721-3d3adb53ada8/std/builtin/require.go:11
  main.main()
  	/home/danny/dev/java/keysmap/tools/cmd/extract-keyid/main.go:23 +0x35c
  panic: failed to extract signature body%!(EXTRA errors.UnsupportedError=openpgp: unsupported feature: public key algorithm 22)
  ```
- ☐ grouping for entries with same groupID + version, but varying artifactID. Useful for simultaneously released modules, all with same version.

