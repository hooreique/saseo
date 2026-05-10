# Spec: SASEO-CMD
let root = ($env.FILE_PWD | path join ../..)
def main [] {
    let saseo = ($root | path join saseo.nu)
    let ansi_prefix = $"(char --unicode "1b")["
    let result = do { ^nu $saseo --help } | complete
    if ($result.exit_code | into string) != "0" {
        print "not ok - captured help exit code"
        print "expected:"
        print "0"
        print "actual:"
        print ($result.exit_code | into string)
        exit 1
    }
    print "ok - captured help exit code"
    if ($result.stdout | str contains $ansi_prefix) {
        print "not ok - captured help is plain"
        print "actual:"
        print $result.stdout
        exit 1
    }
    print "ok - captured help is plain"
    for check in [
        {name: "plain help keeps usage", text: "Usage:"}
        {name: "plain help keeps dry-run option", text: "  --dry-run"}
    ] {
        if not ($result.stdout | str contains $check.text) {
            print $"not ok - ($check.name)"
            print "expected substring:"
            print $check.text
            print "actual:"
            print $result.stdout
            exit 1
        }
        print $"ok - ($check.name)"
    }
}
