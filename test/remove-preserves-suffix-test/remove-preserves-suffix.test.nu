# Spec: SASEO-REMOVE, SASEO-AFFECTED-LINES
let root = ($env.FILE_PWD | path join ../..)
let test_dir = $env.FILE_PWD
def main [] {
    let saseo = ($root | path join saseo.nu)
    let work = ($test_dir | path join target.temp.txt)
    cp --force ($test_dir | path join input.txt) $work
    let result = do { ^nu $saseo --rm $work } | complete
    if ($result.exit_code | into string) != "0" {
        print "not ok - remove preserves suffix exit code"
        print "expected:"
        print "0"
        print "actual:"
        print ($result.exit_code | into string)
        exit 1
    }
    print "ok - remove preserves suffix exit code"
    if $result.stdout != "2\n" {
        print "not ok - remove preserves suffix stdout"
        print "expected:"
        print "2\n"
        print "actual:"
        print $result.stdout
        exit 1
    }
    print "ok - remove preserves suffix stdout"
    let actual = (open --raw $work)
    let expected = (open --raw ($test_dir | path join expected.txt))
    if $actual != $expected {
        print "not ok - remove preserves suffix"
        print "expected:"
        print $expected
        print "actual:"
        print $actual
        exit 1
    }
    print "ok - remove preserves suffix"
    rm --force $work
}
