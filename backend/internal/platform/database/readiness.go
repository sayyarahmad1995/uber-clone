package database

import "context"

type Pinger interface {
	PingContext(context.Context) error
}

type ReadinessChecker struct {
	pinger Pinger
}

func NewReadinessChecker(pinger Pinger) ReadinessChecker {
	return ReadinessChecker{pinger: pinger}
}

func (c ReadinessChecker) Check(ctx context.Context) error {
	return c.pinger.PingContext(ctx)
}
