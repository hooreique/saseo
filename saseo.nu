const EXIT_USAGE = 3
const EXIT_EMPTY_STDIN = 4
const EXIT_MISSING_TARGET = 5
const EXIT_INPUT_VALIDATION = 6
const EXIT_MARKER_STATE = 7
const EXIT_UNSUPPORTED_TARGET = 8
def ensure_final_newline [text: string] {
    if ($text == "" or ($text | str ends-with "\n")) {
        $text
    } else {
        $"($text)\n"
    }
}
def shell_quote [text: string] {
    if $text == "" {
        "''"
    } else {
        let escaped = ($text | str replace --all "'" "'\"'\"'")
        $"'($escaped)'"
    }
}
def saseo_command [file: string, marker: string, rm: bool] {
    let rm_arg = if $rm { " --rm" } else { "" }
    let marker_arg = if $marker == "SASEO" {
        ""
    } else {
        $" --marker (shell_quote $marker)"
    }
    $"saseo($rm_arg)($marker_arg) -- (shell_quote $file)"
}
def fail [
    code: int
    message: string
    --hint: string
    --try-command: string
] {
    print --stderr $"error: ($message)"
    if $hint != null {
        print --stderr $"hint: ($hint)"
    }
    if $try_command != null {
        print --stderr $"try: ($try_command)"
    }
    exit $code
}
def color_enabled [] {
    let no_color = ($env.NO_COLOR? | default "")
    $no_color == "" and (is-terminal --stdout)
}
def color_text [enabled: bool, code: string, text: string] {
    if $enabled {
        $"(ansi $code)($text)(ansi reset)"
    } else {
        $text
    }
}
def usage [] {
    let color = (color_enabled)
    let heading = {|text| color_text $color "cyan_bold" $text }
    let opt = {|text| color_text $color "green" $text }
    print "saseo adds and removes small rc snippets that should stick around for a while,"
    print "but not forever."
    print ""
    print (do $heading "Usage:")
    print "  saseo [--dry-run] [--marker MARKER] [--] FILE"
    print "  saseo --rm [--dry-run] [--marker MARKER] [--] FILE"
    print ""
    print (do $heading "Examples:")
    print (open --raw ($env.FILE_PWD | path join EXAMPLE) | str trim --right)
    print ""
    print "saseo --rm ~/.bashrc"
    print ""
    print (do $heading "Options:")
    let dry_run = (do $opt "--dry-run")
    let marker = (do $opt "--marker MARKER")
    let rm = (do $opt "--rm")
    let help = (do $opt "-h, --help")
    let version = (do $opt "-V, --version")
    print $"  ($dry_run)          Validate and print the same success output without writing"
    print $"  ($marker)    Use marker name MARKER; default: SASEO"
    print $"  ($rm)               Remove the managed block instead of adding or replacing it"
    print $"  ($help)         Show help"
    print $"  ($version)      Show version"
    print ""
    print (do $heading "More details:")
    print "  man saseo"
}
def version [] {
    open --raw ($env.FILE_PWD | path join VERSION) | str trim
}
def count_newlines [text: string] { (($text | split row "\n" | length) - 1) }
def line_count [text: string] {
    if $text == "" {
        0
    } else if ($text | str ends-with "\n") {
        count_newlines $text
    } else {
        (count_newlines $text) + 1
    }
}
def normalize_lf [text: string] {
    $text | str replace --all "\r\n" "\n"
}
def text_to_lines [text: string] {
    if $text == "" {
        []
    } else {
        let lines = ($text | split row "\n")
        if ($text | str ends-with "\n") {
            $lines | drop 1
        } else {
            $lines
        }
    }
}
def lines_to_text [lines: list<string>, final_newline: bool] {
    if ($lines | length) == 0 {
        ""
    } else {
        let text = ($lines | str join "\n")
        if $final_newline {
            $"($text)\n"
        } else {
            $text
        }
    }
}
def marker_line_indices [lines: list<string>, needle: string] {
    $lines | enumerate | where item == $needle | get index
}
def managed_block_lines [marker: string, body: string] {
    let start = $"##($marker)^"
    let end = $"##($marker)$"
    let content = (text_to_lines (ensure_final_newline (normalize_lf $body)))
    [$start] | append $content | append [$end]
}
def validate_body_markers [marker: string, body: string] {
    let start = $"##($marker)^"
    let end = $"##($marker)$"
    let lines = (text_to_lines (ensure_final_newline (normalize_lf $body)))
    let start_count = (marker_line_indices $lines $start | length)
    let end_count = (marker_line_indices $lines $end | length)
    if ($start_count > 0 or $end_count > 0) {
        fail $EXIT_INPUT_VALIDATION $"body contains marker delimiter line: ($start) count=($start_count), ($end) count=($end_count)" --hint "Remove exact marker delimiter lines from stdin, or choose a different marker with --marker."
    }
}
def marker_state [lines: list<string>, marker: string] {
    let start = $"##($marker)^"
    let end = $"##($marker)$"
    let start_indices = (marker_line_indices $lines $start)
    let end_indices = (marker_line_indices $lines $end)
    let start_count = ($start_indices | length)
    let end_count = ($end_indices | length)
    if ($start_count > 1 or $end_count > 1) {
        fail $EXIT_MARKER_STATE $"marker appears more than once: ($start) count=($start_count), ($end) count=($end_count)" --hint "Edit the target file so it contains exactly one start marker and one later end marker, or remove the broken block manually."
    }
    if ($start_count == 1 and $end_count == 0) {
        fail $EXIT_MARKER_STATE $"start marker exists without end marker: ($start)" --hint "Edit the target file to add the matching end marker or remove the incomplete managed block manually."
    }
    if ($start_count == 0 and $end_count == 1) {
        fail $EXIT_MARKER_STATE $"end marker exists without start marker: ($end)" --hint "Edit the target file to add the matching start marker before it or remove the stray end marker manually."
    }
    let start_idx = if $start_count == 1 {
        $start_indices | get 0
    } else { null }
    let end_idx = if $end_count == 1 {
        $end_indices | get 0
    } else { null }
    if ($start_idx != null and $end_idx != null and $end_idx < $start_idx) {
        fail $EXIT_MARKER_STATE $"end marker appears before start marker: ($end) before ($start)" --hint "Edit the target file so the start marker appears before the end marker, or remove the broken block manually."
    }
    {
        start: $start
        end: $end
        start_idx: $start_idx
        end_idx: $end_idx
        exists: ($start_idx != null and $end_idx != null)
    }
}
def replace_or_append [old: string, marker: string, body: string] {
    let lines = (text_to_lines $old)
    let old_final_newline = ($old | str ends-with "\n")
    let state = (marker_state $lines $marker)
    validate_body_markers $marker $body
    let block_lines = (managed_block_lines $marker $body)
    let affected_count = ($block_lines | length)
    if not $state.exists {
        let new_lines = ($lines | append $block_lines)
        {
            text: (lines_to_text $new_lines true)
            start_line: ((line_count $old) + 1)
            affected_count: $affected_count
        }
    } else {
        let prefix_lines = ($lines | take $state.start_idx)
        let suffix_lines = ($lines | skip ($state.end_idx + 1))
        let new_lines = ($prefix_lines | append $block_lines | append $suffix_lines)
        let final_newline = if ($suffix_lines | length) > 0 { $old_final_newline } else { true }
        {
            text: (lines_to_text $new_lines $final_newline)
            start_line: ($state.start_idx + 1)
            affected_count: $affected_count
        }
    }
}
def remove_block [old: string, marker: string] {
    let lines = (text_to_lines $old)
    let old_final_newline = ($old | str ends-with "\n")
    let state = (marker_state $lines $marker)
    if not $state.exists {
        {
            text: $old
            start_line: null
        }
    } else {
        let prefix_lines = ($lines | take $state.start_idx)
        let suffix_lines = ($lines | skip ($state.end_idx + 1))
        let new_lines = ($prefix_lines | append $suffix_lines)
        let final_newline = if ($suffix_lines | length) > 0 {
            $old_final_newline
        } else {
            ($prefix_lines | length) > 0
        }
        {
            text: (lines_to_text $new_lines $final_newline)
            start_line: ($state.start_idx + 1)
        }
    }
}
def is_compat_leading_dashdash [raw_args: list<any>] {
    if (($raw_args | length) < 2 or ($raw_args | get 0) != "--") {
        false
    } else {
        let next = ($raw_args | get 1)
        $next in [
            "--help"
            "-h"
            "--version"
            "-V"
            "--dry-run"
            "--rm"
            "--marker"
        ]
    }
}
def parse_args [raw_args: list<any>] {
    let args = if (is_compat_leading_dashdash $raw_args) {
        $raw_args | skip 1
    } else {
        $raw_args
    }
    mut marker = "SASEO"
    mut dry_run = false
    mut rm = false
    mut file = ""
    mut has_file = false
    mut idx = 0
    while $idx < ($args | length) {
        let arg = ($args | get $idx)
        if ($arg == "--help" or $arg == "-h") {
            return {
                help: true
                version: false
                marker: $marker
                dry_run: $dry_run
                rm: $rm
                file: $file
            }
        } else if ($arg == "--version" or $arg == "-V") {
            return {
                help: false
                version: true
                marker: $marker
                dry_run: $dry_run
                rm: $rm
                file: $file
            }
        } else if $arg == "--dry-run" {
            $dry_run = true
        } else if $arg == "--rm" {
            $rm = true
        } else if $arg == "--marker" {
            $idx = $idx + 1
            if $idx >= ($args | length) {
                fail $EXIT_USAGE "--marker requires a value" --hint "Pass a marker name after --marker." --try-command "saseo --marker WORK FILE"
            }
            $marker = ($args | get $idx)
        } else if $arg == "--" {
            $idx = $idx + 1
            while $idx < ($args | length) {
                let file_arg = ($args | get $idx)
                if $has_file {
                    fail $EXIT_USAGE $"unexpected argument: ($file_arg)" --hint "Pass exactly one target file." --try-command "saseo --help"
                }
                $file = $file_arg
                $has_file = true
                $idx = $idx + 1
            }
            break
        } else if ($arg | str starts-with "-") {
            fail $EXIT_USAGE $"unknown option: ($arg)" --hint "Use -- before FILE if the filename starts with -." --try-command "saseo --help"
        } else {
            if $has_file {
                fail $EXIT_USAGE $"unexpected argument: ($arg)" --hint "Pass exactly one target file." --try-command "saseo --help"
            }
            $file = $arg
            $has_file = true
        }
        $idx = $idx + 1
    }
    if not $has_file {
        fail $EXIT_USAGE "missing file argument" --hint "Pass the target file to update." --try-command "saseo --help"
    }
    {
        help: false
        version: false
        marker: $marker
        dry_run: $dry_run
        rm: $rm
        file: $file
    }
}
def chmod_mode [mode: string] {
    let user = ($mode | str substring 0..2 | str replace --all "-" "")
    let group = ($mode | str substring 3..5 | str replace --all "-" "")
    let other = ($mode | str substring 6..8 | str replace --all "-" "")
    $"u=($user),g=($group),o=($other)"
}
def atomic_save [path: string, text: string] {
    let dir = ($path | path dirname)
    let name = ($path | path basename)
    let tmp = (mktemp --tmpdir-path $dir $".($name).saseo.XXXXXX")
    let mode = (ls -l $path | get 0.mode)
    $text | save --force --raw $tmp
    ^chmod (chmod_mode $mode) $tmp
    mv --force $tmp $path
}
def --wrapped main [...raw_args] {
    let parsed = (parse_args $raw_args)
    if $parsed.help {
        usage
        return
    }
    if $parsed.version {
        print (version)
        return
    }
    let marker = $parsed.marker
    let dry_run = $parsed.dry_run
    let rm = $parsed.rm
    let file = $parsed.file
    if not ($marker =~ '^[A-Za-z0-9_.:][A-Za-z0-9_.:-]*$') {
        fail $EXIT_INPUT_VALIDATION "invalid marker" --hint "Use a-z, A-Z, 0-9, '_', '.', ':', and '-' after the first character." --try-command (saseo_command $file "WORK" $rm)
    }
    let path = ($file | path expand --no-symlink)
    let path_type = ($path | path type)
    if $path_type == "symlink" {
        fail $EXIT_UNSUPPORTED_TARGET $"target file must not be a symlink: ($path)" --hint "Run saseo on the real target file instead of the symlink." --try-command $"realpath (shell_quote $path)"
    }
    if $path_type == null {
        fail $EXIT_MISSING_TARGET $"target file does not exist: ($path)" --hint "Create the target file first, then run saseo again." --try-command $"touch (shell_quote $path)"
    }
    let old = (open --raw $path)
    if ($old | str contains "\r\n") {
        fail $EXIT_UNSUPPORTED_TARGET $"target file uses unsupported CRLF line endings: ($path)" --hint "Convert the target file to LF line endings before running saseo." --try-command $"perl -0pi -e 's/\\r\\n/\\n/g' (shell_quote $path)"
    }
    let stdin = if $rm { "" } else { open --raw /dev/stdin }
    if not $rm and $stdin == "" {
        fail $EXIT_EMPTY_STDIN "stdin must not be empty when adding or replacing a block" --hint "Pipe or redirect the block body into saseo." --try-command $"printf 'export FOO=foo\\n' | (saseo_command $file $marker false)"
    }
    let result = if $rm {
        remove_block $old $marker
    } else {
        replace_or_append $old $marker $stdin
    }
    if not $dry_run and $result.text != $old {
        atomic_save $path $result.text
    }
    if $rm {
        if $result.start_line != null {
            print $result.start_line
        }
    } else {
        print $"($result.start_line) ($result.affected_count)"
    }
}
