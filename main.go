package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

const (
	exitUsage             = 3
	exitEmptyStdin        = 4
	exitMissingTarget     = 5
	exitInputValidation   = 6
	exitMarkerState       = 7
	exitUnsupportedTarget = 8
	exitSyntax            = 9
	exitBoundary          = 10
)

const defaultMarker = "SASEO"

//go:embed VERSION
var versionText string

var markerPattern = regexp.MustCompile(`^[A-Za-z0-9_.:][A-Za-z0-9_.:-]*$`)

type userError struct {
	code    int
	message string
	hint    string
	try     string
}

func (e *userError) Error() string { return e.message }

func fail(code int, message string) *userError {
	return &userError{code: code, message: message}
}

func (e *userError) withHint(hint string) *userError {
	e.hint = hint
	return e
}

func (e *userError) withTry(try string) *userError {
	e.try = try
	return e
}

type command string

const (
	commandPut  command = "put"
	commandMark command = "mark"
	commandRM   command = "rm"
	commandShow command = "show"
)

type options struct {
	command   command
	dryRun    bool
	marker    string
	plainText bool
	file      string
}

func main() {
	code := runMain(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	os.Exit(code)
}

func runMain(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	opts, handled, err := parseArgs(args, stdout)
	if handled {
		return 0
	}
	if err != nil {
		writeError(stderr, err)
		return err.code
	}
	if err := run(opts, stdin, stdout); err != nil {
		writeError(stderr, err)
		return err.code
	}
	return 0
}

func writeError(w io.Writer, err *userError) {
	fmt.Fprintf(w, "error: %s\n", err.message)
	if err.hint != "" {
		fmt.Fprintf(w, "hint: %s\n", err.hint)
	}
	if err.try != "" {
		fmt.Fprintf(w, "try: %s\n", err.try)
	}
}

func parseArgs(args []string, stdout io.Writer) (options, bool, *userError) {
	if len(args) == 0 {
		printRootHelp(stdout)
		return options{}, true, nil
	}
	switch args[0] {
	case "--help":
		if len(args) != 1 {
			return options{}, false, fail(exitUsage, fmt.Sprintf("unexpected argument after --help: %s", args[1])).
				withHint("The root --help form does not take additional arguments.").
				withTry("saseo --help")
		}
		printRootHelp(stdout)
		return options{}, true, nil
	case "--version":
		if len(args) != 1 {
			return options{}, false, fail(exitUsage, fmt.Sprintf("unexpected argument after --version: %s", args[1])).
				withHint("The --version form does not take additional arguments.").
				withTry("saseo --version")
		}
		fmt.Fprintln(stdout, strings.TrimSpace(versionText))
		return options{}, true, nil
	case "put":
		return parseSubcommand(commandPut, args[1:], stdout)
	case "mark":
		return parseSubcommand(commandMark, args[1:], stdout)
	case "rm":
		return parseSubcommand(commandRM, args[1:], stdout)
	case "show":
		return parseSubcommand(commandShow, args[1:], stdout)
	default:
		return options{}, false, fail(exitUsage, fmt.Sprintf("unknown command or option: %s", args[0])).
			withHint("Use one of: put, mark, rm, show, --help, --version.").
			withTry("saseo --help")
	}
}

func parseSubcommand(cmd command, args []string, stdout io.Writer) (options, bool, *userError) {
	opts := options{command: cmd, marker: defaultMarker}
	fileSeen := false
	markerSeen := false
	for idx := 0; idx < len(args); idx++ {
		arg := args[idx]
		if fileSeen {
			return options{}, false, fail(exitUsage, fmt.Sprintf("unexpected argument after file: %s", arg)).
				withHint("Pass options before FILE, and pass exactly one target file.").
				withTry(commandUsageTry(cmd))
		}
		switch {
		case arg == "--help":
			if len(args) != 1 {
				return options{}, false, fail(exitUsage, "--help cannot be combined with other arguments").
					withHint(fmt.Sprintf("Use either 'saseo %s --help' or run the %s command with a file.", cmd, cmd)).
					withTry(fmt.Sprintf("saseo %s --help", cmd))
			}
			printSubcommandHelp(stdout, cmd)
			return options{}, true, nil
		case arg == "--dry-run":
			if !commandAcceptsDryRun(cmd) {
				return options{}, false, unsupportedOptionError(cmd, "--dry-run")
			}
			if opts.dryRun {
				return options{}, false, duplicateOptionError(cmd, "--dry-run")
			}
			opts.dryRun = true
		case arg == "--plain-text":
			if !commandAcceptsPlainText(cmd) {
				return options{}, false, unsupportedOptionError(cmd, "--plain-text")
			}
			if opts.plainText {
				return options{}, false, duplicateOptionError(cmd, "--plain-text")
			}
			opts.plainText = true
		case arg == "--marker":
			idx++
			if idx >= len(args) {
				return options{}, false, fail(exitUsage, "--marker requires a value").
					withHint("Pass a marker name after --marker.").
					withTry(fmt.Sprintf("saseo %s --marker WORK FILE", cmd))
			}
			if markerSeen {
				return options{}, false, duplicateOptionError(cmd, "--marker")
			}
			opts.marker = args[idx]
			markerSeen = true
		case strings.HasPrefix(arg, "--marker="):
			if markerSeen {
				return options{}, false, duplicateOptionError(cmd, "--marker")
			}
			opts.marker = strings.TrimPrefix(arg, "--marker=")
			markerSeen = true
		case arg == "--":
			idx++
			if idx >= len(args) {
				return options{}, false, fail(exitUsage, "missing file argument after --").
					withHint("-- stops option parsing and must be followed by exactly one target file.").
					withTry(fmt.Sprintf("saseo %s -- FILE", cmd))
			}
			opts.file = args[idx]
			fileSeen = true
			if idx+1 < len(args) {
				return options{}, false, fail(exitUsage, fmt.Sprintf("unexpected argument after file: %s", args[idx+1])).
					withHint("-- stops option parsing; everything after the single FILE is extra.").
					withTry(fmt.Sprintf("saseo %s -- FILE", cmd))
			}
		case strings.HasPrefix(arg, "-"):
			return options{}, false, fail(exitUsage, fmt.Sprintf("unknown option: %s", arg)).
				withHint("Use -- before FILE if the filename starts with -.").
				withTry(fmt.Sprintf("saseo %s --help", cmd))
		default:
			opts.file = arg
			fileSeen = true
		}
	}
	if !fileSeen {
		return options{}, false, fail(exitUsage, "missing file argument").
			withHint(missingFileHint(cmd)).
			withTry(fmt.Sprintf("saseo %s --help", cmd))
	}
	return opts, false, nil
}

func commandUsageTry(cmd command) string {
	if cmd == commandShow {
		return "saseo show [--marker WORK] -- FILE"
	}
	return fmt.Sprintf("saseo %s [--dry-run] [--marker WORK] [--plain-text] -- FILE", cmd)
}

func missingFileHint(cmd command) string {
	if cmd == commandShow {
		return "Pass the target file to inspect."
	}
	return "Pass the target file to update."
}

func commandAcceptsDryRun(cmd command) bool {
	return cmd != commandShow
}

func commandAcceptsPlainText(cmd command) bool {
	return cmd != commandShow
}

func unsupportedOptionError(cmd command, option string) *userError {
	return fail(exitUsage, fmt.Sprintf("option not supported for %s: %s", cmd, option)).
		withHint(fmt.Sprintf("Check the %s command usage.", cmd)).
		withTry(fmt.Sprintf("saseo %s --help", cmd))
}

func duplicateOptionError(cmd command, option string) *userError {
	return fail(exitUsage, fmt.Sprintf("duplicate option: %s", option)).
		withHint("Pass each option at most once.").
		withTry(fmt.Sprintf("saseo %s --help", cmd))
}

func printRootHelp(w io.Writer) {
	fmt.Fprint(w, `saseo manages one marked block in an existing rc file.

Cheat sheet:
  saseo put ~/.bashrc <<< 'export FOO=bar'
  saseo show ~/.bashrc
  saseo rm ~/.bashrc

Usage:
  saseo <put|mark|rm|show> [options] [--] FILE
  saseo --version

Commands:
  put    add or replace from stdin
  mark   create an empty block
  show   print the block
  rm     remove the block

Common options:
  --marker NAME   use a named block
  --dry-run       preview put, mark, or rm
  --plain-text    skip shell parsing

More:
  man saseo
`)
}

func printSubcommandHelp(w io.Writer, cmd command) {
	switch cmd {
	case commandPut:
		fmt.Fprint(w, `Usage:
  saseo put [--dry-run] [--marker NAME] [--plain-text] [--] FILE

Add or replace a block from stdin.

Examples:
  saseo put ~/.bashrc <<< 'export FOO=bar'
  saseo put --marker WORK ~/.bashrc < work.env
  saseo put --dry-run ~/.bashrc < block.sh

More:
  man saseo
`)
	case commandMark:
		fmt.Fprint(w, `Usage:
  saseo mark [--dry-run] [--marker NAME] [--plain-text] [--] FILE

Create an empty block.

Examples:
  saseo mark ~/.bashrc
  saseo mark --marker WORK ~/.bashrc
  saseo mark --dry-run ~/.bashrc

More:
  man saseo
`)
	case commandRM:
		fmt.Fprint(w, `Usage:
  saseo rm [--dry-run] [--marker NAME] [--plain-text] [--] FILE

Remove a block.

Examples:
  saseo rm ~/.bashrc
  saseo rm --marker WORK ~/.bashrc
  saseo rm --dry-run ~/.bashrc

More:
  man saseo
`)
	case commandShow:
		fmt.Fprint(w, `Usage:
  saseo show [--marker NAME] [--] FILE

Print the block.

Examples:
  saseo show ~/.bashrc
  saseo show --marker WORK ~/.bashrc

More:
  man saseo
`)
	}
}

func run(opts options, stdin io.Reader, stdout io.Writer) *userError {
	if !markerPattern.MatchString(opts.marker) {
		return fail(exitInputValidation, "invalid marker").
			withHint("Use a-z, A-Z, 0-9, '_', '.', ':', and '-' after the first character.").
			withTry(fmt.Sprintf("saseo %s --marker WORK -- %s", opts.command, shellQuote(opts.file)))
	}

	path, err := filepath.Abs(opts.file)
	if err != nil {
		return fail(exitUsage, fmt.Sprintf("invalid file path: %s", opts.file))
	}
	info, statErr := os.Lstat(path)
	if errors.Is(statErr, os.ErrNotExist) {
		return fail(exitMissingTarget, fmt.Sprintf("target file does not exist: %s", path)).
			withHint("Create the target file first, then run saseo again.").
			withTry(fmt.Sprintf("touch %s", shellQuote(path)))
	}
	if statErr != nil {
		return fail(exitUnsupportedTarget, fmt.Sprintf("cannot inspect target file: %s", path))
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fail(exitUnsupportedTarget, fmt.Sprintf("target file must not be a symlink: %s", path)).
			withHint("Run saseo on the real target file instead of the symlink.").
			withTry(fmt.Sprintf("realpath %s", shellQuote(path)))
	}
	if !info.Mode().IsRegular() {
		return fail(exitUnsupportedTarget, fmt.Sprintf("target file must be a regular file: %s", path))
	}

	oldBytes, readErr := os.ReadFile(path) // #nosec G304 -- saseo intentionally edits the user-specified target file.
	if readErr != nil {
		return fail(exitUnsupportedTarget, fmt.Sprintf("cannot read target file: %s", path))
	}
	old := string(oldBytes)
	if strings.Contains(old, "\r\n") {
		return fail(exitUnsupportedTarget, fmt.Sprintf("target file uses unsupported CRLF line endings: %s", path))
	}

	state, markerErr := markerStateFor(old, opts.marker)
	if opts.command == commandShow {
		if markerErr != nil {
			return markerErr
		}
		printBlock(stdout, old, state)
		return nil
	}
	if !opts.plainText {
		if shellErr := validateShellTarget(old, opts.marker, state); shellErr != nil {
			return shellErr
		}
	}
	if markerErr != nil {
		return markerErr
	}

	var result editResult
	switch opts.command {
	case commandRM:
		result = removeBlock(old, opts.marker, state)
	case commandMark:
		result = replaceOrAppend(old, opts.marker, "", state)
	case commandPut:
		bodyBytes, readErr := io.ReadAll(stdin)
		if readErr != nil {
			return fail(exitInputValidation, "failed to read stdin")
		}
		body := string(bodyBytes)
		if body == "" {
			return fail(exitEmptyStdin, "stdin must not be empty when adding or replacing a block").
				withHint("Pipe or redirect the block body into saseo.").
				withTry(fmt.Sprintf("printf 'export FOO=foo\\n' | saseo put -- %s", shellQuote(opts.file)))
		}
		if err := validateBodyMarkers(opts.marker, body); err != nil {
			return err
		}
		result = replaceOrAppend(old, opts.marker, body, state)
	}

	if !opts.plainText {
		newState, markerErr := markerStateFor(result.text, opts.marker)
		if markerErr != nil {
			return markerErr
		}
		if shellErr := validateShellTarget(result.text, opts.marker, newState); shellErr != nil {
			return shellErr
		}
	}

	if !opts.dryRun && result.text != old {
		if err := atomicSave(path, result.text, info.Mode().Perm()); err != nil {
			return fail(exitUnsupportedTarget, fmt.Sprintf("cannot write target file: %s", path))
		}
	}

	if opts.command == commandRM {
		if result.startLine != 0 {
			fmt.Fprintln(stdout, result.startLine)
		}
	} else {
		fmt.Fprintf(stdout, "%d %d\n", result.startLine, result.affectedCount)
	}
	return nil
}

func printBlock(stdout io.Writer, text string, state markerState) {
	if !state.exists {
		fmt.Fprintln(stdout, "0")
		return
	}
	lines := lineTexts(text)
	lineCount := state.endIndex - state.startIndex + 1
	fmt.Fprintf(stdout, "%d %d\n", state.startIndex+1, lineCount)
	for _, line := range lines[state.startIndex : state.endIndex+1] {
		fmt.Fprintln(stdout, line)
	}
}

type line struct {
	start int
	text  string
}

func scanLines(text string) []line {
	lines := []line{}
	for start := 0; start < len(text); {
		rest := text[start:]
		nl := strings.IndexByte(rest, '\n')
		if nl < 0 {
			lines = append(lines, line{start: start, text: rest})
			break
		}
		lines = append(lines, line{start: start, text: rest[:nl]})
		start += nl + 1
	}
	return lines
}

func lineTexts(text string) []string {
	lines := scanLines(text)
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = line.text
	}
	return out
}

