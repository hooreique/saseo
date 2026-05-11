// Spec: SASEO-CMD, SASEO-MARKER-NAME, SASEO-MARKER-FORMAT,
// SASEO-ADD-REPLACE, SASEO-MARK, SASEO-APPEND-FORMATTING, SASEO-REMOVE,
// SASEO-SHOW, SASEO-PLAIN-TEXT, SASEO-SHELL-VALIDATION, SASEO-SHELL-BOUNDARY,
// SASEO-DRY-RUN, SASEO-TARGET-FILE, SASEO-INVALID-MARKER-STATE,
// SASEO-INVALID-OPTIONS, SASEO-ERROR-OUTPUT, SASEO-AFFECTED-LINES,
// SASEO-EXIT-CODES.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type commandResult struct {
	code   int
	stdout string
	stderr string
}

func TestCommandScenarios(t *testing.T) {
	tests := []struct {
		name              string
		fixture           string
		args              func(string) []string
		stdin             string
		wantCode          int
		wantStdout        string
		wantErrorFormat   bool
		wantFixture       string
		wantMetadataEqual bool
	}{
		{
			name:        "append managed block with default marker",
			fixture:     "append-default-test",
			args:        func(path string) []string { return []string{"put", path} },
			stdin:       "export X=foo\nexport Y=bar\n",
			wantStdout:  "3 4\n",
			wantFixture: "expected.txt",
		},
		{
			name:        "append to empty file",
			fixture:     "append-empty-test",
			args:        func(path string) []string { return []string{"put", path} },
			stdin:       "export N=1",
			wantStdout:  "1 3\n",
			wantFixture: "expected.txt",
		},
		{
			name:        "append after unterminated line",
			fixture:     "append-unterminated-test",
			args:        func(path string) []string { return []string{"put", path} },
			stdin:       "export N=1",
			wantStdout:  "2 3\n",
			wantFixture: "expected.txt",
		},
		{
			name:            "body marker delimiter",
			fixture:         "body-marker-delimiter-test",
			args:            func(path string) []string { return []string{"put", path} },
			stdin:           "before\n##SASEO^\nafter\n",
			wantCode:        exitInputValidation,
			wantErrorFormat: true,
			wantFixture:     "input.txt",
		},
		{
			name:            "crlf target",
			fixture:         "crlf-target-test",
			args:            func(path string) []string { return []string{"put", path} },
			stdin:           "export N=1",
			wantCode:        exitUnsupportedTarget,
			wantErrorFormat: true,
			wantFixture:     "input.txt",
		},
		{
			name:              "dry-run append managed block",
			fixture:           "dry-run-append-test",
			args:              func(path string) []string { return []string{"put", "--dry-run", path} },
			stdin:             "export X=foo\nexport Y=bar\n",
			wantStdout:        "3 4\n",
			wantFixture:       "input.txt",
			wantMetadataEqual: true,
		},
		{
			name:              "dry-run remove absent managed block",
			fixture:           "dry-run-remove-absent-test",
			args:              func(path string) []string { return []string{"rm", "--dry-run", "--marker", "SASEO", path} },
			stdin:             "ignored stdin\n",
			wantFixture:       "input.txt",
			wantMetadataEqual: true,
		},
		{
			name:              "dry-run remove managed block",
			fixture:           "dry-run-remove-test",
			args:              func(path string) []string { return []string{"rm", "--dry-run", "--marker", "SASEO", path} },
			stdin:             "ignored stdin\n",
			wantStdout:        "3\n",
			wantFixture:       "input.txt",
			wantMetadataEqual: true,
		},
		{
			name:              "dry-run replace managed block",
			fixture:           "dry-run-replace-test",
			args:              func(path string) []string { return []string{"put", "--dry-run", "--marker", "SASEO", path} },
			stdin:             "export X=baz\nexport Z=qux\n",
			wantStdout:        "3 4\n",
			wantFixture:       "input.txt",
			wantMetadataEqual: true,
		},
		{
			name:            "duplicate marker",
			fixture:         "duplicate-markers-test",
			args:            func(path string) []string { return []string{"rm", "--marker", "SASEO", path} },
			wantCode:        exitMarkerState,
			wantErrorFormat: true,
			wantFixture:     "input.txt",
		},
		{
			name:            "empty stdin add",
			fixture:         "empty-stdin-add-test",
			args:            func(path string) []string { return []string{"put", path} },
			wantCode:        exitEmptyStdin,
			wantErrorFormat: true,
			wantFixture:     "input.txt",
		},
		{
			name:            "end marker only",
			fixture:         "end-marker-only-test",
			args:            func(path string) []string { return []string{"rm", "--marker", "SASEO", path} },
			wantCode:        exitMarkerState,
			wantErrorFormat: true,
			wantFixture:     "input.txt",
		},
		{
			name:            "invalid marker",
			fixture:         "invalid-marker-test",
			args:            func(path string) []string { return []string{"put", "--marker", "BAD/MARKER", path} },
			stdin:           "export X=foo\nexport Y=bar\n",
			wantCode:        exitInputValidation,
			wantErrorFormat: true,
			wantFixture:     "input.txt",
		},
		{
			name:            "empty marker from equals form",
			fixture:         "invalid-marker-test",
			args:            func(path string) []string { return []string{"put", "--marker=", path} },
			stdin:           "export X=foo\nexport Y=bar\n",
			wantCode:        exitInputValidation,
			wantErrorFormat: true,
			wantFixture:     "input.txt",
		},
		{
			name:            "empty marker from separate value",
			fixture:         "invalid-marker-test",
			args:            func(path string) []string { return []string{"put", "--marker", "", path} },
			stdin:           "export X=foo\nexport Y=bar\n",
			wantCode:        exitInputValidation,
			wantErrorFormat: true,
			wantFixture:     "input.txt",
		},
		{
			name:            "leading dash marker",
			fixture:         "leading-dash-marker-test",
			args:            func(path string) []string { return []string{"put", "--marker", "-BAD", path} },
			stdin:           "export X=foo\nexport Y=bar\n",
			wantCode:        exitInputValidation,
			wantErrorFormat: true,
			wantFixture:     "input.txt",
		},
		{
			name:        "marker substrings are not delimiters",
			fixture:     "marker-substrings-test",
			args:        func(path string) []string { return []string{"put", path} },
			stdin:       "export N=1",
			wantStdout:  "3 3\n",
			wantFixture: "expected.txt",
		},
		{
			name:              "remove absent managed block",
			fixture:           "remove-absent-test",
			args:              func(path string) []string { return []string{"rm", "--marker", "SASEO", path} },
			stdin:             "ignored stdin\n",
			wantFixture:       "expected.txt",
			wantMetadataEqual: true,
		},
		{
			name:              "show absent managed block",
			fixture:           "remove-absent-test",
			args:              func(path string) []string { return []string{"show", "--marker", "SASEO", path} },
			stdin:             "ignored stdin\n",
			wantStdout:        "0\n",
			wantFixture:       "input.txt",
			wantMetadataEqual: true,
		},
		{
			name:        "remove managed block ignores stdin",
			fixture:     "remove-managed-test",
			args:        func(path string) []string { return []string{"rm", "--marker", "SASEO", path} },
			stdin:       "ignored stdin\n",
			wantStdout:  "3\n",
			wantFixture: "expected.txt",
		},
		{
			name:        "remove preserves suffix",
			fixture:     "remove-preserves-suffix-test",
			args:        func(path string) []string { return []string{"rm", path} },
			wantStdout:  "2\n",
			wantFixture: "expected.txt",
		},
		{
			name:        "replace existing managed block",
			fixture:     "replace-existing-test",
			args:        func(path string) []string { return []string{"put", "--marker", "SASEO", path} },
			stdin:       "export X=baz\nexport Z=qux\n",
			wantStdout:  "3 4\n",
			wantFixture: "expected.txt",
		},
		{
			name:              "show existing managed block",
			fixture:           "replace-existing-test",
			args:              func(path string) []string { return []string{"show", "--marker", "SASEO", path} },
			stdin:             "ignored stdin\n",
			wantStdout:        "3 4\n##SASEO^\nexport X=foo\nexport Y=bar\n##SASEO$\n",
			wantFixture:       "input.txt",
			wantMetadataEqual: true,
		},
		{
			name:              "replace identical",
			fixture:           "replace-identical-test",
			args:              func(path string) []string { return []string{"put", path} },
			stdin:             "same\n",
			wantStdout:        "2 3\n",
			wantFixture:       "input.txt",
			wantMetadataEqual: true,
		},
		{
			name:        "replace preserves suffix",
			fixture:     "replace-preserves-suffix-test",
			args:        func(path string) []string { return []string{"put", path} },
			stdin:       "export N=1",
			wantStdout:  "2 3\n",
			wantFixture: "expected.txt",
		},
		{
			name:            "reversed marker",
			fixture:         "reversed-markers-test",
			args:            func(path string) []string { return []string{"rm", "--marker", "SASEO", path} },
			wantCode:        exitMarkerState,
			wantErrorFormat: true,
			wantFixture:     "input.txt",
		},
		{
			name:            "shell boundary",
			fixture:         "shell-boundary-test",
			args:            func(path string) []string { return []string{"put", path} },
			stdin:           "echo new\n",
			wantCode:        exitBoundary,
			wantErrorFormat: true,
			wantFixture:     "input.txt",
		},
		{
			name:        "shell heredoc marker block",
			fixture:     "shell-heredoc-test",
			args:        func(path string) []string { return []string{"put", path} },
			stdin:       "new ) body\n",
			wantStdout:  "2 3\n",
			wantFixture: "expected.txt",
		},
		{
			name:            "shell syntax",
			fixture:         "shell-syntax-test",
			args:            func(path string) []string { return []string{"put", path} },
			stdin:           "echo new\n",
			wantCode:        exitSyntax,
			wantErrorFormat: true,
			wantFixture:     "input.txt",
		},
		{
			name:            "start marker only",
			fixture:         "start-marker-only-test",
			args:            func(path string) []string { return []string{"rm", "--marker", "SASEO", path} },
			wantCode:        exitMarkerState,
			wantErrorFormat: true,
			wantFixture:     "input.txt",
		},
		{
			name:            "unknown option",
			fixture:         "unknown-option-test",
			args:            func(path string) []string { return []string{"put", "--bogus", path} },
			wantCode:        exitUsage,
			wantErrorFormat: true,
			wantFixture:     "input.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			work := copyInputToTemp(t, tt.fixture, "target.txt")
			var before fileMetadata
			if tt.wantMetadataEqual {
				before = setStableMetadata(t, work)
			}
			result := runCommand(tt.args(work), tt.stdin)
			assertCommandResult(t, result, tt.wantCode, tt.wantStdout)
			if tt.wantErrorFormat {
				assertErrorOutputFormat(t, result.stderr)
			}
			if tt.wantFixture != "" {
				assertFileEqualsFixture(t, work, tt.fixture, tt.wantFixture)
			}
			if tt.wantMetadataEqual {
				assertSameMetadata(t, work, before)
			}
		})
	}
}

