package locationsearch

import "testing"

func TestValidSearchPolicyBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name   string
		radius int64
		valid  bool
	}{
		{"minimum", MinAutocompleteRadiusMeters, true},
		{"maximum", MaxAutocompleteRadiusMeters, true},
		{"below minimum", MinAutocompleteRadiusMeters - 1, false},
		{"above maximum", MaxAutocompleteRadiusMeters + 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidSearchPolicy(SearchPolicy{
				AutocompleteRadiusMeters: tc.radius,
			}); got != tc.valid {
				t.Fatalf("ValidSearchPolicy()=%v want %v", got, tc.valid)
			}
		})
	}
}
