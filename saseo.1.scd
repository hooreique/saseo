saseo(1)

# NAME

saseo - add and remove small rc snippets that should stick around for a while, but not forever

# SYNOPSIS

*saseo* [*--dry-run*] [*--marker* _MARKER_] [*--*] _FILE_

*saseo* *--rm* [*--dry-run*] [*--marker* _MARKER_] [*--*] _FILE_

*saseo* *-h*++
*saseo* *--help*++
*saseo* *-V*++
*saseo* *--version*

# DESCRIPTION

*saseo* reads a block body from stdin and inserts it into an existing file as a
marked block. If the target file already has that block, *saseo* replaces it.
With *--rm*, *saseo* removes the marked block instead.

By default, the block is delimited by these exact whole lines:

```
##SASEO^
...
##SASEO$
```

The marker can be changed with *--marker*.

The target file must already exist. Symlink targets and target files using CRLF
line endings are rejected.

# OPTIONS

*--dry-run*
	Validate the request and print the same success output without writing the
	target file.

*--rm*
	Remove the managed block. Stdin is ignored.

*--marker* _MARKER_
	Use _MARKER_ instead of *SASEO*. Marker names may use a-z, A-Z, 0-9,
	underscore, period, colon, and hyphen, but may not start with a hyphen.

*--*
	Stop option parsing before _FILE_. For compatibility, if the first argument
	is *--* and the next argument is a help or version option, the first *--* is
	ignored.

*-h*, *--help*
	Print help. When stdout is a terminal, section headings and option names
	may be colored. Set *NO_COLOR* to disable color.

*-V*, *--version*
	Print the version.

# OUTPUT

Successful add and replace operations print:

```
START_LINE LINE_COUNT
```

Successful remove operations print this when a block is actually removed:

```
START_LINE
```

Line numbers are 1-based. For add and replace, _LINE_COUNT_ includes the start
and end marker lines. Removing an absent block is a successful no-op and prints
nothing.

# EXIT STATUS

*0*
	Success.

*3*
	Invalid option or argument shape.

*4*
	Insert or replace requested, but stdin is empty.

*5*
	Target file does not exist.

*6*
	Invalid marker name, or input body contains a marker delimiter line.

*7*
	Invalid marker state in the target file.

*8*
	Unsupported target file state, such as a symlink target or CRLF line
	endings.

Other non-zero statuses are not classified by *saseo*.

# EXAMPLES

Add or replace the default block:

```
saseo ~/.bashrc <<'EOF'
export FOO=foo
export BAR=bar
EOF
```

Use a named marker:

```
saseo --marker WORK ~/.bashrc <<'EOF'
export AWS_PROFILE=work
EOF
```

Remove a named block:

```
saseo --rm --marker WORK ~/.bashrc
```

Preview a change:

```
saseo --dry-run --marker WORK ~/.bashrc <<'EOF'
export AWS_PROFILE=work
EOF
```

# SEE ALSO

*saseo-spec.md*