func TestMarkerDigits(t *testing.T) {
	work := copyInputToTemp(t, "marker-digits-test", "target.txt")
	result := runCommand([]string{"put", "--marker", "NODE22", work}, "export N=1")
	assertCommandResult(t, result, 0, "3 3\n")
	assertFileContains(t, work, "##NODE22^\n")
}

func TestMarkCommand(t *testing.T) {
	t.Run("allows empty stdin", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "target.txt")
		if err := writeTestFile(path, "export BASE=1\n"); err != nil {
			t.Fatal(err)
		}

		result := runCommand([]string{"mark", path}, "")
		assertCommandResult(t, result, 0, "2 2\n")
		assertFileText(t, path, "export BASE=1\n##SASEO^\n##SASEO$\n")
	})

	t.Run("append empty block and ignore stdin", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "target.txt")
		if err := writeTestFile(path, "# profile init\nexport BASE=1\n"); err != nil {
			t.Fatal(err)
		}

		result := runCommand([]string{"mark", path}, "ignored stdin\n")
		assertCommandResult(t, result, 0, "3 2\n")
		assertFileText(t, path, "# profile init\nexport BASE=1\n##SASEO^\n##SASEO$\n")
	})

	t.Run("dry-run leaves target unchanged", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "target.txt")
		initial := "# profile init\nexport BASE=1\n"
		if err := writeTestFile(path, initial); err != nil {
			t.Fatal(err)
		}
		before := setStableMetadata(t, path)

		result := runCommand([]string{"mark", "--dry-run", path}, "ignored stdin\n")
		assertCommandResult(t, result, 0, "3 2\n")
		assertFileText(t, path, initial)
		assertSameMetadata(t, path, before)
	})

	t.Run("replace existing block with empty content", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "target.txt")
		if err := writeTestFile(path, "# profile init\nexport BASE=1\n##SASEO^\nexport X=foo\nexport Y=bar\n##SASEO$\n"); err != nil {
			t.Fatal(err)
		}

		result := runCommand([]string{"mark", "--marker", "SASEO", path}, "ignored stdin\n")
		assertCommandResult(t, result, 0, "3 2\n")
		assertFileText(t, path, "# profile init\nexport BASE=1\n##SASEO^\n##SASEO$\n")
	})
}

