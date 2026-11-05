package file

import (
	"context"
	"io"
	"net/http"
	"sync"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/lib"
)

type directTransferDownloadSuppressor interface {
	directTransferDownloadAttemptAllowed() bool
	disableDirectTransferDownload(reason string, err error)
}

type directTransferDownloadSuppressorContextKey struct{}

func directTransferDownloadSuppressorFromContext(ctx context.Context) directTransferDownloadSuppressor {
	if ctx == nil {
		return nil
	}
	suppressor, _ := ctx.Value(directTransferDownloadSuppressorContextKey{}).(directTransferDownloadSuppressor)
	return suppressor
}

func directTransferDownloadFailureOptions(suppressor directTransferDownloadSuppressor, opts ...files_sdk.RequestResponseOption) []files_sdk.RequestResponseOption {
	if suppressor == nil {
		return opts
	}

	wrapped := make([]files_sdk.RequestResponseOption, 0, len(opts)+1)
	wrapped = append(wrapped, files_sdk.ResponseOption(func(response *http.Response) error {
		if response != nil && response.Body != nil {
			response.Body = &directTransferDownloadFailureBody{
				ReadCloser: response.Body,
				suppressor: suppressor,
			}
		}
		return nil
	}))
	return append(wrapped, opts...)
}

type directTransferDownloadFailureBody struct {
	io.ReadCloser
	suppressor directTransferDownloadSuppressor
	disable    sync.Once
}

func (b *directTransferDownloadFailureBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if err != nil && err != io.EOF {
		b.disableDirectTransferDownload("direct_body_read_failed", err)
	}
	return n, err
}

func (b *directTransferDownloadFailureBody) Close() error {
	err := b.ReadCloser.Close()
	if err != nil {
		b.disableDirectTransferDownload("direct_body_close_failed", err)
	}
	return err
}

func (b *directTransferDownloadFailureBody) disableDirectTransferDownload(reason string, err error) {
	if b.suppressor == nil || err == nil {
		return
	}
	b.disable.Do(func() {
		b.suppressor.disableDirectTransferDownload(reason, err)
	})
}

type directTransferUploadStats interface {
	recordDirectAttempt()
	recordDirectSuccess()
	recordDirectFailure()
	recordDirectDisabled()
}

type directTransferUploadOptions struct {
	partNumber int64
	canReplay  func() bool
	observer   func(lib.UploadRetryAttempt)
	stats      directTransferUploadStats
}

func (u *uploadIO) callRawWithDirectTransfer(ctx context.Context, info files_sdk.DirectConnectionInfo, params *files_sdk.CallParams, options directTransferUploadOptions) (*http.Response, bool) {
	if !u.directTransferUploadAvailable(info) {
		return nil, false
	}

	options.stats.recordDirectAttempt()
	directURL, directClient, err := files_sdk.DirectTransferRetryableClient(ctx, u.Config, info)
	if err != nil {
		options.stats.recordDirectFailure()
		if u.disableDirectTransfersForUpload("direct_client_unavailable", err, options.partNumber) {
			options.stats.recordDirectDisabled()
		}
		return nil, false
	}

	directParams := *params
	directParams.Uri = directURL
	directParams.Headers = files_sdk.DirectTransferRequestHeaders(params.Headers)
	directParams.Client = lib.UploadRetryableHttpWithObserver(directClient, options.canReplay, options.observer)

	res, err := files_sdk.CallRaw(&directParams)
	if err == nil {
		err = lib.ResponseErrors(res, files_sdk.APIError(), lib.S3XMLError, lib.NonOkError)
		if err == nil {
			options.stats.recordDirectSuccess()
			return res, true
		}
	}

	lib.CloseBody(res)
	options.stats.recordDirectFailure()
	if u.disableDirectTransfersForUpload("direct_request_failed", err, options.partNumber) {
		options.stats.recordDirectDisabled()
	}
	return nil, false
}