func linesToText(lines []string, finalNewline bool) string {
	if len(lines) == 0 {
		return ""
	}
	text := strings.Join(lines, "\n")
	if finalNewline {
		text += "\n"
	}
	return text
}

func normalizeLF(text string) string {
	return strings.ReplaceAll(text, "\r\n", "\n")
}

func ensureFinalNewline(text string) string {
	if text == "" || strings.HasSuffix(text, "\n") {
		return text
	}
	return text + "\n"
}

type markerState struct {
	start       string
	end         string
	startIndex  int
	endIndex    int
	startOffset int
	endOffset   int
	exists      bool
}

func markerStateFor(text, marker string) (markerState, *userError) {
	start := "##" + marker + "^"
	end := "##" + marker + "$"
	lines := scanLines(text)
	state := markerState{start: start, end: end, startIndex: -1, endIndex: -1, startOffset: -1, endOffset: -1}
	startCount := 0
	endCount := 0
	for idx, line := range lines {
		if line.text == start {
			startCount++
			state.startIndex = idx
			state.startOffset = line.start
		}
		if line.text == end {
			endCount++
			state.endIndex = idx
			state.endOffset = line.start
		}
	}
	if startCount > 1 || endCount > 1 {
		return state, fail(exitMarkerState, fmt.Sprintf("marker appears more than once: %s count=%d, %s count=%d", start, startCount, end, endCount)).
			withHint("Edit the target file so it contains exactly one start marker and one later end marker, or remove the broken block manually.")
	}
	if startCount == 1 && endCount == 0 {
		return state, fail(exitMarkerState, fmt.Sprintf("start marker exists without end marker: %s", start)).
			withHint("Edit the target file to add the matching end marker or remove the incomplete managed block manually.")
	}
	if startCount == 0 && endCount == 1 {
		return state, fail(exitMarkerState, fmt.Sprintf("end marker exists without start marker: %s", end)).
			withHint("Edit the target file to add the matching start marker before it or remove the stray end marker manually.")
	}
	if startCount == 1 && endCount == 1 && state.endIndex < state.startIndex {
		return state, fail(exitMarkerState, fmt.Sprintf("end marker appears before start marker: %s before %s", end, start)).
			withHint("Edit the target file so the start marker appears before the end marker, or remove the broken block manually.")
	}
	state.exists = startCount == 1 && endCount == 1
	return state, nil
}