func TestShowCommand(t *testing.T) {
	t.Run("custom marker", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "target.txt")
		initial := "export BASE=1\n##WORK^\nexport AWS_PROFILE=work\n##WORK$\n"
		if err := writeTestFile(path, initial); err != nil {
			t.Fatal(err)
		}
		before := setStableMetadata(t, path)

		result := runCommand([]string{"show", "--marker=WORK", path}, "")
		assertCommandResult(t, result, 0, "2 3\n##WORK^\nexport AWS_PROFILE=work\n##WORK$\n")
		assertFileText(t, path, initial)
		assertSameMetadata(t, path, before)
	})

	t.Run("does not require valid shell syntax", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "target.txt")
		initial := "not shell (((\n##SASEO^\nplain body\n##SASEO$\n"
		if err := writeTestFile(path, initial); err != nil {
			t.Fatal(err)
		}

		result := runCommand([]string{"show", path}, "")
		assertCommandResult(t, result, 0, "2 3\n##SASEO^\nplain body\n##SASEO$\n")
		assertFileText(t, path, initial)
	})
}

func TestDashFilenameAfterEndOfOptions(t *testing.T) {
	dir := t.TempDir()
	work := filepath.Join(dir, "-weird-file")
	copyFixtureToPath(t, "dash-filename-test", "input.txt", work)
	t.Chdir(dir)

	result := runCommand([]string{"put", "--marker", "DASH", "--", "-weird-file"}, "export N=1")
	assertCommandResult(t, result, 0, "3 3\n")
	assertFileContains(t, work, "##DASH^\n")
}

