# Security Policy

## Reporting a vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Report them privately through either channel:

- **GitHub private vulnerability reporting** — use the
  [Report a vulnerability](https://github.com/cloudinary/cloudinary-go/security/advisories/new)
  button on the Security tab of this repository.
- **Email** — [security@cloudinary.com](mailto:security@cloudinary.com).

Please include:

- the SDK version (`const Version` in `api/api.go`) and your Go version;
- a description of the issue and its impact;
- steps to reproduce, ideally a minimal program;
- any suggested remediation.

You will receive an acknowledgement, and we will keep you informed as we investigate and
prepare a fix. Please give us a reasonable opportunity to release one before any public
disclosure.

## Supported versions

Security fixes are applied to the latest `v2` minor release. See
[Version Support](README.md#Version-Support) for the supported Go versions.

## Keeping credentials safe

This is a **server-side** SDK and it holds your API secret. A few rules that prevent the
most common exposures:

- **Never ship the API secret to a browser or mobile client.** Sign uploads on your server
  instead — see [Sign a browser upload](docs/sign-browser-upload.md).
- **Do not log configuration or raw API responses.** `cld.Config` contains the API secret,
  and a result's `Response` field contains your `api_key`. Log `result.Error.Message`.
- **Keep `CLOUDINARY_URL` out of version control.** It embeds both key and secret; use
  environment variables or a secret manager.
- **Verify webhooks before acting on them** with
  `cld.Upload.VerifyNotificationSignature` — notification endpoints are publicly reachable.

If you believe a credential has been exposed, rotate it in the Cloudinary Console under
Settings > API Keys.