func validateBodyMarkers(marker, body string) *userError {
	start := "##" + marker + "^"
	end := "##" + marker + "$"
	lines := lineTexts(ensureFinalNewline(normalizeLF(body)))
	startCount := 0
	endCount := 0
	for _, line := range lines {
		if line == start {
			startCount++
		}
		if line == end {
			endCount++
		}
	}
	if startCount > 0 || endCount > 0 {
		return fail(exitInputValidation, fmt.Sprintf("body contains marker delimiter line: %s count=%d, %s count=%d", start, startCount, end, endCount)).
			withHint("Remove exact marker delimiter lines from stdin, or choose a different marker with --marker.")
	}
	return nil
}

func managedBlockLines(marker, body string) []string {
	content := lineTexts(ensureFinalNewline(normalizeLF(body)))
	block := make([]string, 0, len(content)+2)
	block = append(block, "##"+marker+"^")
	block = append(block, content...)
	block = append(block, "##"+marker+"$")
	return block
}

type editResult struct {
	text          string
	startLine     int
	affectedCount int
}

func replaceOrAppend(old, marker, body string, state markerState) editResult {
	lines := lineTexts(old)
	block := managedBlockLines(marker, body)
	if !state.exists {
		newLines := append(slices.Clone(lines), block...)
		return editResult{
			text:          linesToText(newLines, true),
			startLine:     len(lines) + 1,
			affectedCount: len(block),
		}
	}
	prefix := slices.Clone(lines[:state.startIndex])
	suffix := slices.Clone(lines[state.endIndex+1:])
	newLines := append(prefix, block...)
	newLines = append(newLines, suffix...)
	finalNewline := true
	if len(suffix) > 0 {
		finalNewline = strings.HasSuffix(old, "\n")
	}
	return editResult{
		text:          linesToText(newLines, finalNewline),
		startLine:     state.startIndex + 1,
		affectedCount: len(block),
	}
}