func TestPlainTextModeScenario(t *testing.T) {
	work := copyInputToTemp(t, "plain-text-mode-test", "target.txt")

	shellResult := runCommand([]string{"put", work}, "plain body\n")
	assertCommandResult(t, shellResult, exitSyntax, "")
	assertErrorOutputFormat(t, shellResult.stderr)
	assertFileEqualsFixture(t, work, "plain-text-mode-test", "input.txt")

	result := runCommand([]string{"put", "--plain-text", work}, "plain body\n")
	assertCommandResult(t, result, 0, "2 3\n")
	assertFileEqualsFixture(t, work, "plain-text-mode-test", "expected.txt")
}

func TestMissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.txt")
	result := runCommand([]string{"rm", "--marker", "SASEO", missing}, "")
	assertCommandResult(t, result, exitMissingTarget, "")
	assertErrorOutputFormat(t, result.stderr)
}

func TestSymlinkTarget(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "link")
	copyFixtureToPath(t, "symlink-target-test", "input.txt", target)
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	result := runCommand([]string{"put", link}, "export N=1")
	assertCommandResult(t, result, exitUnsupportedTarget, "")
	assertErrorOutputFormat(t, result.stderr)
	assertFileEqualsFixture(t, target, "symlink-target-test", "input.txt")
}

