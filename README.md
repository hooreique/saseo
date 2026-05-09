# saseo

`saseo` adds and removes small rc snippets that should stick around for a while,
but not forever.

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
<pre lang="bash"><code>saseo ~/.bashrc &lt;&lt;'EOF'
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
existing file, and leaves everything outside that block alone.

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
saseo ~/.bashrc <<'EOF'
export FOO=foo
export BAR=bar
EOF
```

Use a named marker:

```bash
saseo --marker WORK ~/.bashrc <<< 'export AWS_PROFILE=work'
```

Remove a block:

```bash
saseo --rm --marker WORK ~/.bashrc
```

Preview a change without writing the file:

```bash
saseo --dry-run --marker WORK ~/.bashrc <<< 'export AWS_PROFILE=work'
saseo --dry-run --rm --marker WORK ~/.bashrc
```

Show help or version:

```bash
saseo --help
saseo --version
```

If installed, the manual is available with `man saseo`.

## Why

Dotfile managers can be too much.

`saseo` is for profiles with an awkward lifecycle: not a profile that should be
installed forever, and not a completely disposable temporary profile either. It
sits somewhere between those cases.

If you carry one machine between home and work, you probably know this space: a
few settings need to change, but they do not deserve a whole dotfiles system.

You can also make your main rc file source an extra optional rc file. That is
often fine, but the optional file becomes another bit of state to create,
disable, rename, or delete when the profile should stop applying. `saseo` keeps
that state inside the file you already use: one marked block that can be
replaced or removed.

Of course, this still works best when your shell history can find the exact
command again.

## Behavior

- The target file must already exist.
- Without `--rm`, stdin must not be empty.
- `--rm` ignores stdin.
- `--dry-run` validates and reports the same success output without writing the
  file.
- The default marker is `SASEO`.
- Marker names may use `a-z`, `A-Z`, `0-9`, `_`, `.`, `:`, and `-`, but may not
  start with `-`.
- Marker delimiters are exact whole lines; embedded text such as
  `echo "##SASEO^"` is not touched.
- No-op changes leave the target file untouched.
- Symlink targets and CRLF target files are rejected.
- Failures print an `error:` line and may include `hint:` or `try:` guidance.

Successful add or replace prints:

```text
START_LINE LINE_COUNT
```

Successful removal prints:

```text
START_LINE
```

Line numbers are 1-based.
For add or replace, `LINE_COUNT` includes the start and end marker lines.

## Comparison

Related tools solve nearby problems. Check them instead when they match what
you need:

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
to claim one marked section inside one existing file.

## Spec

Exact CLI forms, validation rules, and exit codes are documented in
[`saseo-spec.md`](saseo-spec.md).
