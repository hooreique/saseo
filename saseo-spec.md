# saseo spec

`saseo` adds and removes small rc snippets that should stick around for a while,
but not forever.

## Command [SASEO-CMD]

```bash
saseo [--dry-run] [--marker MARKER] [--] FILE
saseo --rm [--dry-run] [--marker MARKER] [--] FILE
saseo -h
saseo --help
saseo -V
saseo --version
saseo -- -h
saseo -- --help
saseo -- -V
saseo -- --version
```

`--` before `FILE` stops option parsing, so filenames that start with `-` can
be used.

For compatibility with launchers and older examples, if the first argument is
`--` and the next argument is a known option, the first `--` is ignored. This
keeps forms such as `saseo -- --version` working.

`-V` and `--version` print the version from the `VERSION` file.

`-h` and `--help` print usage, options, a `man saseo` details hint, and the
example from the `EXAMPLE` file.

`--dry-run` runs the same validation and computes the same success output as a
normal add, replace, or remove operation, but does not write or rename the target
file.

## Marker Name [SASEO-MARKER-NAME]

`MARKER` must match:

```text
^[A-Za-z0-9_.:][A-Za-z0-9_.:-]*$
```

In words:

- `a-z`, `A-Z`, `0-9`, `_`, `.`, `:`, and `-` are allowed,
- `-` is not allowed as the first character.

`--marker` is optional. If omitted, the marker is `SASEO`.

Invalid marker names fail with exit code `6` and print error details to
stderr.

## Marker Format [SASEO-MARKER-FORMAT]

For `--marker SASEO`, the managed block is delimited by exact whole lines:

```text
##SASEO^
...
##SASEO$
```

In general:

```text
##<MARKER>^
...
##<MARKER>$
```

Marker text embedded inside another line is not a delimiter. For example,
`echo "##SASEO^"` is ordinary content.

## Add Or Replace Mode [SASEO-ADD-REPLACE]

When `--rm` is not provided, `saseo` reads the replacement body from stdin.
If stdin is empty, `saseo` fails with exit code `4` and prints error details to
stderr.

Example:

```bash
saseo ~/init.zsh <<'EOF'
export FOO=foo
export BAR=bar
EOF
```

If the target file does not contain the expected marker block, `saseo` appends
a new managed block to the end of the file.

If the target file already contains the expected marker block, `saseo` replaces
the whole existing block, including the markers, with a new managed block.

The inserted body is normalized to LF line endings and to end with a newline
before the end marker.

If the body contains the current marker's start or end delimiter as an exact
whole line, `saseo` fails with exit code `6`.

Marker creation and marker reuse are quiet. No human-readable informational
message is printed for either case.

## Append Formatting [SASEO-APPEND-FORMATTING]

When appending a new block:

- if the existing file ends with a newline, the start marker begins immediately
  on the next line,
- if the existing file does not end with a newline, `saseo` inserts one newline
  before the start marker.

No extra blank separator line is inserted.

## Remove Mode [SASEO-REMOVE]

When `--rm` is provided, `saseo` removes the managed block and does not read
stdin. Piped or redirected stdin is ignored.

Example:

```bash
saseo --rm --marker SASEO ~/init.zsh
```

The start marker, body, and end marker are removed. If the end marker is
followed by one newline, that newline is removed too.

If no marker block exists, the file is left unchanged and no success output is
printed. The target file is not rewritten or replaced in this no-op case.

## Dry Run Mode [SASEO-DRY-RUN]

When `--dry-run` is provided, `saseo` validates the request and computes the
same generated output as a normal run, but leaves the target file unchanged.

Successful dry-run add and replace operations print the same success output as
successful real add and replace operations. Successful dry-run remove operations
print the same success output as successful real remove operations.

## Target File [SASEO-TARGET-FILE]

`FILE` is expanded as a path before reading or writing.

The target file must already exist. If it does not exist, `saseo` fails with
exit code `5` and prints error details to stderr.

`FILE` itself must not be a symlink. Symlink targets fail with exit code `8`.

The target file must use LF line endings. If CRLF is detected in the existing
target file, `saseo` fails with exit code `8`.

`saseo` writes generated output with LF line endings. It writes to a temporary
file in the same directory, applies the target file's basic rwx mode to the
temporary file, then renames the temporary file over the target. Owner and group
preservation are not guaranteed.

If the generated output is byte-for-byte identical to the existing target
contents, `saseo` does not write or rename the target file.

## Invalid Marker State [SASEO-INVALID-MARKER-STATE]

A marker block is valid only when:

- the start marker line exists,
- the end marker line exists,
- the end marker line appears after the start marker line.

Invalid marker states fail as follows:

- start marker exists but end marker does not: exit code `7`,
- end marker exists but start marker does not: exit code `7`,
- end marker appears before start marker: exit code `7`,
- start marker or end marker appears more than once: exit code `7`.

Each failure prints error details to stderr.

## Invalid Options [SASEO-INVALID-OPTIONS]

Only `--dry-run`, `--rm`, `--marker`, `-h`, `--help`, `-V`, `--version`, and
`--` are accepted as options. Any invalid option or argument shape fails with
exit code `3`.

## Error Output [SASEO-ERROR-OUTPUT]

Classified failures print an `error:` line to stderr. When the next user action
is clear, stderr should include actionable `hint:` or `try:` guidance. Exact
wording is not part of the stable interface; callers should use exit codes for
machine-readable handling.

## Success Output [SASEO-AFFECTED-LINES]

For successful add or replace mode, including dry runs, `saseo` prints two
numbers to stdout:

```text
START_LINE LINE_COUNT
```

`START_LINE` is the 1-based line number where the inserted managed block starts.
`LINE_COUNT` is the number of lines occupied by the inserted managed block,
including the start and end marker delimiter lines. It is not a count of changed
content lines in the pre-existing file.

For successful remove mode, including dry runs, when a block is actually
removed, `saseo` prints only:

```text
START_LINE
```

Line counting treats a line as ending with a newline. The final trailing newline
does not create an extra counted line.

## Exit Codes [SASEO-EXIT-CODES]

- `3`: invalid option or argument shape,
- `4`: insert or replace requested, but stdin is empty,
- `5`: target file does not exist,
- `6`: invalid marker name or input body contains a marker delimiter line,
- `7`: invalid marker state in the target file,
- `8`: unsupported target file state, such as symlink or CRLF line endings.