func TestArgumentFailures(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "missing marker value",
			args: []string{"put", "--marker"},
		},
		{
			name: "missing file argument",
			args: []string{"put"},
		},
		{
			name: "root help with extra argument",
			args: []string{"--help", "extra"},
		},
		{
			name: "version with extra argument",
			args: []string{"--version", "extra"},
		},
		{
			name: "subcommand help with extra argument",
			args: []string{"put", "--help", "extra"},
		},
		{
			name: "subcommand help after option",
			args: []string{"put", "--dry-run", "--help"},
		},
		{
			name: "show dry run option",
			args: []string{"show", "--dry-run", "target.txt"},
		},
		{
			name: "show plain text option",
			args: []string{"show", "--plain-text", "target.txt"},
		},
		{
			name: "duplicate dry run option",
			args: []string{"put", "--dry-run", "--dry-run", "target.txt"},
		},
		{
			name: "duplicate plain text option",
			args: []string{"put", "--plain-text", "--plain-text", "target.txt"},
		},
		{
			name: "duplicate marker option",
			args: []string{"put", "--marker", "ONE", "--marker=TWO", "target.txt"},
		},
		{
			name: "option after file",
			args: []string{"put", "target.txt", "--dry-run"},
		},
		{
			name: "argument after file",
			args: []string{"put", "target.txt", "extra"},
		},
		{
			name: "end of options without file",
			args: []string{"put", "--"},
		},
		{
			name: "extra argument after end of options file",
			args: []string{"put", "--", "target.txt", "extra"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCommand(tt.args, "")
			assertCommandResult(t, result, exitUsage, "")
			assertErrorOutputFormat(t, result.stderr)
		})
	}
}

func TestHelpAndVersion(t *testing.T) {
	noSubcommand := runCommand(nil, "")
	assertExitCode(t, noSubcommand, 0)
	assertStringContains(t, noSubcommand.stdout, "Usage:")
	if noSubcommand.stderr != "" {
		t.Fatalf("stderr = %q, want empty", noSubcommand.stderr)
	}

	help := runCommand([]string{"--help"}, "")
	assertExitCode(t, help, 0)
	assertStringContains(t, help.stdout, "saseo manages one marked block in an existing rc file.")
	for _, text := range []string{
		"Cheat sheet:",
		"saseo put ~/.bashrc <<< 'export FOO=bar'",
		"saseo show ~/.bashrc",
		"saseo rm ~/.bashrc",
		"Usage:",
		"saseo <put|mark|rm|show> [options] [--] FILE",
		"  put    add or replace from stdin",
		"  mark   create an empty block",
		"  show   print the block",
		"  --dry-run",
		"  --plain-text",
		"  man saseo",
	} {
		assertStringContains(t, help.stdout, text)
	}
	for _, text := range []string{
		"\x1b[",
		"  saseo -h\n",
		"  saseo -V\n",
		"saseo -- -",
		"saseo -- --",
		"saseo --rm",
	} {
		assertStringOmits(t, help.stdout, text)
	}

	version := runCommand([]string{"--version"}, "")
	assertCommandResult(t, version, 0, strings.TrimSpace(versionText)+"\n")
	if version.stderr != "" {
		t.Fatalf("version stderr = %q, want empty", version.stderr)
	}

	markHelp := runCommand([]string{"mark", "--help"}, "")
	assertExitCode(t, markHelp, 0)
	assertStringContains(t, markHelp.stdout, "saseo mark [--dry-run] [--marker NAME] [--plain-text] [--] FILE")
	assertStringContains(t, markHelp.stdout, "Create an empty block.")
	assertStringContains(t, markHelp.stdout, "saseo mark --marker WORK ~/.bashrc")
	if markHelp.stderr != "" {
		t.Fatalf("mark help stderr = %q, want empty", markHelp.stderr)
	}

	showHelp := runCommand([]string{"show", "--help"}, "")
	assertExitCode(t, showHelp, 0)
	assertStringContains(t, showHelp.stdout, "saseo show [--marker NAME] [--] FILE")
	assertStringContains(t, showHelp.stdout, "Print the block.")
	assertStringContains(t, showHelp.stdout, "saseo show --marker WORK ~/.bashrc")
	assertStringOmits(t, showHelp.stdout, "--dry-run")
	assertStringOmits(t, showHelp.stdout, "--plain-text")
	if showHelp.stderr != "" {
		t.Fatalf("show help stderr = %q, want empty", showHelp.stderr)
	}
}

