package marketplace

import "testing"

func TestValidTimingPolicyBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name   string
		policy TimingPolicy
		valid  bool
	}{
		{
			name: "minimums",
			policy: TimingPolicy{
				RideRequestTTLSeconds:       MinRideRequestTTLSeconds,
				DriverOpportunityTTLSeconds: MinDriverOpportunityTTLSeconds,
				OfferDecisionTTLSeconds:     MinOfferDecisionTTLSeconds,
			},
			valid: true,
		},
		{
			name: "maximums",
			policy: TimingPolicy{
				RideRequestTTLSeconds:       MaxRideRequestTTLSeconds,
				DriverOpportunityTTLSeconds: MaxDriverOpportunityTTLSeconds,
				OfferDecisionTTLSeconds:     MaxOfferDecisionTTLSeconds,
			},
			valid: true,
		},
		{
			name: "ride_too_short",
			policy: TimingPolicy{
				RideRequestTTLSeconds:       MinRideRequestTTLSeconds - 1,
				DriverOpportunityTTLSeconds: 30,
				OfferDecisionTTLSeconds:     10,
			},
			valid: false,
		},
		{
			name: "opportunity_too_long",
			policy: TimingPolicy{
				RideRequestTTLSeconds:       180,
				DriverOpportunityTTLSeconds: MaxDriverOpportunityTTLSeconds + 1,
				OfferDecisionTTLSeconds:     10,
			},
			valid: false,
		},
		{
			name: "offer_too_short",
			policy: TimingPolicy{
				RideRequestTTLSeconds:       180,
				DriverOpportunityTTLSeconds: 30,
				OfferDecisionTTLSeconds:     MinOfferDecisionTTLSeconds - 1,
			},
			valid: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidTimingPolicy(tc.policy); got != tc.valid {
				t.Fatalf("ValidTimingPolicy()=%v want %v", got, tc.valid)
			}
		})
	}
}
