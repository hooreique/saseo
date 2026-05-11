# saseo

`saseo` provides bash-aware block management for shell rc files.

It lets scripts claim one marked region inside an existing rc file, replace it
safely, and remove it later. Unlike a plain text replacement, `saseo` parses bash
by default and rejects edits that would cross shell syntax boundaries.

<table>
<thead>
<tr>
<th>before</th>
<th>&gt;</th>
<th>after</th>
</tr>
</thead>
<tbody>
<tr>
<td valign="top">
<p><code>~/.bashrc</code>:</p>
<pre lang="bash"><code>export FOO=foo</code></pre>
<p>Suppose this is your existing rc file.</p>
</td>
<td valign="top">
<pre lang="bash"><code>saseo put ~/.bashrc &lt;&lt;'EOF'
export BAR=bar
export BAZ=baz
EOF</code></pre>
<p>Run <code>saseo</code> with the snippet to add.</p>
</td>
<td valign="top">
<p><code>~/.bashrc</code>:</p>
<pre lang="bash"><code>export FOO=foo
##SASEO^
export BAR=bar
export BAZ=baz
##SASEO$</code></pre>
<p>The file keeps its old content and gets one marked block.</p>
</td>
</tr>
</tbody>
</table>

It reads a block body from stdin, inserts or replaces the marked block in an
existing file, can create an empty marked block, can print the current marked
block, and leaves everything outside that block alone.

The name comes from the Korean word `사서` (`saseo`), meaning librarian.

## Getting Started

Prerequisite: _**Nix**_.

Run it once, without installing:

```bash
nix run github:hooreique/saseo -- --help
```

Install it into your Nix profile:

```bash
nix profile install github:hooreique/saseo
```

## Usage

Add or replace the default block:

```bash
saseo put ~/.bashrc <<'EOF'
export FOO=foo
export BAR=bar
EOF
```

Use a named marker:

```bash
saseo put --marker WORK ~/.bashrc <<< 'export AWS_PROFILE=work'
```

Create an empty marked block:

```bash
saseo mark --marker WORK ~/.bashrc
```

Remove a block:

```bash
saseo rm --marker WORK ~/.bashrc
```

Show the current block:

```bash
saseo show --marker WORK ~/.bashrc
```

Preview a change without writing the file:

```bash
saseo put --dry-run --marker WORK ~/.bashrc <<< 'export AWS_PROFILE=work'
saseo rm --dry-run --marker WORK ~/.bashrc
```

Show help or version:

```bash
saseo --help
saseo put --help
saseo mark --help
saseo rm --help
saseo show --help
saseo --version
```

If installed, the manual is available with `man saseo`.

## Why

Generated shell configuration often has nowhere clean to live.

A setup script may know the exact exports, aliases, or shell glue it wants to
install, but still should not take ownership of a whole `.bashrc`. A regex
replacement can work, until it edits the wrong text or leaves the file as invalid
shell. A full dotfile manager can work, but it is the wrong size when all you
need is one generated region.

`saseo` aims at that middle space: one existing rc file, one owned block, one
clear way to replace it, and one clear way to remove it later.

You can also make your main rc file source an extra optional rc file. That is
often fine, but the optional file becomes another bit of state to create,
disable, rename, or delete when the profile should stop applying. `saseo` keeps
that state inside the file you already use: one marked block that can be
replaced or removed.

This is most useful when another script already knows the desired content and
needs a safe way to claim a small part of the user's rc file.

## What Makes It Safer

- By default, the target is parsed as bash before and after the edit.
- Marker blocks must stay inside one shell syntax context. A block may live
  inside a heredoc body, multiline string, subshell, or top-level shell code, but
  it may not open or close that surrounding syntax.
- Delimiters are exact whole lines, so embedded text such as `echo "##SASEO^"`
  is not treated as a marker.
- `--dry-run` performs the same validation and prints the same success output
  without writing the file.
- No-op changes leave the target untouched.
- Failures use classified exit codes and stable stderr prefixes such as
  `error:`, `hint:`, and `try:`.

## Behavior

- The target file must already exist.
- `put` reads a replacement body from stdin, and stdin must not be empty.
- `mark` creates or replaces a marker block with no body lines and ignores
  stdin.
- `rm` ignores stdin.
- `show` prints the current marker block, including marker lines, and ignores
  stdin. If no block exists, it prints `0` and exits successfully.
- `--dry-run` validates and reports the same success output without writing the
  file.
- The default marker is `SASEO`.
- Marker names may use `a-z`, `A-Z`, `0-9`, `_`, `.`, `:`, and `-`, but may not
  start with `-`.
- Marker delimiters are exact whole lines; embedded text such as
  `echo "##SASEO^"` is not touched.
- By default, the target is parsed as bash before and after the edit.
- `show` is read-only and does not parse the target as shell.
- The marker block must stay within one shell syntax context. It may be inside a
  heredoc body, a multiline string, or a subshell, but it must not open or close
  that surrounding construct.
- Use `--plain-text` when editing non-shell files. Plain-text mode skips shell
  syntax checks, but still enforces marker names, target-file rules, stdin
  rules, and exact marker delimiter lines.
- zsh files are not officially supported. They are checked with the bash parser,
  so zsh-only syntax may be rejected.
- Options must appear before the single target file, and each option may appear
  at most once. Use `--` before the file only when the filename should stop
  option parsing.
- No-op changes leave the target file untouched.
- Symlink targets and CRLF target files are rejected.
- Failures print an `error:` line and may include `hint:` or `try:` guidance.

Successful `put` and `mark` print:

```text
START_LINE LINE_COUNT
```

Successful `rm` prints:

```text
START_LINE
```

Successful `show` prints:

```text
START_LINE LINE_COUNT
##MARKER^
...
##MARKER$
```

If no block exists, `show` prints:

```text
0
```

Line numbers are 1-based.
For `put` and `mark`, `LINE_COUNT` includes the start and end marker lines.
For `mark`, `LINE_COUNT` is `2`. For `show`, the block content output contains
exactly `LINE_COUNT` lines after the first line.

## Comparison

Related tools solve nearby problems. Check them instead when they match what
you need:

- Use `sed`, `awk`, or `perl` when raw text replacement is enough and you do not
  need shell syntax validation.
- Use [home-manager](https://github.com/nix-community/home-manager) when you
  want your home profile to be declarative and Nix-managed.
- Use [chezmoi](https://www.chezmoi.io/) or [yadm](https://yadm.io/) when you
  want a real dotfiles manager with a repo as the source of truth, templates,
  secrets, or per-machine state.
- Use [GNU Stow](https://www.gnu.org/software/stow/) or
  [rcm](https://thoughtbot.github.io/rcm/) when you want to expose many whole
  files into `$HOME`, especially through symlinks.
- Use [Dotbot](https://github.com/anishathalye/dotbot) when you want a
  repeatable bootstrap or install script for a dotfiles repo.

Use `saseo` when another script already knows the desired content and only needs
to claim one bash-aware marked section inside one existing rc file.

## Spec

Exact CLI forms, validation rules, and exit codes are documented in
[`saseo-spec.md`](saseo-spec.md).
