package locationsearch

import "testing"

func TestValidPolicyBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name   string
		radius int64
		valid  bool
	}{
		{"minimum", MinNamedPlaceSnapRadiusMeters, true},
		{"default", DefaultNamedPlaceSnapRadiusMeters, true},
		{"maximum", MaxNamedPlaceSnapRadiusMeters, true},
		{"below minimum", MinNamedPlaceSnapRadiusMeters - 1, false},
		{"above maximum", MaxNamedPlaceSnapRadiusMeters + 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidPolicy(Policy{NamedPlaceSnapRadiusMeters: tc.radius})
			if got != tc.valid {
				t.Fatalf("ValidPolicy(%d)=%v want %v", tc.radius, got, tc.valid)
			}
		})
	}
}