func removeBlock(old, marker string, state markerState) editResult {
	if !state.exists {
		return editResult{text: old}
	}
	lines := lineTexts(old)
	prefix := slices.Clone(lines[:state.startIndex])
	suffix := slices.Clone(lines[state.endIndex+1:])
	newLines := append(prefix, suffix...)
	finalNewline := len(prefix) > 0
	if len(suffix) > 0 {
		finalNewline = strings.HasSuffix(old, "\n")
	}
	return editResult{
		text:      linesToText(newLines, finalNewline),
		startLine: state.startIndex + 1,
	}
}

func validateShellTarget(text, marker string, state markerState) *userError {
	file, err := parseShell(text)
	if err != nil {
		return fail(exitSyntax, fmt.Sprintf("target file is not valid bash syntax: %v", err)).
			withHint("Fix the target file syntax, or use --plain-text for non-shell text.")
	}
	if !state.exists {
		if stack := contextStack(file, len(text)); len(stack) != 0 {
			return fail(exitBoundary, "append position is not at top-level shell context").
				withHint("Close the surrounding shell syntax before the end of file, or use --plain-text.")
		}
		return nil
	}
	startStack := contextStack(file, state.startOffset)
	endStack := contextStack(file, state.endOffset)
	if !equalScopes(startStack, endStack) {
		return fail(exitBoundary, fmt.Sprintf("marker block crosses shell syntax boundaries: %s to %s", state.start, state.end)).
			withHint("Move the marker lines so both delimiters are in the same shell context.")
	}
	placeholder := placeholderText(text, state)
	if _, err := parseShell(placeholder); err != nil {
		return fail(exitBoundary, fmt.Sprintf("marker block crosses required shell grammar boundaries: %v", err)).
			withHint("Move the marker lines so replacing the block cannot break surrounding shell syntax.")
	}
	return nil
}

