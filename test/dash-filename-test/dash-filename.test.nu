# Spec: SASEO-CMD, SASEO-ADD-REPLACE, SASEO-MARKER-NAME, SASEO-AFFECTED-LINES
let root = ($env.FILE_PWD | path join ../..)
let test_dir = $env.FILE_PWD
def main [] {
    let saseo = ($root | path join saseo.nu)
    let dash_file = ($test_dir | path join "-weird-file")
    rm --force $dash_file
    cp --force ($test_dir | path join input.txt) $dash_file
    let result = do {
        cd $test_dir
        "export N=1" | ^nu $saseo --marker DASH -- "-weird-file"
    } | complete
    if ($result.exit_code | into string) != "0" {
        print "not ok - dash filename after end of options exit code"
        print "expected:"
        print "0"
        print "actual:"
        print ($result.exit_code | into string)
        rm --force $dash_file
        exit 1
    }
    print "ok - dash filename after end of options exit code"
    if $result.stdout != "3 3\n" {
        print "not ok - dash filename after end of options stdout"
        print "expected:"
        print "3 3\n"
        print "actual:"
        print $result.stdout
        rm --force $dash_file
        exit 1
    }
    print "ok - dash filename after end of options stdout"
    let actual = (open --raw $dash_file)
    if not ($actual | str contains "##DASH^\n") {
        print "not ok - dash filename output"
        print "expected substring:"
        print "##DASH^\n"
        print "actual:"
        print $actual
        rm --force $dash_file
        exit 1
    }
    print "ok - dash filename output"
    rm --force $dash_file
}
