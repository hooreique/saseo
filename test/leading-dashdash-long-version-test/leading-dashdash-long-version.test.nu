# Spec: SASEO-CMD
let root = ($env.FILE_PWD | path join ../..)
def main [] {
    let saseo = ($root | path join saseo.nu)
    let expected_version = (open --raw ($root | path join VERSION) | str trim)
    let result = do { ^nu $saseo -- --version } | complete
    if ($result.exit_code | into string) != "0" {
        print "not ok - leading dashdash long version exit code"
        print "expected:"
        print "0"
        print "actual:"
        print ($result.exit_code | into string)
        exit 1
    }
    print "ok - leading dashdash long version exit code"
    let expected = $"($expected_version)\n"
    if $result.stdout != $expected {
        print "not ok - leading dashdash long version stdout"
        print "expected:"
        print $expected
        print "actual:"
        print $result.stdout
        exit 1
    }
    print "ok - leading dashdash long version stdout"
}
