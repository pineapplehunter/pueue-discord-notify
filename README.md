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

## Install

### Nix

```shell
nix run github:your-user/pueue-discord-notify -- --help
```

Or add to your flake inputs:

```nix
inputs.pueue-discord-notify.url = "github:your-user/pueue-discord-notify";
```

### Build from source

```shell
go build -o pueue-discord-notify .
```
