package pricing
import "math/big"
func Calculate(p Policy,distanceMeters,durationSeconds int64)(int64,error){
 if err:=Validate(p.Draft());err!=nil{return 0,err}
 if p.CalculationRule!=CalculationRule {return 0,ErrInvalidPolicy}
 if distanceMeters<=0||durationSeconds<=0{return 0,ErrInvalidRoute}
 // A common denominator retains both fractional components until final rounding.
 raw:=new(big.Int).Mul(big.NewInt(p.BaseFareMinor),big.NewInt(3000))
 d:=new(big.Int).Mul(big.NewInt(p.RateMinorPerKm),big.NewInt(distanceMeters));d.Mul(d,big.NewInt(3));raw.Add(raw,d)
 t:=new(big.Int).Mul(big.NewInt(p.RateMinorPerMinute),big.NewInt(durationSeconds));t.Mul(t,big.NewInt(50));raw.Add(raw,t)
 denominator:=new(big.Int).Mul(big.NewInt(p.RoundingIncrementMinor),big.NewInt(3000))
 rounded,remainder:=new(big.Int),new(big.Int);rounded.QuoRem(raw,denominator,remainder)
 if remainder.Mul(remainder,big.NewInt(2)).Cmp(denominator)>=0{rounded.Add(rounded,big.NewInt(1))}
 rounded.Mul(rounded,big.NewInt(p.RoundingIncrementMinor))
 if rounded.Cmp(big.NewInt(p.MinimumFareMinor))<0 {rounded.SetInt64(p.MinimumFareMinor)}
 if rounded.Cmp(big.NewInt(MaxAmountMinor))>0{return 0,ErrFareOutOfRange}
 return rounded.Int64(),nil
}
