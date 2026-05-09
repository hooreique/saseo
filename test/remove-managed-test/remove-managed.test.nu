# Spec: SASEO-REMOVE, SASEO-AFFECTED-LINES
let root = ($env.FILE_PWD | path join ../..)
let test_dir = $env.FILE_PWD
def main [] {
    let saseo = ($root | path join saseo.nu)
    let work = ($test_dir | path join target.temp.txt)
    cp --force ($test_dir | path join input.txt) $work
    let result = do {
        "ignored stdin\n" | ^nu $saseo --rm --marker SASEO $work
    } | complete
    if ($result.exit_code | into string) != "0" {
        print "not ok - remove managed block ignores stdin exit code"
        print "expected:"
        print "0"
        print "actual:"
        print ($result.exit_code | into string)
        exit 1
    }
    print "ok - remove managed block ignores stdin exit code"
    if $result.stdout != "3\n" {
        print "not ok - remove managed block ignores stdin stdout"
        print "expected:"
        print "3\n"
        print "actual:"
        print $result.stdout
        exit 1
    }
    print "ok - remove managed block ignores stdin stdout"
    let actual = (open --raw $work)
    let expected = (open --raw ($test_dir | path join expected.txt))
    if $actual != $expected {
        print "not ok - remove managed block"
        print "expected:"
        print $expected
        print "actual:"
        print $actual
        exit 1
    }
    print "ok - remove managed block"
    rm --force $work
}
