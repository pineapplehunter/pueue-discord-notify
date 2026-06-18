# pueue-discord-notify

> vibe coded

A simple Go program that sends [pueue](https://github.com/Nukesor/pueue) task
completion notifications to a Discord webhook.

## Usage

```
pueue-discord-notify \
  --webhook-file /path/to/webhook.txt \
  --id {{id}} \
  --command "{{command}}" \
  --result {{result}} \
  --exit-code {{exit_code}} \
  --group {{group}}
```

The webhook file should contain the Discord webhook URL (and nothing else).

### Host label

Use `--host` to identify which machine the task ran on. Defaults to the system
hostname.

```
pueue-discord-notify --host myserver ...
```

## Pueue configuration

Add to your `~/.config/pueue/pueue.yml`:

```yaml
daemon:
  callback: >-
    pueue-discord-notify
    --webhook-file /etc/pueue/discord-webhook.txt
    --id {{id}}
    --command "{{command}}"
    --result {{result}}
    --exit-code {{exit_code}}
  callback_log_lines: 10
```

Restart the daemon: `pueue daemon restart`.

### NixOS / home-manager

```nix
{ pkgs, lib, config, ... }:

{
  services.pueue = {
    enable = true;
    settings.daemon = {
      callback = lib.replaceString "\n" " " ''
        "${lib.getExe pkgs.pueue-discord-notify}"
        --webhook-file "${config.sops.secrets.pueue-discord-webhook.path}"
        --id "{{id}}"
        --command "{{command}}"
        --result "{{result}}"
        --exit-code "{{exit_code}}"
        --group "{{group}}"
      '';
      callback_log_lines = 10;
    };
  };

  sops.secrets.pueue-discord-webhook = {
    key = "pueue-discord-webhook";
  };
}
```

The callback string uses `lib.replaceString` to collapse the multi-line
expression into a single line (pueue expects a single-line callback command).

The webhook URL is stored encrypted in a [sops](https://github.com/getsops/sops)
secrets file:

```yaml
pueue-discord-webhook: ENC[AES256_GCM,data:...,type:str]
```

## Install

### Flake input

```nix
inputs.pueue-discord-notify.url = "github:takata/pueue-discord-notify";
```

### Overlay package (from GitHub)

```nix
buildGoModule {
  src = fetchFromGitHub {
    owner = "takata";
    repo = "pueue-discord-notify";
    rev = "...";
    hash = "...";
  };
  vendorHash = null;
  doCheck = true;
  checkPhase = ''
    runHook preCheck
    go test -v ./...
    runHook postCheck
  '';
}
```

### Build from source

```shell
go build -o pueue-discord-notify .
```
