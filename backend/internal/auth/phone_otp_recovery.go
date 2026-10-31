package auth

import (
	"context"
	"errors"
)

// RecoverDue processes a bounded batch of due jobs. The caller controls the
// cadence and shutdown. Enable only with a provider that guarantees deduplication
// using challenge ID; do not spin or automatically retry without that guarantee.
func (w OTPDispatchWorker) RecoverDue(ctx context.Context, limit int64) (int, error) {
	ids, err := w.dispatch.Due(ctx, limit)
	if err != nil {
		return 0, err
	}
	processed := 0
	var firstErr error
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return processed, err
		}
		err := w.Deliver(ctx, id)
		if errors.Is(err, ErrOTPDispatchBusy) || errors.Is(err, ErrOTPDispatchMissing) {
			continue
		}
		if err != nil && firstErr == nil {
			firstErr = err
		}
		processed++
	}
	return processed, firstErr
}
