# saseo

`saseo` is a small tool for claiming one bash-aware marked block inside an existing shell rc file.

It is built for the moment when a script knows exactly what it wants to install, but should not take over the whole file. `saseo` lets you add, replace, show, or remove that one marked region without touching anything else.

Why use it:

- keep generated shell config in the file you already use
- replace one owned block without regex surgery
- validate shell boundaries by default
- remove the block later with the same tool

## Install

Try it once with Nix:

```bash
nix run github:hooreique/saseo -- --help
```

Install it into your Nix profile:

```bash
nix profile install github:hooreique/saseo
```

After installing, run `saseo --help`, `saseo put --help`, or `man saseo` for details.

## Quick Start

Add or replace a block:

```bash
saseo put ~/.bashrc <<'EOF'
export FOO=foo
export BAR=bar
EOF
```

Use a named block:

```bash
saseo put --marker WORK ~/.bashrc <<'EOF'
export AWS_PROFILE=work
EOF
```

Create an empty block:

```bash
saseo mark --marker WORK ~/.bashrc
```

Remove a block:

```bash
saseo rm --marker WORK ~/.bashrc
```

Inspect a block:

```bash
saseo show --marker WORK ~/.bashrc
```

Preview without writing:

```bash
saseo put --dry-run --marker WORK ~/.bashrc <<'EOF'
export AWS_PROFILE=work
EOF
```

Plain-text mode:

```bash
touch notes.txt
saseo put --plain-text notes.txt <<'EOF'
temporary note
EOF
```

## Notes

- `put`, `mark`, and `rm` accept `--dry-run`
- `show` is read-only and prints `0` when the block is absent
- `--plain-text` skips shell parsing for non-shell files
- the target file must already exist
- the default marker is `SASEO`
- zsh-only syntax is not officially supported

For the full behavior contract, see [saseo-spec.md](saseo-spec.md).
