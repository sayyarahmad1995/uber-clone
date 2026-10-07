package pricing
import "testing"
func TestParsePKR(t *testing.T){
 for _,tc:=range []struct{in string;want int64}{{"12.34",1234},{"0",0},{"12.3",1230},{"12",1200},{"10000000000.00",1000000000000}}{got,err:=ParsePKR(tc.in);if err!=nil||got!=tc.want{t.Fatalf("%q got %d %v",tc.in,got,err)}}
 for _,in:=range []string{"","-1","+1","NaN","1e3","12.345",".5","1.","1 2","10000000000.01","999999999999999999999999999999"} {if _,err:=ParsePKR(in);err==nil{t.Fatalf("accepted %q",in)}}
}
