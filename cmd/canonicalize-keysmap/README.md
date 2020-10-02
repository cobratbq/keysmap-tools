# README

A program for canonicalizing the keysmap entries into ranges for conciseness.

## Design

- Deterministic ordered generation of pgp-keys map.
- Prioritize special-cases 'noSig' and 'noKey'.
- Group all public keys for any version of an artifact, i.e. `groupID:artifactID = key1, key2, key3, ...`.
  - assumes that untrusted keys are revoked.
  - assumes that once public key is used to sign an artifact version once, it may reappear for future versions.

## Versions

- Separation of version components:
  - `.` separates version components.
  - Transition from digit to alpha or from alpha to digit is considered implicit separation of version components.
- Dash (`-`) separates between levels:  
  Interpret '`-`' as raising "sublevel" integer with one. One never goes back to previous level. Part of version before '`-`' remains at current level, while everything after '`-`' will go one "sublevel" up. (Consider "sublevel" as being lower in preference, hence `2.0-1 < 2.0.1`)  
  Dash-separators are typically used to indicate a second iteration of packaging a single version, for example to tackle issues with forgotten dependencies. `1.0-1` is a first attempt at packaging version `1.0`, while `1.0-2` is the second attempt.
- missing component implies `0` (digit)/`` (empty string alpha).

### Order of priority

Equality of different version specifications:

```
1 = 1.0 = 1-0 = 1.0-0 = 1.0-GA
```

The following priority ordering ensures different versions (version components) are ordered correctly. This assumes that version components are encountered at same "sublevel", meaning same number of '`-`' have been encountered.

```text
"a"/"alpha" < "b"/"beta" < "m"/"milestone" < "rc"/"cr" < "snapshot" < ""/"ga"/"final"/"release"/0 < "sp" < any non-predefined string < positive integer value
```

Ordering under sublevel difference:

```text
1.0-SNAPSHOT < 1.0.0 = 1.0-0 = 1.0-ga < 1.0-magic < 1.0-1 < 1.0.1 = 1.0.1-0 < 1.0.1-1
```

## References

- https://cwiki.apache.org/confluence/display/MAVENOLD/Versioning
- https://maven.apache.org/ref/3.6.3/maven-artifact/xref/org/apache/maven/artifact/versioning/ComparableVersion.html
