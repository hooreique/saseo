# Spec: SASEO-CMD
let root = ($env.FILE_PWD | path join ../..)
def main [] {
    let saseo = ($root | path join saseo.nu)
    let result = with-env { NO_COLOR: "1" } { ^nu $saseo -- -h } | complete
    if ($result.exit_code | into string) != "0" {
        print "not ok - leading dashdash short help exit code"
        print "expected:"
        print "0"
        print "actual:"
        print ($result.exit_code | into string)
        exit 1
    }
    print "ok - leading dashdash short help exit code"
    if not ($result.stdout | str contains "Usage:") {
        print "not ok - leading dashdash short help stdout"
        print "expected substring:"
        print "Usage:"
        print "actual:"
        print $result.stdout
        exit 1
    }
    print "ok - leading dashdash short help stdout"
}
