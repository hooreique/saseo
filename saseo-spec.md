# saseo spec

`saseo` provides bash-aware block management for shell rc files.

## Command [SASEO-CMD]

```bash
saseo put  [--dry-run] [--marker MARKER] [--plain-text] [--] FILE
saseo mark [--dry-run] [--marker MARKER] [--plain-text] [--] FILE
saseo rm   [--dry-run] [--marker MARKER] [--plain-text] [--] FILE
saseo show [--marker MARKER] [--] FILE
saseo --help
saseo put --help
saseo mark --help
saseo rm --help
saseo show --help
saseo --version
```

`put` adds or replaces a managed block from stdin. `mark` adds or replaces a
managed block with empty content. `rm` removes a managed block. `show` prints
the current managed block.

A subcommand is required for edit operations. If no argument is provided,
`saseo` prints the root help and exits successfully.

`--` before `FILE` stops option parsing, so filenames that start with `-` can
be used.

`--version` prints the version from the embedded `VERSION` file.

`--help` prints command usage and options. `put --help`, `mark --help`,
`rm --help`, and `show --help` print subcommand-specific help.

For `put`, `mark`, and `rm`, `--dry-run` runs the same validation and computes
the same success output as a normal operation, but does not write or rename the
target file.

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

Exact delimiter lines may appear in shell command context, inside heredoc
bodies, or inside multiline shell strings.

## Add Or Replace Mode [SASEO-ADD-REPLACE]

`put` reads the replacement body from stdin. If stdin is empty, `saseo` fails
with exit code `4` and prints error details to stderr.

Example:

