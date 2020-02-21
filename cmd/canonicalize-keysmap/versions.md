# Versions

- Separation of version components:
  - `.` separates version components.
  - Transition from digit to alpha or from alpha to digit is considered implicit separation of version components.
- `-` separates between levels:  
  Interpret `-` as raising "sublevel" integer with one. One never goes back to previous level. Part of version before `-` remains at current level, while everything after `-` will go one "sublevel" up. (Consider "sublevel" as being lower in preference, hence `2.0-1 < 2.0.1`)
- missing component implies `0` (digit)/`` (empty string alpha).

## Order of priority

The following priority ordering ensures different versions (version components) are ordered correctly. This assumes that version components are encountered at same "sublevel", meaning same number of '`-`' have been encountered.

```text
"a"/"alpha" < "b"/"beta" < "m"/"milestone" < "rc"/"cr" < "snapshot" < ""/"ga"/"final"/"release"/0 < "sp" < any non-predefined string < positive integer value
```

## References

- https://cwiki.apache.org/confluence/display/MAVENOLD/Versioning
- https://maven.apache.org/ref/3.6.3/maven-artifact/xref/org/apache/maven/artifact/versioning/ComparableVersion.html
