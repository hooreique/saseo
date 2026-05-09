# Spec: SASEO-ADD-REPLACE, SASEO-APPEND-FORMATTING, SASEO-AFFECTED-LINES
let root = ($env.FILE_PWD | path join ../..)
let test_dir = $env.FILE_PWD
def main [] {
    let saseo = ($root | path join saseo.nu)
    let work = ($test_dir | path join target.temp.txt)
    cp --force ($test_dir | path join input.txt) $work
    let result = do {
        "export N=1" | ^nu $saseo $work
    } | complete
    if ($result.exit_code | into string) != "0" {
        print "not ok - append to empty file exit code"
        print "expected:"
        print "0"
        print "actual:"
        print ($result.exit_code | into string)
        exit 1
    }
    print "ok - append to empty file exit code"
    if $result.stdout != "1 3\n" {
        print "not ok - append to empty file stdout"
        print "expected:"
        print "1 3\n"
        print "actual:"
        print $result.stdout
        exit 1
    }
    print "ok - append to empty file stdout"
    let actual = (open --raw $work)
    let expected = (open --raw ($test_dir | path join expected.txt))
    if $actual != $expected {
        print "not ok - append to empty file"
        print "expected:"
        print $expected
        print "actual:"
        print $actual
        exit 1
    }
    print "ok - append to empty file"
    rm --force $work
}