```bash
saseo put ~/.bashrc <<'EOF'
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

## Mark Mode [SASEO-MARK]

`mark` does not read stdin. Piped or redirected stdin is ignored.

Example:

```bash
saseo mark --marker WORK ~/.bashrc
```

If the target file does not contain the expected marker block, `saseo` appends
a new managed block with no body lines:

```text
##WORK^
##WORK$
```

If the target file already contains the expected marker block, `saseo` replaces
the whole existing block, including the markers and body, with a new empty
managed block.

Marker creation and marker reuse are quiet. No human-readable informational
message is printed for either case.

## Append Formatting [SASEO-APPEND-FORMATTING]

When appending a new block:

- if the existing file ends with a newline, the start marker begins immediately
  on the next line,
- if the existing file does not end with a newline, `saseo` inserts one newline
  before the start marker.

No extra blank separator line is inserted.

In shell mode, append is only valid at top-level shell context at the end of the
file.

## Remove Mode [SASEO-REMOVE]

`rm` removes the managed block and does not read stdin. Piped or redirected
stdin is ignored.

Example:

```bash
saseo rm --marker SASEO ~/.bashrc
```

The start marker, body, and end marker are removed. If the end marker is
followed by one newline, that newline is removed too.

If no marker block exists, the file is left unchanged and no success output is
printed. The target file is not rewritten or replaced in this no-op case.

## Show Mode [SASEO-SHOW]

`show` prints the current managed block and does not read stdin. Piped or
redirected stdin is ignored.

Example:

```bash
saseo show --marker SASEO ~/.bashrc
```

If the target file contains the expected marker block, `show` prints:

```text
START_LINE LINE_COUNT
##<MARKER>^
...
##<MARKER>$
```

`START_LINE` is the 1-based line number where the managed block starts.
`LINE_COUNT` is the number of lines in the managed block, including the start
and end marker delimiter lines. The next `LINE_COUNT` output lines are the
current managed block contents, including marker delimiter lines.

If no marker block exists, `show` prints only:

```text
0
```

The no-block case exits successfully with exit code `0`.

`show` never writes or renames the target file.

Because `show` does not edit the target file or generate shell code, it does
not perform shell syntax or shell boundary validation. It still validates marker
names, target file support, LF target line endings, and marker state.

## Plain-Text Mode [SASEO-PLAIN-TEXT]

For `put`, `mark`, and `rm`, `--plain-text` disables shell syntax and shell
boundary validation.

Plain-text mode still validates target state, marker names, marker delimiters,
target file support, stdin emptiness for `put`, and exact body delimiter lines.

Use plain-text mode when editing non-shell text files.

## Shell Validation [SASEO-SHELL-VALIDATION]

Without `--plain-text`, `put`, `mark`, and `rm` treat the target as bash
syntax.

The target file must parse before any `put`, `mark`, `rm`, or edit dry-run
operation. For `put` and `mark`, the generated output must also parse. For
`rm`, the generated output must parse when a block is removed.

Shell syntax failures fail with exit code `9`.

zsh files are not officially supported. They are checked with the bash parser;
zsh-only syntax may be rejected.

## Shell Boundary [SASEO-SHELL-BOUNDARY]

In shell mode, the start and end marker delimiters must be in the same shell
syntax context.

Allowed examples include a managed block entirely inside:

- top-level command context,
- a subshell,
- a heredoc body,
- a multiline shell string.

The managed block must not open or close the syntactic construct that surrounds
it. For example, if an existing file provides parentheses and the managed block
is represented as `<>`, then `(<>)` is valid, but `<(>)` and `(<)>` are invalid.

When an existing marker block is present, `saseo` validates the existing marker
boundary before editing. If it is already illegal, the operation is rejected.

When replacing a block, the generated marker boundary must also be valid. A
replacement body that closes a surrounding shell construct and reopens another
one is rejected even if the final file parses.

Shell boundary failures fail with exit code `10`.

## Dry Run Mode [SASEO-DRY-RUN]

When `--dry-run` is provided to `put`, `mark`, or `rm`, `saseo` validates the
request and computes the same generated output as a normal run, but leaves the
target file unchanged.

Successful dry-run `put` and `mark` operations print the same success output as
successful real `put` and `mark` operations. Successful dry-run `rm` operations
print the same success output as successful real `rm` operations.

## Target File [SASEO-TARGET-FILE]

`FILE` is expanded to an absolute path before reading or writing.

The target file must already exist. If it does not exist, `saseo` fails with
exit code `5` and prints error details to stderr.

`FILE` itself must not be a symlink. Symlink targets fail with exit code `8`.

The target file must be a regular file. Non-regular files fail with exit code
`8`.

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

In shell mode, shell syntax validation may fail before marker-state validation
if the target file itself is not parseable as bash.

## Invalid Options [SASEO-INVALID-OPTIONS]

The root command accepts only `put`, `mark`, `rm`, `show`, `--help`, and
`--version`.

The `put`, `mark`, and `rm` subcommands accept only `--dry-run`, `--marker`,
`--marker=MARKER`, `--plain-text`, `--help`, and `--`.

The `show` subcommand accepts only `--marker`, `--marker=MARKER`, `--help`, and
`--`.

Root `--help`, root `--version`, and subcommand `--help` forms do not accept
additional arguments.

For subcommand operations, options must appear before `FILE`. After `FILE` has
been read, any additional argument is invalid. After `--`, exactly one `FILE`
argument must follow.

Each subcommand option may appear at most once.

Any invalid option or argument shape fails with exit code `3`.

## Error Output [SASEO-ERROR-OUTPUT]

Classified failures print an `error:` line to stderr. When the next user action
is clear, stderr should include actionable `hint:` or `try:` guidance. Exact
wording is not part of the stable interface; callers should use exit codes for
machine-readable handling.

## Success Output [SASEO-AFFECTED-LINES]

For successful `put` and `mark`, including dry runs, `saseo` prints two numbers
to stdout:

```text
START_LINE LINE_COUNT
```

`START_LINE` is the 1-based line number where the inserted managed block starts.
`LINE_COUNT` is the number of lines occupied by the inserted managed block,
including the start and end marker delimiter lines. It is not a count of changed
content lines in the pre-existing file. For `mark`, `LINE_COUNT` is `2`.

For successful `rm`, including dry runs, when a block is actually removed,
`saseo` prints only:

```text
START_LINE
```

For successful `show`, when a block exists, `saseo` prints:

```text
START_LINE LINE_COUNT
```

followed by `LINE_COUNT` lines containing the current managed block, including
the marker delimiter lines. When no block exists, `saseo` prints only:

```text
0
```

Line counting treats a line as ending with a newline. The final trailing newline
does not create an extra counted line.

## Exit Codes [SASEO-EXIT-CODES]

- `3`: invalid option or argument shape,
- `4`: `put` requested, but stdin is empty,
- `5`: target file does not exist,
- `6`: invalid marker name or input body contains a marker delimiter line,
- `7`: invalid marker state in the target file,
- `8`: unsupported target file state, such as a symlink target, non-regular
  file, or CRLF line endings,
- `9`: shell syntax validation failed,
- `10`: shell marker boundary validation failed.
