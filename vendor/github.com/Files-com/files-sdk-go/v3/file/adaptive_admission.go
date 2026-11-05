package file

import (
	"context"
	"time"

	"github.com/Files-com/files-sdk-go/v3/file/manager"
)

// adaptiveFileAdmissionRecheckInterval bounds how long admission waits before
// re-reading the learned part-concurrency target, which can grow without any
// file completing.
const adaptiveFileAdmissionRecheckInterval = time.Second

// adaptivePartTargets reports the learned part-concurrency target for the
// adaptive work this admission gate is protecting. ok is false until a shared
// manager exists.
type adaptivePartTargets interface {
	admissionTarget() (target int, ok bool)
}

// fileAdmissionPool is the subset of the FilesManager pool the admission gate
// reads. The pool's own Wait remains the static admission ceiling.
type fileAdmissionPool interface {
	RunningCount() int
	WaitForADoneWithContext(ctx context.Context) bool
}

func adaptiveFileAdmissionInitialTarget() int {
	return manager.AdaptiveFileAdmissionFloor
}

// waitForAdaptiveFileAdmission keeps the number of in-flight files near the
// learned part-concurrency target so a constrained connection runs a few
// files to completion instead of starving every admitted file. It blocks
// while the in-flight count is at or above max(floor, learned target) and
// returns false only when ctx ends. Before any shared adaptive manager
// exists the caller-provided initial target is used so the static pool is not
// filled before the adaptive manager has a chance to learn.
func waitForAdaptiveFileAdmission(ctx context.Context, pool fileAdmissionPool, targets adaptivePartTargets, initialTarget int) bool {
	for {
		limit, ok := adaptiveFileAdmissionLimit(targets, initialTarget)
		if !ok || pool.RunningCount() < limit {
			return ctx.Err() == nil
		}
		waitCtx, cancel := context.WithTimeout(ctx, adaptiveFileAdmissionRecheckInterval)
		pool.WaitForADoneWithContext(waitCtx)
		cancel()
		if ctx.Err() != nil {
			return false
		}
	}
}

// adaptiveFileAdmissionLimit allows the learned target plus growth headroom.
// The headroom (target/4 + 1) always covers the controller's next growth step
// (max(growStep, sqrt(target)), and t/4+1 >= sqrt(t) for all t), so workloads
// whose queued part demand equals the admitted file count — single-part files
// — still present enough excess demand for a growth probe to demonstrate a
// throughput gain. Admission equal to the target would starve the probe and
// pin ramp-up on fast connections.
func adaptiveFileAdmissionLimit(targets adaptivePartTargets, initialTarget int) (int, bool) {
	target, ok := targets.admissionTarget()
	if !ok {
		if initialTarget <= 0 {
			return 0, false
		}
		target = initialTarget
	}
	return max(target+target/4+1, manager.AdaptiveFileAdmissionFloor), true
}
