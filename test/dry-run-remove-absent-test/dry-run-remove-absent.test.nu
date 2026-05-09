# Spec: SASEO-DRY-RUN, SASEO-REMOVE
let root = ($env.FILE_PWD | path join ../..)
let test_dir = $env.FILE_PWD
def main [] {
    let saseo = ($root | path join saseo.nu)
    let work = ($test_dir | path join target.temp.txt)
    cp --force ($test_dir | path join input.txt) $work
    ^touch -t 202001010000 $work
    let before_meta = (ls -l $work | get 0 | select inode modified size)
    let result = do {
        "ignored stdin\n" | ^nu $saseo --dry-run --rm --marker SASEO $work
    } | complete
    if ($result.exit_code | into string) != "0" {
        print "not ok - dry-run remove absent managed block ignores stdin exit code"
        print "expected:"
        print "0"
        print "actual:"
        print ($result.exit_code | into string)
        exit 1
    }
    print "ok - dry-run remove absent managed block ignores stdin exit code"
    if $result.stdout != "" {
        print "not ok - dry-run remove absent managed block stdout"
        print "expected:"
        print ""
        print "actual:"
        print $result.stdout
        exit 1
    }
    print "ok - dry-run remove absent managed block stdout"
    let actual = (open --raw $work)
    let expected = (open --raw ($test_dir | path join input.txt))
    if $actual != $expected {
        print "not ok - dry-run remove absent leaves target contents untouched"
        print "expected:"
        print $expected
        print "actual:"
        print $actual
        exit 1
    }
    print "ok - dry-run remove absent leaves target contents untouched"
    let after_meta = (ls -l $work | get 0 | select inode modified size)
    if $after_meta != $before_meta {
        print "not ok - dry-run remove absent leaves target metadata untouched"
        print "expected:"
        print $before_meta
        print "actual:"
        print $after_meta
        exit 1
    }
    print "ok - dry-run remove absent leaves target metadata untouched"
    rm --force $work
}
