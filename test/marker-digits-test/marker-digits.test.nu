# Spec: SASEO-MARKER-NAME, SASEO-MARKER-FORMAT, SASEO-ADD-REPLACE
let root = ($env.FILE_PWD | path join ../..)
let test_dir = $env.FILE_PWD
def main [] {
    let saseo = ($root | path join saseo.nu)
    let work = ($test_dir | path join target.temp.txt)
    cp --force ($test_dir | path join input.txt) $work
    let result = do {
        "export N=1" | ^nu $saseo --marker NODE22 $work
    } | complete
    if ($result.exit_code | into string) != "0" {
        print "not ok - marker digits are valid exit code"
        print "expected:"
        print "0"
        print "actual:"
        print ($result.exit_code | into string)
        exit 1
    }
    print "ok - marker digits are valid exit code"
    if $result.stdout != "3 3\n" {
        print "not ok - marker digits are valid stdout"
        print "expected:"
        print "3 3\n"
        print "actual:"
        print $result.stdout
        exit 1
    }
    print "ok - marker digits are valid stdout"
    let actual = (open --raw $work)
    if not ($actual | str contains "##NODE22^\n") {
        print "not ok - marker digits output"
        print "expected substring:"
        print "##NODE22^\n"
        print "actual:"
        print $actual
        exit 1
    }
    print "ok - marker digits output"
    rm --force $work
}
