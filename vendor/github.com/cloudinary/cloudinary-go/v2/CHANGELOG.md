2.13.0 / 2025-08-20
==================

New Functionality And Features
------------------------------

  * Add support for `Fields` in `AutoVideoDetails` upload parameter

2.12.0 / 2025-08-11
==================

New Functionality And Features
------------------------------

  * Add support for `AutoVideoDetails` upload parameter

2.11.1 / 2025-08-07
==================

  * Fix handling of multiple formats of contextual metadata

2.11.0 / 2025-07-16
==================

New Functionality And Features
------------------------------

  * Add support for `AutoChaptering` and `AutoTranscript` upload parameters

2.10.1 / 2025-06-17
==================

  * Fix API parameters signature

2.10.0 / 2025-05-14
==================

New Functionality And Features
------------------------------

  * Add support for `Context` in `AssetResult`

Other Changes
-------------

  * Add advanced examples

2.9.1 / 2025-01-29
==================

  * Fix possible race condition when generating analytics token

2.9.0 / 2024-08-18
==================

New Functionality And Features
------------------------------

  * Add support for `RenameFolder` Admin API
  * Add support for `GetConfig` Admin API

2.8.0 / 2024-07-30
==================

New Functionality And Features
------------------------------

  * Add support for `AccessControl` in `uploader.UploadParams`
  * Add support for `Analyze` API

Other Changes
-------------

  * Bump `go` version to `1.20`
  * Bump vulnerable dependencies

2.7.0 / 2024-01-08
==================

New Functionality And Features
------------------------------

  * Add support for uploading `multipart.FileHeader`
  * Add support for `Fields` parameter in Search and Admin APIs

Other Changes
-------------

    * Reduce memory consumption in `Upload` API

2.6.2 / 2023-11-30
==================

  * Add support for uploading large `io.SectionReader` type

2.6.1 / 2023-11-30
==================

  * Add support for uploading large `os.File` type

2.6.0 / 2023-11-01
==================

New Functionality And Features
------------------------------

  * Add support for `ImageFile` in `VisualSearch` Admin API
  * Add support for `PlaybackURL` in `UploadResult` on video upload
  * Add support for `Context` field in `SearchResult`
  * Add support for configured signature algorithm in Upload API


2.5.1 / 2023-09-20
==================

  * Fix `AuthToken` for shared CDN configuration

2.5.0 / 2023-09-19
==================

New Functionality And Features
------------------------------

  * Add support for `OnSuccess` upload parameter

Other Changes
-------------

  * Fix `AuthToken` generation without ACL

2.4.0 / 2023-07-31
==================

New Functionality And Features
------------------------------

  * Add support for `SearchFolders` API
  * Add support for `VisualSearch` Admin API
  * Add support for `MediaMetadata` API parameter

2.3.0 / 2023-06-19
==================

New Functionality And Features
------------------------------

  * Add support for Search URL
  * Add support for related assets APIs
  * Add support for `LastUpdated` field in `AssetResult`
  * Add support for `NextCursor` in `GetTransformationResult`

Other Changes
-------------

  * Fix race condition in analytics token generation
  * Clarify error message when unsupported file parameter is provided

2.2.0 / 2022-09-14
==================

New Functionality And Features
------------------------------

  * Add support for `AssetsByAssetFolder` Admin API

Other Changes
-------------

  * Fix support for multiple `PublicIDs` in `CreateArchive`
  * Fix support for custom `timestamp` in `api.SignParameters`

2.1.0 / 2022-07-07
==================

New Functionality And Features
------------------------------

  * Add support for `Phash` in `UploadResult`

2.0.2 / 2022-06-16
==================

  * Fix parameters alignment

2.0.1 / 2022-05-23
==================

  * Fix module path

2.0.0 / 2022-05-11
==================

Breaking Changes
----------------
  * Fix boolean values handling in API functions 
  * Align parameter names

New Functionality And Features
------------------------------
  * Add support for folder decoupling

Other Changes
-------------
  * Fix `golint` issues

1.7.0 / 2022-03-24
==================

New Functionality And Features
------------------------------

  * Add support for `Response` field for future proofing
  * Add support for `ReorderMetadataFields` Admin API
  * Add support for `AssetByAssetID` Admin API
  * Add support for `Credits` field in `UsageResult`
  * Add support for `AuthToken`

Other Changes
-------------

  * Fix `Metadata` field  in `UploadParams` serialization
  * Update `README`

1.6.0 / 2022-01-17
==================

New Functionality And Features
------------------------------

  * Add support for `OAuth` authorization
  * Add support for `UserPlatform`

1.5.0 / 2021-12-15
==================

New Functionality And Features
------------------------------
  * Add support for `ResponsiveBreakpoints` upload parameter
 
Other Changes
-------------
  * Fix upload parameters signature
  * Fix multiple `PublicIDs` support in `Tags` Upload API
  * Fix `UpdateUploadPresetParams`
  * Make API client public

1.4.0 / 2021-11-18
==================

New Functionality And Features
------------------------------
  * Add support for `Moderation` in `UploadResult`

Other Changes
-------------
  * Fix `UpdateMetadata` Upload API parameters

1.3.0 / 2021-08-20
==================

New Functionality And Features
------------------------------
  * Add support for query parameters in `CLOUDINARY_URL`
  * Add support for `DeliveryType` in `Assets` Admin API
  * Add support for `ReorderMetadataFieldDatasource` Admin API
  * Add `Info` field to `AssetResult`

Other Changes
-------------
* Add code generation for typed setters
* Add support for acceptance tests

1.2.0 / 2021-06-24
==================

New Functionality And Features
------------------------------
  * Add support for URL generation
  * Add missing `NextCursor` to `DeleteAssetsResult`

Other Changes
-------------
  * Fix request signature of the Upload API

1.1.0 / 2021-05-13
==================

New Functionality And Features
------------------------------
  * Add `Logger`
  * Add `NextCursor` field to search API result

Other Changes
-------------
  * Fix Upload with Context

1.0.0 / 2021-04-06
==================

Breaking Changes
------------------------------
  * Rename `Create()` to `New()`
  * Rename `NewFromUrl` to `NewFromURL`
  * Rename `Account` config to `Cloud`
  * Apply `golint` fixes
  * Unexport redundant public constants
  * Refactor Search Api

New Functionality And Features
------------------------------
  * Add `PrivateDownloadURL` helper
  * Add support for `Eager` upload transformations
  * Add support for large file upload
  * Add SDK usage example
  * Add `CONTRIBUTING.md`
    
Other Changes
-------------
  * Improve error handling in Configuration
  * Fix `Config.Api.Timeout` type
  * Fix api call timeout handling
  * Fix formatting
  * Update README
  * Add travis support
  * Refactor tests
  * Code cleanup
  * Add a link to the package documentation
  * Update README code samples


0.2.0 / 2021-01-05
==================

  * Fix package name
  * Reorganise files
  * Cleanup files

0.1.0 / 2021-01-04
========================

  * The first public beta
