package file

import (
	"net/url"
	"strings"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	"github.com/Files-com/files-sdk-go/v3/lib"
)

const (
	uploadV2MiB = int64(1024 * 1024)
	uploadV2GiB = int64(1024) * uploadV2MiB
	uploadV2TiB = int64(1024) * uploadV2GiB

	uploadV2S3MinPartSize    = 5 * uploadV2MiB
	uploadV2S3MaxPartSize    = 5 * uploadV2GiB
	uploadV2S3MaxObjectSize  = 5 * uploadV2TiB
	uploadV2S3MaxPartCount   = int64(10_000)
	uploadV2UnknownGrowEvery = 128
)

const (
	uploadV2TargetS3      = TransferV2TargetS3
	uploadV2TargetDefault = TransferV2TargetDefault
	uploadV2TargetDirect  = TransferV2TargetDirect
)

type uploadV2PartPlan struct {
	target     TransferV2TargetClass
	totalSize  *int64
	partSize   int64
	unknownCap int64
	mode       string
}

func newUploadV2PartIterator(part files_sdk.FileUploadPart, totalSize *int64, off int64, index int) (Iterator, bool) {
	plan, ok, _ := newUploadV2PartPlanForUpload(part, totalSize)
	if !ok {
		return nil, false
	}

	return plan.resume(off, index), true
}

func newUploadV2PartPlanForUpload(part files_sdk.FileUploadPart, totalSize *int64, classifiers ...UploadV2TargetClassifier) (uploadV2PartPlan, bool, string) {
	if !lib.UnWrapBool(part.ParallelParts) {
		return uploadV2PartPlan{}, false, "parallel_parts_disabled"
	}

	target := classifyUploadV2Target(part, classifiers...)
	plan, ok, reason := newUploadV2PartPlan(target, totalSize)
	if !ok {
		return uploadV2PartPlan{}, false, reason
	}
	return plan, true, ""
}

func (u *uploadIO) newUploadV2PartPlanForUpload() (uploadV2PartPlan, bool, string) {
	plan, ok, reason := newUploadV2PartPlanForUpload(u.FileUploadPart, u.Size, u.uploadV2TargetClassifier)
	if !ok {
		return uploadV2PartPlan{}, false, reason
	}
	return plan.withTuning(u.uploadV2Tuning)
}

func uploadV2PartPlanEligible(part files_sdk.FileUploadPart, size int64, classifier UploadV2TargetClassifier, tuning UploadV2Tuning) bool {
	plan, ok, _ := newUploadV2PartPlanForUpload(part, &size, classifier)
	if !ok {
		return false
	}
	_, ok, _ = plan.withTuning(tuning)
	return ok
}

func (p uploadV2PartPlan) withTuning(tuning UploadV2Tuning) (uploadV2PartPlan, bool, string) {
	if p.target != uploadV2TargetS3 {
		return p, true, ""
	}

	if tuning.S3PartSizeMiB > 0 {
		if tuning.S3PartSizeMiB > uploadV2S3MaxPartSize/uploadV2MiB {
			return uploadV2PartPlan{}, false, "s3_tuned_part_too_large"
		}
		partSize := tuning.S3PartSizeMiB * uploadV2MiB
		if partSize < uploadV2S3MinPartSize {
			return uploadV2PartPlan{}, false, "s3_tuned_part_too_small"
		}
		if p.totalSize != nil && ceilDiv(*p.totalSize, partSize) > uploadV2S3MaxPartCount {
			return uploadV2PartPlan{}, false, "s3_tuned_part_too_small_for_size"
		}
		p.partSize = partSize
		if p.mode == "" {
			p.mode = "tuned"
		} else {
			p.mode += "_tuned"
		}
		return p, true, ""
	}

	if tuning.S3WorkloadBytes <= 0 || p.totalSize == nil {
		return p, true, ""
	}
	return p.withS3WorkloadTuning(tuning)
}