func parseShell(text string) (*syntax.File, error) {
	parser := syntax.NewParser(syntax.KeepComments(true), syntax.Variant(syntax.LangBash))
	return parser.Parse(strings.NewReader(text), "target")
}

func placeholderText(text string, state markerState) string {
	lines := lineTexts(text)
	replacement := []string{placeholderLine(lines)}
	prefix := slices.Clone(lines[:state.startIndex])
	suffix := slices.Clone(lines[state.endIndex+1:])
	newLines := append(prefix, replacement...)
	newLines = append(newLines, suffix...)
	finalNewline := true
	if len(suffix) > 0 {
		finalNewline = strings.HasSuffix(text, "\n")
	}
	return linesToText(newLines, finalNewline)
}

func placeholderLine(lines []string) string {
	for idx := 0; ; idx++ {
		candidate := fmt.Sprintf("__SASEO_PLACEHOLDER_%d__", idx)
		if !slices.Contains(lines, candidate) {
			return candidate
		}
	}
}

type scope struct {
	kind  string
	start int
	end   int
}

func contextStack(file *syntax.File, offset int) []scope {
	var scopes []scope
	syntax.Walk(file, func(node syntax.Node) bool {
		if node == nil {
			return true
		}
		switch node.(type) {
		case *syntax.File, *syntax.Stmt, *syntax.Comment:
			return true
		}
		if scope, ok := nodeScope(node); ok && containsOffset(scope, offset) {
			scopes = append(scopes, scope)
		}
		addPartitionScopes(node, offset, &scopes)
		return true
	})
	slices.SortFunc(scopes, func(a, b scope) int {
		if a.start != b.start {
			return a.start - b.start
		}
		if a.end != b.end {
			return b.end - a.end
		}
		return strings.Compare(a.kind, b.kind)
	})
	return scopes
}

