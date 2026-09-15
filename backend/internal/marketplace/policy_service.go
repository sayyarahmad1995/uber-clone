package marketplace

import (
	"context"
	"strings"
)

type TimingPolicyStore interface {
	Load(context.Context) (TimingPolicy, error)
	Update(context.Context, TimingPolicy, string) (TimingPolicy, error)
}

type PostgresTimingPolicyStore struct {
	db QueryRower
}

func NewPostgresTimingPolicyStore(db QueryRower) PostgresTimingPolicyStore {
	return PostgresTimingPolicyStore{db: db}
}

func (s PostgresTimingPolicyStore) Load(ctx context.Context) (TimingPolicy, error) {
	return LoadTimingPolicy(ctx, s.db)
}

func (s PostgresTimingPolicyStore) Update(ctx context.Context, policy TimingPolicy, actor string) (TimingPolicy, error) {
	return UpdateTimingPolicy(ctx, s.db, policy, actor)
}

type PolicyService struct {
	store TimingPolicyStore
}

func NewPolicyService(store TimingPolicyStore) PolicyService {
	return PolicyService{store: store}
}

func (s PolicyService) Load(ctx context.Context) (TimingPolicy, error) {
	return s.store.Load(ctx)
}

func (s PolicyService) Update(ctx context.Context, policy TimingPolicy, actor string) (TimingPolicy, error) {
	actor = strings.TrimSpace(actor)
	if actor == "" || !ValidTimingPolicy(policy) {
		return TimingPolicy{}, ErrInvalidTimingPolicy
	}
	return s.store.Update(ctx, policy, actor)
}
