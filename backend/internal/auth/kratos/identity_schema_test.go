package kratos

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Keep the phone identifier and optional email contract visible to tests
// without assuming that Kratos marks an SMS-verified phone as a verified email.
func TestKratosIdentitySchemaSupportsPhoneWithoutEmail(t *testing.T) {
	path:=filepath.Join("..","..","..","..","identity","kratos","identity.schema.json")
	raw,err:=os.ReadFile(path)
	if err!=nil{t.Fatal(err)}
	var schema struct{
		Properties map[string]struct{
			Properties map[string]struct{
				Pattern string `json:"pattern"`
				Extensions struct{
					Credentials struct{
						Password struct{
							Identifier bool `json:"identifier"`
						} `json:"password"`
					} `json:"credentials"`
					Verification map[string]any `json:"verification"`
				} `json:"ory.sh/kratos"`
			} `json:"properties"`
			AnyOf []struct{Required []string `json:"required"`} `json:"anyOf"`
		} `json:"traits"`
	} `json:"properties"`
	}
	if err:=json.Unmarshal(raw,&schema);err!=nil{t.Fatal(err)}
	traits:=schema.Properties["traits"]
	phone,ok:=traits.Properties["phone"]
	if !ok||phone.Pattern!="^[+][1-9][0-9]{7,14}$"||!phone.Extensions.Credentials.Password.Identifier{
		t.Fatal("phone must be an E.164 password identifier")
	}
	if len(phone.Extensions.Verification)!=0 {t.Fatal("Kratos must not claim to verify SMS without a configured channel")}
	email,ok:=traits.Properties["email"]
	if !ok||!email.Extensions.Credentials.Password.Identifier||email.Extensions.Verification["via"]!="email"{
		t.Fatal("legacy email password and verification semantics must remain")
	}
	var phoneOnly,emailOnly bool
	for _,option:=range traits.AnyOf{
		if len(option.Required)==1&&option.Required[0]=="phone"{phoneOnly=true}
		if len(option.Required)==1&&option.Required[0]=="email"{emailOnly=true}
	}
	if !phoneOnly||!emailOnly{t.Fatal("schema must accept either phone-only or email-only traits")}
}
