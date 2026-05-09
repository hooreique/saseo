# Spec: SASEO-INVALID-OPTIONS, SASEO-EXIT-CODES, SASEO-ERROR-OUTPUT
let root = ($env.FILE_PWD | path join ../..)
def has_prefixed_content [line: string, prefix: string] { ($line | str starts-with $prefix) and (($line | str length) > ($prefix | str length)) }
def assert_error_output_format [stderr: string, label: string] {
    if $stderr == "" {
        print $"not ok - ($label) stderr format"
        print "expected:"
        print "non-empty stderr"
        print "actual:"
        print $stderr
        exit 1
    }
    let lines = ($stderr | str trim --right | split row "\n")
    if not (has_prefixed_content ($lines | get 0) "error: ") {
        print $"not ok - ($label) stderr format"
        print "expected:"
        print "first stderr line to start with error: and include content"
        print "actual:"
        print $stderr
        exit 1
    }
    for line in ($lines | skip 1) {
        if not ((has_prefixed_content $line "hint: ") or (has_prefixed_content $line "try: ")) {
            print $"not ok - ($label) stderr format"
            print "expected:"
            print "later stderr lines to start with hint: or try: and include content"
            print "actual:"
            print $stderr
            exit 1
        }
    }
    print $"ok - ($label) stderr format"
}
def main [] {
    let saseo = ($root | path join saseo.nu)
    let result = do { ^nu $saseo } | complete
    if ($result.exit_code | into string) != "3" {
        print "not ok - missing file argument exit code"
        print "expected:"
        print "3"
        print "actual:"
        print ($result.exit_code | into string)
        exit 1
    }
    print "ok - missing file argument exit code"
    assert_error_output_format $result.stderr "missing file argument"
}
