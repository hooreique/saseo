{
  lib,
  buildGoModule,
  go-tools,
  gosec,
  installShellFiles,
  scdoc,
}:

buildGoModule {
  pname = "saseo";
  version = lib.fileContents ./VERSION;

  src = lib.fileset.toSource {
    root = ./.;
    fileset = lib.fileset.unions [
      ./VERSION
      ./go.mod
      ./go.sum
      ./main.go
      ./main_test.go
      ./saseo.1.scd
      ./test
    ];
  };

  vendorHash = "sha256-tCFu9E2pFBWBQFiRVvI16FNI3dE1bUKJlsEbvDAo7lo=";

  nativeBuildInputs = [
    installShellFiles
    scdoc
  ];

  nativeCheckInputs = [
    go-tools
    gosec
  ];

  postBuild = ''
    scdoc < saseo.1.scd > saseo.1
  '';

  doCheck = true;

  checkPhase = ''
    runHook preCheck

    export HOME="$TMPDIR"

    gofmt_files="$(gofmt -l .)"
    if [ -n "$gofmt_files" ]; then
      echo "Go files are not formatted:" >&2
      echo "$gofmt_files" >&2
      exit 1
    fi

    go test ./...
    staticcheck ./...
    gosec -quiet ./...

    runHook postCheck
  '';

  postInstall = ''
    installManPage saseo.1
  '';

  doInstallCheck = true;

  installCheckPhase = ''
    runHook preInstallCheck

    $out/bin/saseo --help > /dev/null
    $out/bin/saseo --version > /dev/null
    $out/bin/saseo put --help > /dev/null
    $out/bin/saseo mark --help > /dev/null
    $out/bin/saseo rm --help > /dev/null
    $out/bin/saseo show --help > /dev/null
    test -s "$out/share/man/man1/saseo.1" || test -s "$out/share/man/man1/saseo.1.gz"

    runHook postInstallCheck
  '';

  meta = {
    description = "bash-aware block management for shell rc files";
    license = lib.licenses.mit;
    mainProgram = "saseo";
  };
}
