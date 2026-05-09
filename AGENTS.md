# AGENTS.md

This file is for people maintaining `saseo`.

Keep `README.md` and `AGENTS.md` separate:

- `README.md`: what users of the program need to know.
- `AGENTS.md`: what maintainers of the program need to know.

Avoid duplicating operational details between the two. If a detail is only
needed for development, keep it here.

## Development Commands

Enter the development shell:

```bash
nix develop
```

Run the program from source:

```bash
nix develop --command -- nu saseo.nu --help
```

Build the package:

```bash
nix build
```

Format Nushell files:

```bash
nix fmt
```

Or directly:

```bash
nix develop --command -- nufmt
```

Run tests directly:

```bash
nix develop --command -- nu test/saseo-test.nu
```

## Change Coordination

When changing any one of the spec, documentation, implementation, or tests,
check whether the others must change too.

Keep these in sync:

- `saseo.nu`
- `saseo-spec.md`
- `saseo.1.scd`
- `README.md`
- `test/**/*.nu`
- `package.nix`
- `flake.nix`

Behavioral changes should normally update the spec and tests in the same
change. User-visible changes should normally update `README.md`.

## Specification IDs

Behavioral spec sections in `saseo-spec.md` should have stable IDs in their
headings, such as `[SASEO-ADD-REPLACE]`.

Tests should point back to the behavior they cover with a top-of-file
`# Spec:` comment listing the relevant spec IDs. Keep this reference
one-directional: tests may reference the spec, but the spec should not list
test paths or implementation details.

When changing behavior, update affected spec IDs and test `# Spec:` comments
in the same change.

## Exit Code Policy

Do not introduce explicit `saseo` exit codes `1` or `2`.

Leave `1` for general, unknown, or unexpected failures that are not explicitly
classified by `saseo`. Leave `2` unused.

## Test Organization

The test runner is `test/saseo-test.nu`. It discovers and runs immediate
`test/*-test/*.test.nu` files in sorted order.

Each `*-test/` directory is a test item. Keep the Nushell test script and all
files needed by that item in the same directory. Do not add shared test helpers;
each test item should compute paths, prepare targets, and perform assertions
itself.

Test runs should create scratch files in the test item directory unless there
is a specific reason to use another location. Prefer scratch file names that
match `*.temp.txt`, which are ignored by `test/.gitignore`. If a test needs
some other intermediate file name, add a `.gitignore` to that test item
directory so every such intermediate artifact remains untracked. A scratch file
may be left behind after a failed test. Tests must reset scratch targets from
their local tracked inputs before relying on their contents.

## Error Output Test Conventions

Tests that cover `SASEO-ERROR-OUTPUT` should verify the stable stderr shape,
not the exact prose.

For classified failures:

- include `SASEO-ERROR-OUTPUT` in the test's top-of-file `# Spec:` comment,
- assert that stderr is not empty,
- assert that the first stderr line starts with `error: ` and has text after
  the prefix,
- assert that any later stderr line starts with `hint: ` or `try: ` and has
  text after the prefix.

Do not lock down exact error text, paths, suggested commands, the number of
stderr lines, or whether `hint:` or `try:` guidance is present for a particular
failure.

## Commit Message Style

Write commit messages as the `{}` part of these sentences:

```text
This commit will {}.
```

```text
This commit is {}.
```

Rules:

- Start with a lowercase letter.
- Do not end with a period.
- Do not use prefixes such as `feat:`, `fix:`, or `chore:`.

Examples:

```text
add readme
rename foo -> bar
```