func TestShellBoundaryValidation(t *testing.T) {
	tests := []struct {
		name string
		text string
		code int
	}{
		{
			name: "top level",
			text: "##SASEO^\necho ok\n##SASEO$\n",
		},
		{
			name: "subshell",
			text: "(\n##SASEO^\necho ok\n##SASEO$\n)\n",
		},
		{
			name: "heredoc body",
			text: "cat <<'EOF'\n##SASEO^\nnot shell ) text\n##SASEO$\nEOF\n",
		},
		{
			name: "heredoc body with placeholder-looking delimiter",
			text: "cat <<'__SASEO_PLACEHOLDER_0__'\n##SASEO^\nnot shell ) text\n##SASEO$\n__SASEO_PLACEHOLDER_0__\n",
		},
		{
			name: "single quoted multiline string",
			text: "value='before\n##SASEO^\nnot shell ) text\n##SASEO$\nafter'\n",
		},
		{
			name: "snippet opens subshell",
			text: "##SASEO^\n(\necho bad\n##SASEO$\n)\n",
			code: exitBoundary,
		},
		{
			name: "snippet closes subshell",
			text: "(\n##SASEO^\necho bad\n)\n##SASEO$\n",
			code: exitBoundary,
		},
		{
			name: "snippet closes and reopens quote",
			text: "value='before\n##SASEO^\n'\nvalue='\n##SASEO$\nafter'\n",
			code: exitBoundary,
		},
		{
			name: "snippet closes heredoc before marker end",
			text: "cat <<'EOF'\n##SASEO^\nEOF\n##SASEO$\n",
			code: exitBoundary,
		},
		{
			name: "snippet crosses if then else",
			text: "if true; then\n##SASEO^\necho then\nelse\n##SASEO$\necho else\nfi\n",
			code: exitBoundary,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, markerErr := markerStateFor(tt.text, defaultMarker)
			if markerErr != nil {
				t.Fatalf("markerStateFor returned error: %v", markerErr)
			}
			err := validateShellTarget(tt.text, defaultMarker, state)
			if tt.code == 0 {
				if err != nil {
					t.Fatalf("validateShellTarget returned error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateShellTarget returned nil, want code %d", tt.code)
			}
			if err.code != tt.code {
				t.Fatalf("validateShellTarget code = %d, want %d: %v", err.code, tt.code, err)
			}
		})
	}
}

func TestFinalValidationRejectsBodyThatBreaksContext(t *testing.T) {
	old := "(\n##SASEO^\necho old\n##SASEO$\n)\n"
	state, markerErr := markerStateFor(old, defaultMarker)
	if markerErr != nil {
		t.Fatalf("markerStateFor returned error: %v", markerErr)
	}
	result := replaceOrAppend(old, defaultMarker, "true\n)\n(\ntrue\n", state)
	newState, markerErr := markerStateFor(result.text, defaultMarker)
	if markerErr != nil {
		t.Fatalf("markerStateFor returned error for result: %v", markerErr)
	}
	err := validateShellTarget(result.text, defaultMarker, newState)
	if err == nil {
		t.Fatalf("validateShellTarget returned nil for %q", result.text)
	}
	if err.code != exitBoundary {
		t.Fatalf("validateShellTarget code = %d, want %d: %v", err.code, exitBoundary, err)
	}
}

func TestPlainTextAllowsInvalidShell(t *testing.T) {
	stdout := &strings.Builder{}
	dir := t.TempDir()
	path := dir + "/target.txt"
	if err := writeTestFile(path, "not shell (((\n"); err != nil {
		t.Fatal(err)
	}
	code := runMain([]string{"put", "--plain-text", path}, strings.NewReader("plain body\n"), stdout, &strings.Builder{})
	if code != 0 {
		t.Fatalf("runMain code = %d, want 0", code)
	}
	if stdout.String() != "2 3\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), "2 3\n")
	}
}

func writeTestFile(path, text string) error {
	return os.WriteFile(path, []byte(text), 0o644)
}