func nodeScope(node syntax.Node) (scope, bool) {
	pos := node.Pos()
	end := node.End()
	if !pos.IsValid() || !end.IsValid() {
		return scope{}, false
	}
	startOffset := int(pos.Offset())
	endOffset := int(end.Offset())
	if endOffset <= startOffset {
		return scope{}, false
	}
	return scope{kind: fmt.Sprintf("%T", node), start: startOffset, end: endOffset}, true
}

func addPartitionScopes(node syntax.Node, offset int, scopes *[]scope) {
	switch n := node.(type) {
	case *syntax.IfClause:
		addDelimitedScope(scopes, "if-cond", n.Position, n.ThenPos, offset)
		thenEnd := n.FiPos
		if n.Else != nil {
			thenEnd = n.Else.Position
		}
		addDelimitedScope(scopes, "if-then", n.ThenPos, thenEnd, offset)
	case *syntax.WhileClause:
		addDelimitedScope(scopes, "while-cond", n.WhilePos, n.DoPos, offset)
		addDelimitedScope(scopes, "while-do", n.DoPos, n.DonePos, offset)
	case *syntax.ForClause:
		addDelimitedScope(scopes, "for-head", n.ForPos, n.DoPos, offset)
		addDelimitedScope(scopes, "for-do", n.DoPos, n.DonePos, offset)
	case *syntax.CaseClause:
		addDelimitedScope(scopes, "case-head", n.Case, n.In, offset)
		for _, item := range n.Items {
			addDelimitedScope(scopes, "case-item", item.OpPos, item.End(), offset)
		}
	}
}

func addDelimitedScope(scopes *[]scope, kind string, start, end syntax.Pos, offset int) {
	if !start.IsValid() || !end.IsValid() {
		return
	}
	scope := scope{kind: kind, start: int(start.Offset()), end: int(end.Offset())}
	if containsOffset(scope, offset) {
		*scopes = append(*scopes, scope)
	}
}

func containsOffset(scope scope, offset int) bool {
	return scope.start <= offset && offset < scope.end
}

func equalScopes(a, b []scope) bool {
	return slices.Equal(a, b)
}

func atomicSave(path, text string, mode os.FileMode) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	tmp, err := os.CreateTemp(dir, "."+base+".saseo.*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.WriteString(text); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	return nil
}

func shellQuote(text string) string {
	if text == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(text, "'", "'\"'\"'") + "'"
}
