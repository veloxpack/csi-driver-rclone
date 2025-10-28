[![Tests](https://github.com/cloudinary/cloudinary-go/actions/workflows/test.yaml/badge.svg)](https://github.com/cloudinary/cloudinary-go/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudinary/cloudinary-go/v2)](https://goreportcard.com/report/github.com/cloudinary/cloudinary-go/v2)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/cloudinary/cloudinary-go/v2)](https://pkg.go.dev/github.com/cloudinary/cloudinary-go/v2)

Cloudinary Go SDK
==================

## About

The Cloudinary Go SDK allows you to quickly and easily integrate your application with Cloudinary.
Effortlessly optimize, transform, upload and manage your cloud's assets.

#### Note

This Readme provides basic installation and usage information.
For the complete documentation, see the [Go SDK Guide](https://cloudinary.com/documentation/go_integration).

## Table of Contents

- [Key Features](#key-features)
- [Version Support](#Version-Support)
- [Installation](#installation)
- [Usage](#usage)
    - [Setup](#Setup)
    - [Transform and Optimize Assets](#Transform-and-Optimize-Assets)

## Key Features

- [Transform](https://cloudinary.com/documentation/go_media_transformations) assets.
- [Asset Management](https://cloudinary.com/documentation/go_asset_administration).
- [Secure URLs](https://cloudinary.com/documentation/video_manipulation_and_delivery#generating_secure_https_urls_using_sdks).

## Version Support

| **SDK Version** | **Go 1.13 - 1.19** | **Go 1.20** | **Go 1.21** | **Go 1.22** | **Go 1.23** |
|-----------------|--------------------|-------------|-------------|-------------|-------------|
| **2.8 & Up**    | ❌                  | ✔️          | ✔️          | ✔️          | ✔️          |
| **2.7**         | ✔️                 | ✔️          | ✔️          | ✔️          | ✔️          |
| **1.x**         | ✔️                 | ✔️          | ✔️          | ✔️          | ✔️          |

## Installation

```bash
go get github.com/cloudinary/cloudinary-go/v2
```

# Usage

### Setup

```go
import (
    "github.com/cloudinary/cloudinary-go/v2"
)

cld, _ := cloudinary.New()
```

- [See full documentation](https://cloudinary.com/documentation/go_integration#configuration).

### Transform and Optimize Assets

- [See full documentation](https://cloudinary.com/documentation/go_media_transformations).

```go
image, err := cld.Image("sample.jpg")
if err != nil {...}

image.Transformation = "c_fill,h_150,w_100"

imageURL, err := image.String()
```

### Upload

- [See full documentation](https://cloudinary.com/documentation/go_image_and_video_upload).
- [Learn more about configuring your uploads with upload presets](https://cloudinary.com/documentation/upload_presets).

```go
resp, err := cld.Upload.Upload(ctx, "my_picture.jpg", uploader.UploadParams{})
```

### Security options

- [See full documentation](https://cloudinary.com/documentation/solution_overview#security).

### Logging

Cloudinary SDK logs errors using standard `go log` functions.

For details on redefining the logger or adjusting the logging level, see [Logging](logger/README.md).

### Complete SDK Example

See [Complete SDK Example](example/example.go).

## Contributions

- Ensure tests run locally
- Open a PR and ensure Travis tests pass
- For more information on how to contribute, take a look at the [contributing](CONTRIBUTING.md) page.

## Get Help

If you run into an issue or have a question, you can either:

- Issues related to the SDK: [Open a GitHub issue](https://github.com/cloudinary/cloudinary-go/issues).
- Issues related to your account: [Open a support ticket](https://cloudinary.com/contact)

## About Cloudinary

Cloudinary is a powerful media API for websites and mobile apps alike, Cloudinary enables developers to efficiently
manage, transform, optimize, and deliver images and videos through multiple CDNs. Ultimately, viewers enjoy responsive
and personalized visual-media experiences—irrespective of the viewing device.

## Additional Resources

- [Cloudinary Transformation and REST API References](https://cloudinary.com/documentation/cloudinary_references):
  Comprehensive references, including syntax and examples for all SDKs.
- [MediaJams.dev](https://mediajams.dev/): Bite-size use-case tutorials written by and for Cloudinary Developers
- [DevJams](https://www.youtube.com/playlist?list=PL8dVGjLA2oMr09amgERARsZyrOz_sPvqw): Cloudinary developer podcasts on
  YouTube.
- [Cloudinary Academy](https://training.cloudinary.com/): Free self-paced courses, instructor-led virtual courses, and
  on-site courses.
- [Code Explorers and Feature Demos](https://cloudinary.com/documentation/code_explorers_demos_index): A one-stop shop
  for all code explorers, Postman collections, and feature demos found in the docs.
- [Cloudinary Roadmap](https://cloudinary.com/roadmap): Your chance to follow, vote, or suggest what Cloudinary should
  develop next.
- [Cloudinary Facebook Community](https://www.facebook.com/groups/CloudinaryCommunity): Learn from and offer help to
  other Cloudinary developers.
- [Cloudinary Account Registration](https://cloudinary.com/users/register/free): Free Cloudinary account registration.
- [Cloudinary Website](https://cloudinary.com): Learn about Cloudinary's products, partners, customers, pricing, and
  more.

## Licence

Released under the MIT license.
