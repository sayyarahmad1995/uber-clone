package pricing
import ("testing";"errors")
func TestValidatePolicy(t *testing.T){
 good:=Draft{ServiceCode:"economy",Currency:"PKR",MinimumFareMinor:1,RoundingIncrementMinor:1}
 if err:=Validate(good);err!=nil{t.Fatal(err)}
 for _,mutate:=range []func(*Draft){func(p *Draft){p.ServiceCode=""},func(p *Draft){p.Currency="USD"},func(p *Draft){p.BaseFareMinor=-1},func(p *Draft){p.RateMinorPerKm=-1},func(p *Draft){p.RateMinorPerMinute=-1},func(p *Draft){p.MinimumFareMinor=0},func(p *Draft){p.RoundingIncrementMinor=0},func(p *Draft){p.BaseFareMinor=MaxAmountMinor+1}} {p:=good;mutate(&p);if err:=Validate(p);!errors.Is(err,ErrInvalidPolicy){t.Fatalf("accepted %+v: %v",p,err)}}
}
