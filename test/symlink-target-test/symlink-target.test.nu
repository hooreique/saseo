# Spec: SASEO-TARGET-FILE, SASEO-EXIT-CODES, SASEO-ERROR-OUTPUT
let root = ($env.FILE_PWD | path join ../..)
let test_dir = $env.FILE_PWD
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
    let link_target = ($test_dir | path join target)
    let link_path = ($test_dir | path join link)
    rm --force $link_path $link_target
    cp --force ($test_dir | path join input.txt) $link_target
    ^ln -s $link_target $link_path
    let result = do {
        "export N=1" | ^nu $saseo $link_path
    } | complete
    if ($result.exit_code | into string) != "8" {
        print "not ok - symlink target exit code"
        print "expected:"
        print "8"
        print "actual:"
        print ($result.exit_code | into string)
        rm --force $link_path $link_target
        exit 1
    }
    print "ok - symlink target exit code"
    assert_error_output_format $result.stderr "symlink target"
    rm --force $link_path $link_target
}
