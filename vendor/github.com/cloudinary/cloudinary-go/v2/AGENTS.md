# Contributor guide for coding agents

This file is for agents **contributing to this repository**. If you are *using* the
installed `cloudinary-go` module in another project, read the bundled docs instead — they
ship inside the module and are version-matched to the code you resolved:

```bash
go list -m -f '{{.Dir}}' github.com/cloudinary/cloudinary-go/v2
# then look in <that path>/docs/
```

## Commands

```bash
go build ./...                  # build everything
go test ./...                   # unit, acceptance, and E2E tests
go test ./api/admin/            # one package
go vet ./...                    # vet
gofmt -l .                      # list unformatted files (no linter is configured)
make generate                   # regenerate parameter setters (see gen/)

bash scripts/get_test_cloud.sh  # allocate a throwaway test sub-account, prints a CLOUDINARY_URL
```

`go test ./...` **requires `CLOUDINARY_URL` in the environment**, including for the mocked
tests: `config` panics with a nil-pointer dereference without it, which looks like a code
bug but is a missing variable. Any syntactically valid value works for the mocked packages:

```bash
export CLOUDINARY_URL=cloudinary://key:secret@test-cloud
```

The E2E tests perform real network calls and need a real cloud. Use
`scripts/get_test_cloud.sh` for a throwaway sub-account rather than a personal or
production environment. Note that script uses `grep -oiP`, which is GNU-only and fails on
macOS BSD grep — it still prints the URL, but the version prefix is lost.

## Testing

- `*_test.go` — E2E tests that talk to the Cloudinary API. They need real credentials.
- `*_acceptance_test.go` — acceptance tests against a mocked HTTP server, asserting the
  shape of the outgoing request. See `TEST.md` for the table-driven case format.
- Do not add tests that consume paid add-ons without a skip guard.
- Nondeterministic AI output (captions, tags, moderation verdicts) must be asserted by
  request shape, state transition, and response schema — never by exact output values.
- Record base-vs-branch failure counts when changing anything; do not claim "tests pass".

## Project structure

- `cloudinary.go` — package entry point; `New`, `NewFromURL`, `NewFromParams`,
  `NewFromOAuthToken`, and the `Image`/`Video`/`File`/`Media` URL builders.
- `api/` — shared types, signing, and parameter serialization (`StructToParams`).
- `api/uploader/` — Upload API. Chunking for large files is handled transparently here.
- `api/admin/` — Admin API (~70 methods), plus `admin/search` and `admin/metadata`.
- `asset/` — URL construction for images, video, raw files, and search URLs.
- `config/` — configuration structs and `CLOUDINARY_URL` parsing.
- `logger/` — logging; overridable, see `logger/README.md`.
- `transformation/` — transformation types. These are string aliases, not a builder.
- `gen/generate_setters/` — code generator for parameter setters; run via `make generate`.
- `internal/` — signature helpers and shared test fixtures (`internal/cldtest`).
- `docs/` — version-matched Markdown docs, shipped inside the module.
- `examples/` — runnable task examples. A **nested module** with its own `go.mod`, so it is
  excluded from the parent module and from `go build ./...` at the root.
- `scripts/` — test-cloud allocation and version bumping.

## Code style

- Standard Go. `gofmt` is authoritative; **there is no linter configured in this repo, and
  you should not add one.** Run `gofmt -l .` and `go vet ./...` before opening a PR.
- Exported symbols carry doc comments, ending with a link to the relevant Cloudinary API
  reference page where one exists.
- Parameters are structs with `json` tags consumed by `api.StructToParams`. Optional
  booleans are `*bool` so that "false" is distinguishable from "unset"; use `api.Bool`.
- Every network method takes `ctx context.Context` first and returns `(*Result, error)`.

### The error convention

The SDK reports failures on two channels. `api.HandleRawResponse` populates the result's
`Response` field and returns only decoding errors; a Cloudinary rejection lands in
`result.Error.Message` with `err == nil`.

Preserve this when adding methods — every result struct needs an
`Error api.ErrorResp \`json:"error,omitempty"\`` field, so callers can read rejections
uniformly. Two older types (`uploader.UpdateMetadataResult`, `uploader.RenameResult`) type
the field as `interface{}`; prefer `api.ErrorResp` in new code.

This convention is stable and callers depend on it — see `docs/handle-errors.md`.

## Documentation

`docs/` ships inside the module because Go's module cache holds the whole repository at a
version. There is no manifest to update and **no version number in the docs** — the
version-matched guarantee comes from shipping in the module, so nothing needs bumping at
release time. Do not add a version stamp.

`examples/` is a nested module and therefore does **not** ship in the resolved module. Every
`docs/` page carries its complete flow inline for that reason. Keep it that way.

When you change a page, verify it rather than trusting the prose:

- **Every Go snippet must compile against the current SDK.** The fastest check is to paste
  it into a scratch `main.go` in a module that `replace`s this one, and build. This is what
  catches invented symbols and wrong field types — the failure mode these docs exist to
  prevent. Snippets that cannot compile standalone are marked `// illustrative`.
- **Build and vet the examples**: `cd examples && go build ./... && go vet ./...`.
- **Check relative links and heading anchors** still resolve after renaming a section.

Claims in `docs/` are expected to have been **verified by execution**, not inferred from
reading the source. If you change documented behaviour, re-run the relevant example against
a real cloud.

## Version

The version string lives in **`api/api.go`** as `const Version`. `scripts/get_test_cloud.sh`
greps it, and `scripts/update_version.sh` maintains it — do not reformat that line or move
the constant.

## Git workflow

- Branch from `main`; keep changes focused; one topic per pull request.
- Run `go build ./...`, `go vet ./...`, `gofmt -l .`, and `go test ./...` before opening a PR.
- Add new `CHANGELOG.md` entries at the top; do not rewrite published entries. Docs-only
  changes get no changelog entry.
- Never commit credentials, `.env` files, or downloaded media left behind by examples.

## Boundaries

**Always**
- Give every new result struct an `Error api.ErrorResp` field.
- Keep `docs/` and `examples/` consistent with the code they document.
- Keep API secrets and real cloud names out of docs, examples, tests, and fixtures.
- Use `scripts/get_test_cloud.sh` for live testing, never a production environment.

**Ask first**
- Changing the supported Go version in `go.mod`, or adding a dependency.
- Renaming or removing any exported symbol, or changing a params/result struct field.
- Changing the error-reporting convention described above.
- Adding a linter, formatter, or CI check.
- Changing release, CI, or publishing configuration.

**Never**
- Commit credentials or real account identifiers.
- Add a network call to an acceptance test.
- Document a Cloudinary platform capability as an SDK method unless this module implements
  it (see `docs/platform-capabilities.md`).
- Add a runnable example outside `examples/` — that directory is the single home for them.
