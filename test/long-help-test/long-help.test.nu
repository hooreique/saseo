# Spec: SASEO-CMD
let root = ($env.FILE_PWD | path join ../..)
def main [] {
    let saseo = ($root | path join saseo.nu)
    let expected_example = (open --raw ($root | path join EXAMPLE))
    let result = do { ^nu $saseo --help } | complete
    if ($result.exit_code | into string) != "0" {
        print "not ok - long help exit code"
        print "expected:"
        print "0"
        print "actual:"
        print ($result.exit_code | into string)
        exit 1
    }
    print "ok - long help exit code"
    for check in [
        {name: "long help stdout", text: "Usage:"}
        {name: "long help includes dry-run option", text: "  --dry-run"}
        {name: "long help includes man hint", text: "  man saseo"}
        {name: "long help includes example heading", text: "Example:"}
        {
            name: "long help includes example file"
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
        {name: "long help omits short help usage", text: "  saseo -h\n"}
        {name: "long help omits short version usage", text: "  saseo -V\n"}
        {name: "long help omits leading dashdash short forms", text: "saseo -- -"}
        {name: "long help omits leading dashdash long forms", text: "saseo -- --"}
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
