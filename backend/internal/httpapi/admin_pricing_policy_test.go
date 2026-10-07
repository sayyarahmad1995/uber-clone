package httpapi
import("context";"encoding/json";"net/http";"net/http/httptest";"net/url";"strings";"testing")
func adminPricingRequest(api *API,method,path,body string,form,authorized,validOrigin bool)*httptest.ResponseRecorder{
 req:=httptest.NewRequest(method,"http://application.test"+path,strings.NewReader(body))
 if form{req.Header.Set("Content-Type","application/x-www-form-urlencoded")}else{req.Header.Set("Content-Type","application/json")}
 if authorized{req.SetBasicAuth("reviewer","test-password")};if validOrigin{req.Header.Set("Origin","http://application.test")}else{req.Header.Set("Origin","http://wrong.test")}
 res:=httptest.NewRecorder();api.Handler().ServeHTTP(res,req);return res
}
func adminPricingAPI(t *testing.T)(*API,string){
 db,code,id,p:=pricedDB(t);api:=pricedAPI(db,id,p,false,&routePreviewSpy{});api.adminReviewUsername="reviewer";api.adminReviewPassword="test-password";return api,code
}
func pricingForm(code string)string{return url.Values{"service_code":{code},"base_fare":{"100.00"},"rate_per_km":{"10.00"},"rate_per_minute":{"1.00"},"minimum_fare":{"100.00"},"rounding_increment":{"1.00"}}.Encode()}
func TestAdminPricingAuthorization(t *testing.T){
 api,code:=adminPricingAPI(t)
 for _,tc:=range []struct{auth,origin bool;want int}{{false,true,401},{true,false,403}}{
  res:=adminPricingRequest(api,"POST","/admin/operations/pricing/publish",pricingForm(code),true,tc.auth,tc.origin)
  if res.Code!=tc.want{t.Fatalf("auth/origin %d %s",res.Code,res.Body.String())}
 }
 policies,e:=api.pricing.List(context.Background());if e!=nil{t.Fatal(e)};for _,p:=range policies{if p.ServiceCode==code{t.Fatal("unauthorized publication persisted")}}
 res:=adminPricingRequest(api,"GET","/admin/operations/pricing","",false,true,true);if res.Code!=200||!strings.Contains(res.Body.String(),"Pricing")||!strings.Contains(res.Body.String(),"service_code"){t.Fatalf("page %d %s",res.Code,res.Body.String())}
}
func TestAdminPricingPublish(t *testing.T){
 api,code:=adminPricingAPI(t)
 res:=adminPricingRequest(api,"POST","/admin/operations/pricing/publish",pricingForm(code),true,true,true);if res.Code!=303{t.Fatalf("publish %d %s",res.Code,res.Body.String())}
 a,e:=api.pricing.Current(context.Background(),code,"PKR");if e!=nil||a.BaseFareMinor!=10000||a.RateMinorPerKm!=1000||a.PublishedBy!="reviewer"{t.Fatalf("policy %+v %v",a,e)}
 if api.suggestedFaresEnabled{t.Fatal("publication enabled rollout")}
 res=adminPricingRequest(api,"POST","/v1/admin/pricing-policies",`{"service_code":"`+code+`","currency":"PKR","base_fare_minor":20000,"rate_minor_per_km":1000,"rate_minor_per_minute":100,"minimum_fare_minor":10000,"rounding_increment_minor":100}`,false,true,true)
 if res.Code!=201{t.Fatalf("JSON publish %d %s",res.Code,res.Body.String())}
 b,e:=api.pricing.Current(context.Background(),code,"PKR");if e!=nil||b.Version!=2||b.BaseFareMinor!=20000{t.Fatalf("version %+v %v",b,e)}
 res=adminPricingRequest(api,"POST","/v1/admin/pricing-policies",`{"unknown":true}`,false,true,true);if res.Code!=400{t.Fatalf("unknown fields %d",res.Code)}
 invalid:=strings.Replace(pricingForm(code),"100.00","100.001",1);res=adminPricingRequest(api,"POST","/admin/operations/pricing/publish",invalid,true,true,true);if res.Code!=400{t.Fatalf("decimal %d",res.Code)}
 res=adminPricingRequest(api,"POST","/admin/operations/pricing/publish",pricingForm("unknown_service"),true,true,true);if res.Code!=400{t.Fatalf("unknown service %d",res.Code)}
 res=adminPricingRequest(api,"GET","/v1/admin/pricing-policies","",false,true,true);if res.Code!=200||!json.Valid(res.Body.Bytes())||!strings.Contains(res.Body.String(),a.ID.String())||!strings.Contains(res.Body.String(),b.ID.String()){t.Fatalf("history %d %s",res.Code,res.Body.String())}
}
func TestAdminPricingDisable(t *testing.T){
 api,code:=adminPricingAPI(t);a,e:=api.pricing.Publish(context.Background(),pricedDraft(code),"A");if e!=nil{t.Fatal(e)};b,e:=api.pricing.Publish(context.Background(),pricedDraft(code),"B");if e!=nil{t.Fatal(e)}
 res:=adminPricingRequest(api,"POST","/v1/admin/pricing-policies/"+a.ID.String()+"/disable","{}",false,true,true);if res.Code!=409{t.Fatalf("stale disable %d %s",res.Code,res.Body.String())}
 p,e:=api.pricing.Current(context.Background(),code,"PKR");if e!=nil||p.ID!=b.ID{t.Fatalf("stale disable changed current %+v %v",p,e)}
 res=adminPricingRequest(api,"POST","/admin/operations/pricing/disable",url.Values{"policy_id":{b.ID.String()}}.Encode(),true,true,true);if res.Code!=303{t.Fatalf("disable %d %s",res.Code,res.Body.String())}
 res=adminPricingRequest(api,"GET","/v1/admin/pricing-policies","",false,true,true);if !strings.Contains(res.Body.String(),b.ID.String()){t.Fatal("disable removed history")}
}
