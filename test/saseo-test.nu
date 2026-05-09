let test_root = $env.FILE_PWD
def main [] {
    let scripts = (glob ($test_root | path join "*-test" "*.test.nu") | sort)
    if ($scripts | length) == 0 {
        print --stderr "no test scripts found"
        exit 1
    }
    for script in $scripts {
        print $"running ($script)"
        let result = do { ^nu $script } | complete
        if $result.stdout != "" {
            print --no-newline $result.stdout
        }
        if $result.stderr != "" {
            print --stderr --no-newline $result.stderr
        }
        if $result.exit_code != 0 {
            print --stderr $"failed ($script)"
            exit $result.exit_code
        }
    }
}
