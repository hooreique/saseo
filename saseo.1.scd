saseo(1)

# NAME

saseo - bash-aware block management for shell rc files

# SYNOPSIS

*saseo* *put* [*--dry-run*] [*--marker* _MARKER_] [*--plain-text*] [*--*] _FILE_

*saseo* *mark* [*--dry-run*] [*--marker* _MARKER_] [*--plain-text*] [*--*] _FILE_

*saseo* *rm* [*--dry-run*] [*--marker* _MARKER_] [*--plain-text*] [*--*] _FILE_

*saseo* *show* [*--marker* _MARKER_] [*--*] _FILE_

*saseo* *--help*++
*saseo* *put* *--help*++
*saseo* *mark* *--help*++
*saseo* *rm* *--help*++
*saseo* *show* *--help*++
*saseo* *--version*

# DESCRIPTION

*saseo* lets scripts claim one marked region inside an existing shell rc file,
replace it safely, and remove it later.

*saseo put* reads a block body from stdin and inserts it into an existing file
as a marked block. If the target file already has that block, *saseo put*
replaces it.

*saseo mark* inserts an empty marked block. If the target file already has that
block, *saseo mark* replaces it with start and end marker lines only. Stdin is
ignored.

*saseo rm* removes the marked block. Stdin is ignored.

*saseo show* prints the current marked block, including marker lines. If the
block is absent, it prints *0*. Stdin is ignored.

By default, the block is delimited by these exact whole lines:

```
##SASEO^
...
##SASEO$
```

The marker can be changed with *--marker*.

By default, *saseo* parses the target as bash syntax before and after the edit.
The marker block may live inside a heredoc body, multiline shell string, or
subshell, but the marker block must not open or close that surrounding syntax.

Use *--plain-text* when editing non-shell text files.

zsh files are not officially supported. They are checked with the bash parser,
so zsh-only syntax may be rejected.

The target file must already exist. Symlink targets and target files using CRLF
line endings are rejected.

# OPTIONS

*--dry-run*
	Validate the request and print the same success output without writing the
	target file. This option applies to *put*, *mark*, and *rm*.

*--marker* _MARKER_
	Use _MARKER_ instead of *SASEO*. Marker names may use a-z, A-Z, 0-9,
	underscore, period, colon, and hyphen, but may not start with a hyphen.

*--plain-text*
	Skip shell syntax and shell boundary validation. Exact marker validation
	still applies. This option applies to *put*, *mark*, and *rm*.

*--*
	Stop option parsing before _FILE_.

*--help*
	Print root or subcommand help.

*--version*
	Print the version.

Root *--help*, root *--version*, and subcommand *--help* forms do not accept
additional arguments. For edit operations, options must appear before _FILE_.
After *--*, exactly one _FILE_ argument must follow. Each subcommand option may
appear at most once.

# OUTPUT

Successful *put* and *mark* operations print:

```
START_LINE LINE_COUNT
```

Successful *rm* operations print this when a block is actually removed:

```
START_LINE
```

Line numbers are 1-based. For *put* and *mark*, _LINE_COUNT_ includes the start
and end marker lines. For *mark*, _LINE_COUNT_ is *2*. Removing an absent block
is a successful no-op and prints nothing.

Successful *show* operations print this when a block exists:

```
START_LINE LINE_COUNT
##MARKER^
...
##MARKER$
```

The block content output contains exactly _LINE_COUNT_ lines after the first
line, including marker lines. If no block exists, *show* prints only:

```
0
```

The no-block case exits successfully. *show* does not write the target file and
does not perform shell syntax or shell boundary validation.

# EXIT STATUS

*0*
	Success.

*3*
	Invalid option or argument shape.

*4*
	*put* requested, but stdin is empty.

*5*
	Target file does not exist.

*6*
	Invalid marker name, or input body contains a marker delimiter line.

*7*
	Invalid marker state in the target file.

*8*
	Unsupported target file state, such as a symlink target, non-regular file,
	or CRLF line endings.

*9*
	Shell syntax validation failed.

*10*
	Shell marker boundary validation failed.

Other non-zero statuses are not classified by *saseo*.

# EXAMPLES

Add or replace the default block:

```
saseo put ~/.bashrc <<'EOF'
export FOO=foo
export BAR=bar
EOF
```

Use a named marker:

```
saseo put --marker WORK ~/.bashrc <<'EOF'
export AWS_PROFILE=work
EOF
```

Create an empty named block:

```
saseo mark --marker WORK ~/.bashrc
```

Remove a named block:

```
saseo rm --marker WORK ~/.bashrc
```

Show a named block:

```
saseo show --marker WORK ~/.bashrc
```

Preview a change:

```
saseo put --dry-run --marker WORK ~/.bashrc <<'EOF'
export AWS_PROFILE=work
EOF
```

Edit a non-shell text file:

```
saseo put --plain-text notes.txt <<'EOF'
temporary note
EOF
```

# SEE ALSO

*saseo-spec.md*
