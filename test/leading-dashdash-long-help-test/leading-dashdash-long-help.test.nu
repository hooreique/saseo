# Spec: SASEO-CMD
let root = ($env.FILE_PWD | path join ../..)
def main [] {
    let saseo = ($root | path join saseo.nu)
    let result = do { ^nu $saseo -- --help } | complete
    if ($result.exit_code | into string) != "0" {
        print "not ok - leading dashdash long help exit code"
        print "expected:"
        print "0"
        print "actual:"
        print ($result.exit_code | into string)
        exit 1
    }
    print "ok - leading dashdash long help exit code"
    if not ($result.stdout | str contains "Usage:") {
        print "not ok - leading dashdash long help stdout"
        print "expected substring:"
        print "Usage:"
        print "actual:"
        print $result.stdout
        exit 1
    }
    print "ok - leading dashdash long help stdout"
}
