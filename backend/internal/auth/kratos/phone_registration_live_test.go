package kratos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestLiveKratosPhoneRegistration is opt-in, writes one isolated identity, and
// removes it through the admin API. Run against a disposable Kratos database.
// It tests actual Kratos schema behavior, unlike the httptest-based unit tests.
func TestLiveKratosPhoneRegistration(t *testing.T) {
	publicURL := os.Getenv("TEST_KRATOS_PUBLIC_URL")
	adminURL := os.Getenv("TEST_KRATOS_ADMIN_URL")
	if publicURL == "" || adminURL == "" {
		t.Skip("requires TEST_KRATOS_PUBLIC_URL and TEST_KRATOS_ADMIN_URL")
	}
	provider, err := New(publicURL, adminURL)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Unique valid E.164 identifier within the disposable Kratos test DB.
	id := uuid.New()
	phone := fmt.Sprintf("+%d", 900000000000000+(int64(id[0])<<16)+(int64(id[1])<<8)+int64(id[2]))
	// Normalize to valid E.164 length and avoid potentially live phone allocations:
	// no SMS is sent by this test.
	created, err := provider.RegisterPhoneIdentity(ctx, phone, "Kratos-Phone-Test-Password-123!")
	if err != nil {
		t.Fatalf("live phone registration rejected (%s): %v", phone, err)
	}
	if created.IdentitySubject == "" {
		t.Fatal("registration returned no identity subject")
	}
	t.Cleanup(func() {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, provider.adminURL+"/admin/identities/"+created.IdentitySubject, nil)
		if err != nil {
			t.Errorf("cleanup request: %v", err)
			return
		}
		resp, err := provider.client.Do(req)
		if err != nil {
			t.Errorf("cleanup failed: %v", err)
			return
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			t.Errorf("cleanup returned %s", resp.Status)
		}
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.adminURL+"/admin/identities/"+created.IdentitySubject, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := provider.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("identity lookup: %s", resp.Status)
	}
	var result struct {
		ID     string `json:"id"`
		Traits struct {
			Phone string  `json:"phone"`
			Email *string `json:"email"`
		} `json:"traits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.ID != created.IdentitySubject || result.Traits.Phone != phone || result.Traits.Email != nil {
		t.Fatalf("unexpected identity phone-only traits: %+v", result)
	}
}
