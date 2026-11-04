package native

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type memorySessions struct {
	mu       sync.Mutex
	sessions map[uuid.UUID]SessionRecord
	revoked  map[uuid.UUID]bool
}

func newMemorySessions() *memorySessions {
	return &memorySessions{sessions: make(map[uuid.UUID]SessionRecord), revoked: make(map[uuid.UUID]bool)}
}

func (s *memorySessions) InsertSession(_ context.Context, record SessionRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[record.ID] = record
	return nil
}

func (s *memorySessions) ResolveAccess(_ context.Context, id uuid.UUID, digest [32]byte, now time.Time) (uuid.UUID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, found := s.sessions[id]
	if !found || s.revoked[id] || record.AccessDigest != digest ||
		!record.AccessExpiresAt.After(now) || !record.RefreshExpiresAt.After(now) {
		return uuid.Nil, ErrInvalidSession
	}
	return record.UserID, nil
}

func (s *memorySessions) RotateRefresh(_ context.Context, id uuid.UUID, old, access, refresh [32]byte, nextAccess, now time.Time) (time.Time, time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, found := s.sessions[id]
	if !found || s.revoked[id] || record.RefreshDigest != old || !record.RefreshExpiresAt.After(now) {
		return time.Time{}, time.Time{}, ErrInvalidSession
	}
	record.AccessDigest = access
	record.RefreshDigest = refresh
	if nextAccess.After(record.RefreshExpiresAt) {
		nextAccess = record.RefreshExpiresAt
	}
	record.AccessExpiresAt = nextAccess
	s.sessions[id] = record
	return record.AccessExpiresAt, record.RefreshExpiresAt, nil
}

func (s *memorySessions) RevokeAccess(_ context.Context, id uuid.UUID, digest [32]byte, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, found := s.sessions[id]
	if !found || s.revoked[id] || record.AccessDigest != digest {
		return ErrInvalidSession
	}
	s.revoked[id] = true
	return nil
}

func (s *memorySessions) RevokeUserSessions(_ context.Context, userID uuid.UUID, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, record := range s.sessions {
		if record.UserID == userID {
			s.revoked[id] = true
		}
	}
	return nil
}

func TestNativeSessionIssueRotateRevoke(t *testing.T) {
	ctx := context.Background()
	store := newMemorySessions()
	manager, err := NewSessionManager(store)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }
	userID := uuid.New()
	first, err := manager.Issue(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if first.AccessToken == first.RefreshToken || first.AccessExpiresAt.Sub(now) != accessLifetime ||
		first.RefreshExpiresAt.Sub(now) != refreshLifetime {
		t.Fatal("bad token pair or lifetime")
	}
	if resolved, err := manager.Authenticate(ctx, first.AccessToken); err != nil || resolved != userID {
		t.Fatalf("authenticate returned %v, %v", resolved, err)
	}
	rotated, err := manager.Refresh(ctx, first.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if rotated.AccessToken == first.AccessToken || rotated.RefreshToken == first.RefreshToken ||
		!rotated.RefreshExpiresAt.Equal(first.RefreshExpiresAt) {
		t.Fatal("refresh must rotate secrets without extending absolute session lifetime")
	}
	if _, err := manager.Authenticate(ctx, first.AccessToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("old access secret was not invalidated: %v", err)
	}
	if _, err := manager.Refresh(ctx, first.RefreshToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("replayed refresh succeeded: %v", err)
	}
	if resolved, err := manager.Authenticate(ctx, rotated.AccessToken); err != nil || resolved != userID {
		t.Fatalf("rotated session invalid: %v, %v", resolved, err)
	}
	if err := manager.Logout(ctx, rotated.AccessToken); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Authenticate(ctx, rotated.AccessToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("logged-out access accepted: %v", err)
	}
	if _, err := manager.Refresh(ctx, rotated.RefreshToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("logged-out refresh accepted: %v", err)
	}
}

func TestNativeSessionExpiryAndAccountWideRevoke(t *testing.T) {
	ctx := context.Background()
	manager, err := NewSessionManager(newMemorySessions())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }
	userID := uuid.New()
	first, err := manager.Issue(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	manager.now = func() time.Time { return now.Add(accessLifetime) }
	if _, err := manager.Authenticate(ctx, first.AccessToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("access survived expiry: %v", err)
	}
	next, err := manager.Refresh(ctx, first.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	manager.now = func() time.Time { return now.Add(refreshLifetime) }
	if _, err := manager.Refresh(ctx, next.RefreshToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("refresh survived absolute expiry: %v", err)
	}
	manager.now = func() time.Time { return now }
	second, err := manager.Issue(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.RevokeUser(ctx, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Authenticate(ctx, second.AccessToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("account-wide revocation failed: %v", err)
	}
}

func TestNativeConcurrentRefreshSingleWinner(t *testing.T) {
	ctx := context.Background()
	manager, err := NewSessionManager(newMemorySessions())
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := manager.Issue(ctx, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 12)
	for i := 0; i < cap(results); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := manager.Refresh(ctx, tokens.RefreshToken)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, ErrInvalidSession) {
			t.Fatalf("unexpected refresh error: %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly one refresh success, got %d", successes)
	}
}

func TestNativeSessionRejectsMalformedTokens(t *testing.T) {
	manager, err := NewSessionManager(newMemorySessions())
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"", "Bearer abc", uuid.NewString(), uuid.NewString() + ".malformed"} {
		if _, err := manager.Authenticate(context.Background(), token); !errors.Is(err, ErrInvalidSession) {
			t.Fatalf("malformed access token accepted: %q, %v", token, err)
		}
	}
	if _, err := NewSessionManager(nil); !errors.Is(err, ErrSessionStore) {
		t.Fatalf("missing store accepted: %v", err)
	}
}
