package marketplace

import "context"

// ExpiryService is the application-owned entry point for materializing
// deadline-driven marketplace state. Business operations must continue to
// enforce persisted deadlines independently of when Sweep runs.
type ExpiryService struct {
	executor Execer
}

func NewExpiryService(executor Execer) ExpiryService {
	return ExpiryService{executor: executor}
}

func (s ExpiryService) Sweep(ctx context.Context) error {
	return Expire(ctx, s.executor)
}
