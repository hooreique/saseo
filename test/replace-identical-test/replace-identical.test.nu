# Spec: SASEO-ADD-REPLACE, SASEO-TARGET-FILE, SASEO-AFFECTED-LINES
let root = ($env.FILE_PWD | path join ../..)
let test_dir = $env.FILE_PWD
def main [] {
    let saseo = ($root | path join saseo.nu)
    let work = ($test_dir | path join target.temp.txt)
    cp --force ($test_dir | path join input.txt) $work
    ^touch -t 202001010000 $work
    let before_meta = (ls -l $work | get 0 | select inode modified size)
    let result = do {
        "same\n" | ^nu $saseo $work
    } | complete
    if ($result.exit_code | into string) != "0" {
        print "not ok - replace identical exit code"
        print "expected:"
        print "0"
        print "actual:"
        print ($result.exit_code | into string)
        exit 1
    }
    print "ok - replace identical exit code"
    if $result.stdout != "2 3\n" {
        print "not ok - replace identical stdout"
        print "expected:"
        print "2 3\n"
        print "actual:"
        print $result.stdout
        exit 1
    }
    print "ok - replace identical stdout"
    let actual = (open --raw $work)
    let expected = (open --raw ($test_dir | path join input.txt))
    if $actual != $expected {
        print "not ok - replace identical content"
        print "expected:"
        print $expected
        print "actual:"
        print $actual
        exit 1
    }
    print "ok - replace identical content"
    let after_meta = (ls -l $work | get 0 | select inode modified size)
    if $after_meta != $before_meta {
        print "not ok - replace identical leaves target untouched"
        print "expected:"
        print $before_meta
        print "actual:"
        print $after_meta
        exit 1
    }
    print "ok - replace identical leaves target untouched"
    rm --force $work
}
