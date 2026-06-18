{
  lib,
  buildGoModule,
}:

buildGoModule {
  pname = "pueue-discord-notify";
  version = "0.1.0";

  src = lib.fileset.toSource {
    root = ./.;
    fileset = lib.fileset.unions [
      ./go.mod
      ./main.go
      ./main_test.go
    ];
  };

  vendorHash = null;

  ldflags = [
    "-s"
    "-w"
  ];

  doCheck = true;

  checkPhase = ''
    runHook preCheck
    go test -v ./...
    runHook postCheck
  '';

  meta = {
    description = "Simple Discord webhook notifier for pueue task completion";
    homepage = "https://github.com/takata/pueue-discord-notify";
    license = lib.licenses.mit;
    maintainers = [ ];
    mainProgram = "pueue-discord-notify";
  };
}
