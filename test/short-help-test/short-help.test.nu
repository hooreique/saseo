# Spec: SASEO-CMD
let root = ($env.FILE_PWD | path join ../..)
def main [] {
    let saseo = ($root | path join saseo.nu)
    let expected_example = (open --raw ($root | path join EXAMPLE))
    let result = do { ^nu $saseo -h } | complete
    if ($result.exit_code | into string) != "0" {
        print "not ok - short help exit code"
        print "expected:"
        print "0"
        print "actual:"
        print ($result.exit_code | into string)
        exit 1
    }
    print "ok - short help exit code"
    for check in [
        {name: "short help stdout", text: "Usage:"}
        {name: "short help includes dry-run option", text: "  --dry-run"}
        {name: "short help includes man hint", text: "  man saseo"}
        {name: "short help includes example heading", text: "Example:"}
        {
            name: "short help includes example file"
            text: $expected_example
        }
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
    for check in [
        {name: "short help omits short help usage", text: "  saseo -h\n"}
        {name: "short help omits short version usage", text: "  saseo -V\n"}
    ] {
        if ($result.stdout | str contains $check.text) {
            print $"not ok - ($check.name)"
            print "unexpected substring:"
            print $check.text
            print "actual:"
            print $result.stdout
            exit 1
        }
        print $"ok - ($check.name)"
    }
}
