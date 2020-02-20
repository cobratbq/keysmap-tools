# Versions

- Separation of version components:
  - `.` separates version components.
  - Transition from digit to alpha or from alpha to digit is considered implicit separation of version components.
- `-` separates between levels:  
  Interpret `-` as raising "sublevel" integer with one. One never goes back to previous level. Part of version before `-` remains at current level, while everything after `-` will go one "sublevel" up. (Consider "sublevel" as being lower in preference, hence `2.0-1 < 2.0.1`)
- missing component implies `0` (digit)/`` (empty string alpha).

## Order of priority

```
"a"/"alpha" < "b"/"beta" < "m"/"milestone" < "rc"/"cr" < "snapshot" < ""/"ga"/"final"/"release"/0 < "sp" < positive integer value
```

## References

- https://cwiki.apache.org/confluence/display/MAVENOLD/Versioning
- https://maven.apache.org/ref/3.6.3/maven-artifact/xref/org/apache/maven/artifact/versioning/ComparableVersion.html
