package kratos

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestKratosIdentitySchemaSupportsPhoneWithoutEmail(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "identity", "kratos", "identity.schema.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var root struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatal(err)
	}
	var traits struct {
		Properties map[string]struct {
			Pattern string `json:"pattern"`
			Kratos struct {
				Credentials struct {
					Password struct {
						Identifier bool `json:"identifier"`
					} `json:"password"`
				} `json:"credentials"`
				Verification map[string]any `json:"verification"`
			} `json:"ory.sh/kratos"`
		} `json:"properties"`
		AnyOf []struct {
			Required []string `json:"required"`
		} `json:"anyOf"`
	}
	if err := json.Unmarshal(root.Properties["traits"], &traits); err != nil {
		t.Fatal(err)
	}
	phone, ok := traits.Properties["phone"]
	if !ok || phone.Pattern != "^[+][1-9][0-9]{7,14}$" || !phone.Kratos.Credentials.Password.Identifier {
		t.Fatal("phone must be an E.164 password identifier")
	}
	if len(phone.Kratos.Verification) != 0 {
		t.Fatal("Kratos must not claim SMS verification without a configured channel")
	}
	email, ok := traits.Properties["email"]
	if !ok || !email.Kratos.Credentials.Password.Identifier || email.Kratos.Verification["via"] != "email" {
		t.Fatal("legacy email password identifier and verification must remain")
	}
	var phoneOnly, emailOnly bool
	for _, option := range traits.AnyOf {
		if len(option.Required) != 1 {
			continue
		}
		switch option.Required[0] {
		case "phone":
			phoneOnly = true
		case "email":
			emailOnly = true
		}
	}
	if !phoneOnly || !emailOnly {
		t.Fatal("schema must accept either phone-only or email-only traits")
	}
}