func (p uploadV2PartPlan) withS3WorkloadTuning(tuning UploadV2Tuning) (uploadV2PartPlan, bool, string) {
	targetParts := uploadV2S3WorkloadTargetPartCount(tuning)
	partSize := s3WorkloadPartSize(tuning.S3WorkloadBytes, p.partSize, targetParts, tuning)
	partSize = max(partSize, roundUpToMiB(ceilDiv(*p.totalSize, uploadV2S3MaxPartCount)))
	if partSize <= 0 || partSize >= p.partSize {
		return p, true, ""
	}
	if ceilDiv(*p.totalSize, partSize) > uploadV2S3MaxPartCount {
		return uploadV2PartPlan{}, false, "s3_workload_tuned_part_too_small_for_size"
	}
	p.partSize = partSize
	if p.mode == "" {
		p.mode = "workload_tuned"
	} else {
		p.mode += "_workload_tuned"
	}
	return p, true, ""
}

func newUploadV2PartPlan(target TransferV2TargetClass, totalSize *int64) (uploadV2PartPlan, bool, string) {
	target = normalizeTransferV2TargetClass(target)
	plan := uploadV2PartPlan{target: target, totalSize: totalSize}
	if totalSize == nil {
		if target == uploadV2TargetS3 {
			return uploadV2PartPlan{}, false, "s3_unknown_size"
		}
		plan.partSize = 8 * uploadV2MiB
		plan.mode = "unknown_size_growth"
		plan.unknownCap = 32 * uploadV2MiB
		return plan, true, ""
	}

	plan.mode = "known_size"
	switch target {
	case uploadV2TargetS3:
		if *totalSize > uploadV2S3MaxObjectSize {
			return uploadV2PartPlan{}, false, "s3_object_too_large"
		}
		plan.partSize = roundUpToMiB(max(
			s3KnownSizePreferredPartSize(*totalSize),
			uploadV2S3MinPartSize,
			ceilDiv(*totalSize, uploadV2S3MaxPartCount),
		))
		if plan.partSize > uploadV2S3MaxPartSize {
			return uploadV2PartPlan{}, false, "s3_part_too_large"
		}
	default:
		plan.partSize = defaultKnownSizePreferredPartSize(*totalSize)
	}

	if plan.partSize <= 0 {
		return uploadV2PartPlan{}, false, "invalid_part_size"
	}
	return plan, true, ""
}

func (p uploadV2PartPlan) resume(off int64, index int) Iterator {
	currentOff := off
	currentIndex := index
	exhausted := false
	var iterator Iterator
	iterator = func() (OffSet, Iterator, int) {
		if exhausted || p.done(currentOff, currentIndex) {
			return OffSet{}, nil, currentIndex
		}

		partIndex := currentIndex
		partSize := p.partSizeForIndex(partIndex)
		if p.totalSize != nil {
			remaining := *p.totalSize - currentOff
			if remaining < 0 {
				exhausted = true
				return OffSet{}, nil, partIndex
			}
			if remaining < partSize {
				partSize = remaining
			}
		}

		offset := OffSet{off: currentOff, len: partSize}
		currentOff += partSize
		currentIndex++
		if p.totalSize != nil && currentOff >= *p.totalSize {
			exhausted = true
			return offset, nil, partIndex
		}

		return offset, iterator, partIndex
	}
	return iterator
}

func (p uploadV2PartPlan) done(off int64, index int) bool {
	if p.totalSize == nil || off < *p.totalSize {
		return false
	}
	return *p.totalSize != 0 || off != 0 || index != 0
}

func s3KnownSizePreferredPartSize(totalSize int64) int64 {
	switch {
	case totalSize < uploadV2GiB:
		return 16 * uploadV2MiB
	case totalSize < 16*uploadV2GiB:
		return 32 * uploadV2MiB
	case totalSize < 512*uploadV2GiB:
		return 64 * uploadV2MiB
	case totalSize < uploadV2TiB:
		return 128 * uploadV2MiB
	case totalSize < 2*uploadV2TiB:
		return 256 * uploadV2MiB
	default:
		return 512 * uploadV2MiB
	}
}

func defaultKnownSizePreferredPartSize(totalSize int64) int64 {
	switch {
	case totalSize < 8*uploadV2GiB:
		return 16 * uploadV2MiB
	case totalSize < 128*uploadV2GiB:
		return 32 * uploadV2MiB
	default:
		return 64 * uploadV2MiB
	}
}