func runCommand(args []string, stdin string) commandResult {
	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	code := runMain(args, strings.NewReader(stdin), stdout, stderr)
	return commandResult{
		code:   code,
		stdout: stdout.String(),
		stderr: stderr.String(),
	}
}

func assertCommandResult(t *testing.T, result commandResult, wantCode int, wantStdout string) {
	t.Helper()
	assertExitCode(t, result, wantCode)
	if result.stdout != wantStdout {
		t.Fatalf("stdout = %q, want %q", result.stdout, wantStdout)
	}
	if wantCode != 0 && result.stdout != "" {
		t.Fatalf("stdout = %q, want empty on failure", result.stdout)
	}
}

func assertExitCode(t *testing.T, result commandResult, wantCode int) {
	t.Helper()
	if result.code != wantCode {
		t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s", result.code, wantCode, result.stdout, result.stderr)
	}
}

func assertErrorOutputFormat(t *testing.T, stderr string) {
	t.Helper()
	if stderr == "" {
		t.Fatal("stderr is empty")
	}
	lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
	if !hasPrefixedContent(lines[0], "error: ") {
		t.Fatalf("first stderr line = %q, want error prefix with content\nstderr:\n%s", lines[0], stderr)
	}
	for _, line := range lines[1:] {
		if !hasPrefixedContent(line, "hint: ") && !hasPrefixedContent(line, "try: ") {
			t.Fatalf("later stderr line = %q, want hint or try prefix with content\nstderr:\n%s", line, stderr)
		}
	}
}

func hasPrefixedContent(line, prefix string) bool {
	return strings.HasPrefix(line, prefix) && len(line) > len(prefix)
}

func fixturePath(fixture, name string) string {
	stem := strings.TrimSuffix(name, ".txt")
	return filepath.Join("test", fixture+"-"+stem+".txt")
}

func copyInputToTemp(t *testing.T, fixture, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	copyFixtureToPath(t, fixture, "input.txt", path)
	return path
}

func copyFixtureToPath(t *testing.T, fixture, name, path string) {
	t.Helper()
	content := readFile(t, fixturePath(fixture, name))
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertFileEqualsFixture(t *testing.T, path, fixture, name string) {
	t.Helper()
	actual := readFile(t, path)
	expected := readFile(t, fixturePath(fixture, name))
	if string(actual) != string(expected) {
		t.Fatalf("%s contents differ from %s\nexpected:\n%s\nactual:\n%s", path, fixturePath(fixture, name), expected, actual)
	}
}

func assertFileContains(t *testing.T, path, substring string) {
	t.Helper()
	actual := string(readFile(t, path))
	assertStringContains(t, actual, substring)
}

func assertFileText(t *testing.T, path, expected string) {
	t.Helper()
	actual := string(readFile(t, path))
	if actual != expected {
		t.Fatalf("%s contents differ\nexpected:\n%s\nactual:\n%s", path, expected, actual)
	}
}

func assertStringContains(t *testing.T, text, substring string) {
	t.Helper()
	if !strings.Contains(text, substring) {
		t.Fatalf("missing substring %q in:\n%s", substring, text)
	}
}

func assertStringOmits(t *testing.T, text, substring string) {
	t.Helper()
	if strings.Contains(text, substring) {
		t.Fatalf("unexpected substring %q in:\n%s", substring, text)
	}
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}

type fileMetadata struct {
	info os.FileInfo
}

func setStableMetadata(t *testing.T, path string) fileMetadata {
	t.Helper()
	when := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(path, when, when); err != nil {
		t.Fatal(err)
	}
	return statFile(t, path)
}

func statFile(t *testing.T, path string) fileMetadata {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return fileMetadata{info: info}
}

func assertSameMetadata(t *testing.T, path string, before fileMetadata) {
	t.Helper()
	after := statFile(t, path)
	if !os.SameFile(before.info, after.info) {
		t.Fatalf("%s was replaced", path)
	}
	if before.info.Size() != after.info.Size() {
		t.Fatalf("%s size = %d, want %d", path, after.info.Size(), before.info.Size())
	}
	if !before.info.ModTime().Equal(after.info.ModTime()) {
		t.Fatalf("%s mtime = %s, want %s", path, after.info.ModTime(), before.info.ModTime())
	}
}
