package pricing
import ("strings";"strconv")
func ParsePKR(input string)(int64,error){
 input=strings.TrimSpace(input);parts:=strings.Split(input,".");if len(parts)>2||parts[0]==""{return 0,ErrInvalidPolicy}
 for _,part:=range parts{if part==""{return 0,ErrInvalidPolicy};for _,r:=range part{if r<'0'||r>'9'{return 0,ErrInvalidPolicy}}}
 whole,err:=strconv.ParseInt(parts[0],10,64);if err!=nil||whole>MaxAmountMinor/100{return 0,ErrInvalidPolicy}
 var fraction int64
 if len(parts)==2 {if len(parts[1])>2{return 0,ErrInvalidPolicy};digits:=parts[1];if len(digits)==1{digits+="0"};fraction,_=strconv.ParseInt(digits,10,64)}
 minor:=whole*100+fraction;if minor>MaxAmountMinor{return 0,ErrInvalidPolicy};return minor,nil
}