func (p uploadV2PartPlan) partSizeForIndex(index int) int64 {
	if p.totalSize != nil || p.unknownCap <= p.partSize {
		return p.partSize
	}

	partSize := p.partSize
	growEvery := uploadV2UnknownGrowEvery
	growthSteps := index / growEvery
	for i := 0; i < growthSteps && partSize < p.unknownCap; i++ {
		partSize *= 2
		if partSize > p.unknownCap {
			return p.unknownCap
		}
	}
	return partSize
}

func (p uploadV2PartPlan) estimatedPartCount() int {
	if p.totalSize == nil || p.partSize <= 0 {
		return 0
	}
	count := ceilDiv(*p.totalSize, p.partSize)
	if count <= 0 {
		return 1
	}
	maxInt := int64(^uint(0) >> 1)
	if count > maxInt {
		return 0
	}
	return int(count)
}

func uploadV2S3WorkloadTargetPartCount(tuning UploadV2Tuning) int64 {
	target := AdaptiveTransferS3InitialTarget
	if tuning.InitialTarget > 0 {
		target = tuning.InitialTarget
	}
	if tuning.S3InitialTarget > 0 {
		target = tuning.S3InitialTarget
	}
	multiplier := AdaptiveTransferS3WorkloadTargetPartMultiplier
	if tuning.S3WorkloadTargetPartMultiplier > 0 {
		multiplier = tuning.S3WorkloadTargetPartMultiplier
	}
	return int64(max(1, target) * max(1, multiplier))
}

func s3WorkloadPartSize(workloadBytes int64, currentPartSize int64, targetParts int64, tuning UploadV2Tuning) int64 {
	if workloadBytes <= 0 || currentPartSize <= 0 || targetParts <= 0 {
		return currentPartSize
	}
	if ceilDiv(workloadBytes, currentPartSize) >= targetParts {
		return currentPartSize
	}
	minPartSize := AdaptiveTransferS3WorkloadMinPartSizeMiB * uploadV2MiB
	if tuning.S3WorkloadMinPartSizeMiB > 0 {
		minPartSize = tuning.S3WorkloadMinPartSizeMiB * uploadV2MiB
	}
	partSize := floorPowerOfTwoMiB(ceilDiv(workloadBytes, targetParts))
	partSize = max(partSize, uploadV2S3MinPartSize, minPartSize)
	if partSize > currentPartSize {
		return currentPartSize
	}
	return partSize
}

func floorPowerOfTwoMiB(bytes int64) int64 {
	mib := bytes / uploadV2MiB
	if mib <= 1 {
		return uploadV2MiB
	}
	var power int64 = 1
	for power <= mib/2 {
		power *= 2
	}
	return power * uploadV2MiB
}

func classifyUploadV2Target(part files_sdk.FileUploadPart, classifiers ...UploadV2TargetClassifier) TransferV2TargetClass {
	if len(classifiers) > 0 && classifiers[0] != nil {
		return normalizeTransferV2TargetClass(classifiers[0](part))
	}

	parsed, err := url.Parse(part.UploadUri)
	if err != nil {
		return uploadV2TargetDefault
	}
	if files_sdk.DirectConnectionInfoPresent(part.DirectConnectionInfo) {
		return uploadV2TargetDirect
	}
	host := strings.ToLower(parsed.Hostname())

	if isS3UploadPartURL(parsed) || isS3UploadHost(host) {
		return uploadV2TargetS3
	}
	return uploadV2TargetDefault
}

func isS3UploadPartURL(parsed *url.URL) bool {
	query := parsed.Query()
	return query.Get("X-Amz-Algorithm") == "AWS4-HMAC-SHA256" &&
		query.Get("X-Amz-Credential") != "" &&
		query.Get("X-Amz-Signature") != "" &&
		query.Get("uploadId") != "" &&
		query.Get("partNumber") != ""
}

func isS3UploadHost(host string) bool {
	if host == "s3.amazonaws.com" ||
		strings.HasPrefix(host, "s3.") ||
		strings.HasPrefix(host, "s3-") ||
		strings.Contains(host, ".s3.") ||
		strings.Contains(host, ".s3-") {
		return strings.Contains(host, "amazonaws.com") || strings.Contains(host, "amazonaws.com.cn")
	}
	return false
}

func ceilDiv(n int64, d int64) int64 {
	if n <= 0 {
		return 0
	}
	return (n + d - 1) / d
}

func roundUpToMiB(n int64) int64 {
	return ceilDiv(n, uploadV2MiB) * uploadV2MiB
}
