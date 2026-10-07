package pricing

import ("testing";"errors")
func fixture() Policy {return Policy{ServiceCode:"economy",Currency:"PKR",BaseFareMinor:10000,RateMinorPerKm:1000,RateMinorPerMinute:100,MinimumFareMinor:10000,RoundingIncrementMinor:100,CalculationRule:"half_up_once_v1"}}
func TestCalculateSingleRounding(t *testing.T) {
 p:=fixture(); got,err:=Calculate(p,2500,90);if err!=nil||got!=12700 {t.Fatalf("got %d, %v; want 12700",got,err)}
 p.BaseFareMinor=0;p.RateMinorPerKm=1;p.RateMinorPerMinute=1;p.MinimumFareMinor=1;p.RoundingIncrementMinor=1
 got,err=Calculate(p,500,30);if err!=nil||got!=1 {t.Fatalf("single final rounding got %d, %v; want 1",got,err)}
 p.RoundingIncrementMinor=10;p.RateMinorPerKm=10;p.RateMinorPerMinute=0
 got,err=Calculate(p,500,1);if err!=nil||got!=10 {t.Fatalf("half tie got %d, %v",got,err)}
}
func TestCalculateMinimum(t *testing.T){p:=fixture();p.BaseFareMinor=0;p.RateMinorPerKm=0;p.RateMinorPerMinute=0;p.MinimumFareMinor=125;got,err:=Calculate(p,1,1);if err!=nil||got!=125 {t.Fatalf("minimum got %d %v",got,err)}}
func TestCalculateInvalidAndOverflow(t *testing.T){
 for _,metrics:=range [][2]int64{{0,1},{1,0},{-1,1},{1,-1}} {_,err:=Calculate(fixture(),metrics[0],metrics[1]);if !errors.Is(err,ErrInvalidRoute){t.Fatalf("metrics %v: %v",metrics,err)}}
 p:=fixture();p.RateMinorPerKm=MaxAmountMinor;_,err:=Calculate(p,9223372036854775807,1);if !errors.Is(err,ErrFareOutOfRange){t.Fatalf("overflow: %v",err)}
 p.RateMinorPerKm=0;p.RateMinorPerMinute=0;p.BaseFareMinor=MaxAmountMinor;p.RoundingIncrementMinor=1
 got,err:=Calculate(p,9223372036854775807,9223372036854775807);if err!=nil||got!=MaxAmountMinor{t.Fatalf("large valid metrics %d %v",got,err)}
 p.MinimumFareMinor=-1;if _,err:=Calculate(p,1,1);!errors.Is(err,ErrInvalidPolicy){t.Fatalf("invalid policy: %v",err)}
}
func TestCalculateServiceFixtures(t *testing.T){
 for _,tc:=range []struct{service string;base,want int64}{{"economy",10000,12700},{"comfort",20000,22700},{"synthetic",30000,32700}}{p:=fixture();p.ServiceCode=tc.service;p.BaseFareMinor=tc.base;got,err:=Calculate(p,2500,90);if err!=nil||got!=tc.want{t.Fatalf("%s: %d %v",tc.service,got,err)}}
}
