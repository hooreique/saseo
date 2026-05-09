{
  lib,
  stdenvNoCC,
  makeWrapper,
  installShellFiles,
  nushell,
  nufmt,
  scdoc,
}:

stdenvNoCC.mkDerivation {
  pname = "saseo";
  version = lib.fileContents ./VERSION;

  src = ./.;

  nativeBuildInputs = [
    makeWrapper
    installShellFiles
    nushell
    nufmt
    scdoc
  ];

  buildPhase = ''
    runHook preBuild
    scdoc < saseo.1.scd > saseo.1
    runHook postBuild
  '';

  doCheck = true;

  checkPhase = ''
    runHook preCheck

    nu_files="$(find test -name '*.nu' -type f | sort)"
    for file in saseo.nu $nu_files; do
      formatted="$(mktemp)"
      nufmt --stdin < "$file" > "$formatted"
      if ! cmp -s "$file" "$formatted"; then
        echo "$file is not formatted" >&2
        diff -u "$file" "$formatted" >&2 || true
        exit 1
      fi
    done

    nu --ide-check 0 saseo.nu > /dev/null
    for file in $nu_files; do
      nu --ide-check 0 "$file" > /dev/null
    done

    nu test/saseo-test.nu

    runHook postCheck
  '';

  installPhase = ''
    runHook preInstall

    install -Dm644 saseo.nu $out/share/saseo/saseo.nu
    install -Dm644 VERSION $out/share/saseo/VERSION
    install -Dm644 EXAMPLE $out/share/saseo/EXAMPLE
    installManPage saseo.1
    mkdir -p $out/bin
    makeWrapper ${nushell}/bin/nu $out/bin/saseo \
      --add-flags "$out/share/saseo/saseo.nu"

    runHook postInstall
  '';

  doInstallCheck = true;

  installCheckPhase = ''
    runHook preInstallCheck

    $out/bin/saseo --help > /dev/null
    $out/bin/saseo --version > /dev/null
    test -s "$out/share/man/man1/saseo.1" || test -s "$out/share/man/man1/saseo.1.gz"

    runHook postInstallCheck
  '';

  meta = {
    description = "saseo adds and removes small rc snippets that should stick around for a while, but not forever.";
    license = lib.licenses.mit;
    mainProgram = "saseo";
  };
}
