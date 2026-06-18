# AGENTS.md

## Build & test

```shell
nix build                                       # build + run tests
go test -v ./...                                # faster iteration (no nix)
```

Tests run automatically via `doCheck = true` in the Nix build.

## Nix source gotcha

`package.nix` uses `lib.fileset.toSource` with an explicit file list:

```nix
src = lib.fileset.toSource {
  root = ./.;
  fileset = lib.fileset.unions [ ./go.mod ./main.go ./main_test.go ];
};
```

New `.go` files must be added to that list AND `git add`-ed before Nix will pick them up (Nix reads from the git index, not the working tree).

## Project structure

- `main.go` — single binary, stdlib only (net/http, encoding/json, flag)
- `main_test.go` — tests use `httptest` to mock the Discord webhook
- `go.mod` — Go 1.22, zero external dependencies
- `package.nix` — `buildGoModule` with `vendorHash = null`

## Formatting

```shell
nix fmt                 # runs nixfmt-tree
```

## CLI flags

`--webhook-file` (required), `--id`, `--command`, `--result`, `--exit-code`, `--group`, `--host` (defaults to `os.Hostname()`).

## Pueue callback context

Callback template variables: `{{id}}`, `{{command}}`, `{{result}}`, `{{exit_code}}`, `{{group}}`, `{{path}}`, `{{start}}`, `{{end}}`, `{{output}}`, `{{output_path}}`, `{{queued_count}}`, `{{stashed_count}}`. No automatic CLI args — user builds the full command from the template.
