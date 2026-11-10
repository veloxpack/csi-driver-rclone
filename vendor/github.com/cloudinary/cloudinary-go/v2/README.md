[![Tests](https://github.com/cloudinary/cloudinary-go/actions/workflows/test.yaml/badge.svg)](https://github.com/cloudinary/cloudinary-go/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudinary/cloudinary-go/v2)](https://goreportcard.com/report/github.com/cloudinary/cloudinary-go/v2)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/cloudinary/cloudinary-go/v2)](https://pkg.go.dev/github.com/cloudinary/cloudinary-go/v2)
[![License](https://img.shields.io/github/license/cloudinary/cloudinary-go.svg)](LICENSE)

# Cloudinary Go SDK

Upload, transform, optimize, and manage images and videos with Cloudinary from Go — the `cloudinary-go` module.

## Install

```bash
go get github.com/cloudinary/cloudinary-go/v2
```

## Quick start

Set your API environment variable (Console > Settings > API Keys):

```bash
export CLOUDINARY_URL=cloudinary://<api_key>:<api_secret>@<cloud_name>
```

Upload an image and get an optimized delivery URL:

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Quick start failed:", err)
		fmt.Fprintln(os.Stderr, "Check that CLOUDINARY_URL is set (Console > Settings > API Keys).")
		os.Exit(1)
	}
}

func run() error {
	cld, err := cloudinary.New()
	if err != nil {
		return err
	}
	ctx := context.Background()

	// Upload a remote image (a local file path works the same way).
	result, err := cld.Upload.Upload(ctx,
		"https://res.cloudinary.com/demo/image/upload/sample.jpg",
		uploader.UploadParams{PublicID: "quickstart-sample"})
	if err != nil {
		return err
	}
	if result.Error.Message != "" {
		// Cloudinary rejected the request; this arrives with err == nil.
		return fmt.Errorf("upload rejected: %s", result.Error.Message)
	}
	fmt.Println("Uploaded:", result.PublicID)

	// Build a 400x400 auto-cropped URL with automatic format and quality.
	image, err := cld.Image(result.PublicID)
	if err != nil {
		return err
	}
	image.Transformation = "c_fill,g_auto,h_400,w_400/f_auto,q_auto"
	url, err := image.String()
	if err != nil {
		return err
	}
	fmt.Println("Optimized URL:", url)
	return nil
}
```

Save as `quickstart.go` and run `go run quickstart.go`. [Create a free account](https://cloudinary.com/users/register_free) if you don't have one — or run `npx @cloudinary/cloud` to [provision one without signing up](docs/get-credentials.md).

Note the two checks in `run`: `err` reports transport, context, and decoding failures, while a Cloudinary rejection arrives with `err == nil` and a populated `result.Error.Message`. See [Handle errors](docs/handle-errors.md).

## Common tasks

- [Get Cloudinary credentials](docs/get-credentials.md)
- [Import and call the SDK](docs/import-and-call.md)
- [Configure Cloudinary](docs/configure-cloudinary.md)
- [Upload an image](docs/upload-image.md)
- [Upload a large video](docs/upload-large-video.md)
- [Sign a browser upload](docs/sign-browser-upload.md)
- [Transform and deliver an image](docs/transform-and-deliver-image.md)
- [Transform and deliver a video](docs/transform-and-deliver-video.md)
- [Search and manage assets](docs/search-and-manage-assets.md)
- [Moderate an upload](docs/moderate-upload.md)
- [Use structured metadata](docs/use-structured-metadata.md)
- [Serve uploads over HTTP](docs/serve-uploads-over-http.md)
- [Handle errors](docs/handle-errors.md)
- [Troubleshoot errors](docs/troubleshoot-errors.md)

Runnable versions live in [`examples/`](examples/) — each is a complete program you can run directly. It is a nested module, so run them from inside `examples/`.

## When to use this SDK

Use this module in **Go server-side code**: uploads, signed operations, asset administration, search, moderation, and delivery URL generation.

For other jobs, better-fitting tools exist:

- Browser or frontend framework rendering: the [frontend SDKs](https://cloudinary.com/documentation/frontend_sdks) ([md](https://cloudinary.com/documentation/frontend_sdks.md)) — this module generates URLs, not markup.
- Complete in-browser upload UI: [Upload Widget](https://cloudinary.com/documentation/upload_widget) ([md](https://cloudinary.com/documentation/upload_widget.md)), signed from Go with [Sign a browser upload](docs/sign-browser-upload.md).
- Video playback UI: [Cloudinary Video Player](https://cloudinary.com/documentation/cloudinary_video_player) ([md](https://cloudinary.com/documentation/cloudinary_video_player.md)).
- Text-to-image generation and image-to-video: [platform APIs](https://cloudinary.com/documentation/image_generation_addon) ([md](https://cloudinary.com/documentation/image_generation_addon.md)), not wrapped by this module.
- Account and sub-account provisioning: [Provisioning API](https://cloudinary.com/documentation/provisioning_api) ([md](https://cloudinary.com/documentation/provisioning_api.md)) over HTTP.
- Multi-step media workflow automation: [MediaFlows](https://cloudinary.com/documentation/mediaflows_user_guide) ([md](https://cloudinary.com/documentation/mediaflows_user_guide.md)).
- Interactive agent-driven asset operations: [Cloudinary MCP servers and Skills](https://cloudinary.com/documentation/cloudinary_llm_mcp) ([md](https://cloudinary.com/documentation/cloudinary_llm_mcp.md)).

The full capability map — plus the Skills, MCP servers, and CLI worth setting up first — is in [docs/platform-capabilities.md](docs/platform-capabilities.md).

## Status and compatibility

Stable, actively maintained. See [CHANGELOG.md](CHANGELOG.md).

| SDK version | Go 1.13 - 1.19 | Go 1.20 - 1.23 | Go 1.24 - 1.27 |
|-------------|----------------|----------------|----------------|
| 2.8 and up  | ❌             | ✔️             | ✔️             |
| 2.7         | ✔️             | ✔️             | ✔️             |
| 1.x         | ✔️             | ✔️             | ✔️             |

## Documentation

- [Bundled task docs](docs/README.md) — ship inside the module, version-matched.
- [Go SDK guide](https://cloudinary.com/documentation/go_integration) — the full documentation ([md](https://cloudinary.com/documentation/go_integration.md)).
- [Logging](logger/README.md) — redefining the logger and adjusting the log level.

Documentation links in this README point at the browsable HTML page, with an `(md)` companion link that returns the same page as raw Markdown. Inside `docs/` and `examples/` the links are Markdown-only, since those files are written to be read by coding agents. Either form works for any page: add `.md` for Markdown, drop it for HTML.

## For AI coding agents

- Contributing to this repo: read [AGENTS.md](AGENTS.md).
- Using the installed module: the docs bundled in the module match your resolved version and
  are the source of truth; start with
  [platform-capabilities](docs/platform-capabilities.md) before assuming a feature exists.

Go has no fixed install path, so locate the bundled docs with:

```bash
go list -m -f '{{.Dir}}' github.com/cloudinary/cloudinary-go/v2
# then read <that path>/docs/README.md
```

## Support

- SDK bugs and feature requests: [GitHub issues](https://github.com/cloudinary/cloudinary-go/issues)
- Account and platform questions: [Cloudinary support](https://support.cloudinary.com)

Contributing: see [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md) for private vulnerability reporting. Keep your `api_secret` in server-side code; for client uploads, use the server-signed pattern in [Sign a browser upload](docs/sign-browser-upload.md).

## License

Released under the MIT license — see [LICENSE](LICENSE). Copyright (c) Cloudinary Ltd.
