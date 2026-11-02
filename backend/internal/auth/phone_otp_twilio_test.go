package auth

import (
 "bytes"
 "context"
 "errors"
 "net/http"
 "net/http/httptest"
 "net/url"
 "strings"
 "testing"
 "time"

 "github.com/google/uuid"
)

func TestTwilioOTPSenderHTTPContract(t *testing.T) {
 account := "AC"+strings.Repeat("a",32)
 key := "SK"+strings.Repeat("b",32)
 service := "MG"+strings.Repeat("c",32)
 phone := "+923001234567"
 calls := 0
 status := http.StatusCreated
 response := `{"sid":"SM0123456789abcdef0123456789abcdef"}`
 server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
  calls++
  if r.Method!=http.MethodPost || r.URL.Path!="/2010-04-01/Accounts/"+account+"/Messages.json" { t.Errorf("unexpected route: %s",r.URL.Path) }
  user,pass,ok:=r.BasicAuth()
  if !ok || user!=key || pass!="dummy-secret" { t.Error("bad authorization") }
  if r.Header.Get("Content-Type")!="application/x-www-form-urlencoded" { t.Error("content type") }
  if err:=r.ParseForm();err!=nil { t.Error(err) }
  if r.Form.Get("To")!=phone || r.Form.Get("MessagingServiceSid")!=service || r.Form.Get("From")!="" ||
   !strings.Contains(r.Form.Get("Body"),"123456") { t.Error("unexpected form") }
  w.WriteHeader(status)
  _,_=w.Write([]byte(response))
 }))
 defer server.Close()
 sender,err:=NewTwilioOTPSender(server.Client(),account,key,"dummy-secret",service,true)
 if err!=nil { t.Fatal(err) }
 sender.endpoint=server.URL+"/2010-04-01/Accounts/"+account+"/Messages.json"
 if err:=sender.SendPhoneOTP(context.Background(),uuid.New(),phone,"123456");err!=nil { t.Fatal(err) }
 status=http.StatusBadRequest
 if err:=sender.SendPhoneOTP(context.Background(),uuid.New(),phone,"123456");err==nil || errors.Is(err,ErrPhoneOTPSendUncertain) { t.Fatalf("expected definite rejection: %v",err) }
 status=http.StatusInternalServerError
 if err:=sender.SendPhoneOTP(context.Background(),uuid.New(),phone,"123456");!errors.Is(err,ErrPhoneOTPSendUncertain) { t.Fatalf("expected uncertainty: %v",err) }
 status=http.StatusCreated
 response="{}"
 if err:=sender.SendPhoneOTP(context.Background(),uuid.New(),phone,"123456");!errors.Is(err,ErrPhoneOTPSendUncertain) { t.Fatalf("missing SID: %v",err) }
 if calls!=4 { t.Fatalf("calls=%d",calls) }
}

func TestTwilioOTPSenderTransportFailureIsUncertain(t *testing.T) {
 sender,err:=NewTwilioOTPSender(&http.Client{Timeout:time.Second},"AC"+strings.Repeat("a",32),
  "SK"+strings.Repeat("b",32),"dummy-secret","+15551234567",false)
 if err!=nil { t.Fatal(err) }
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {}))
 sender.endpoint=server.URL
 server.Close()
 err=sender.SendPhoneOTP(context.Background(),uuid.New(),"+923001234567","123456")
 if !errors.Is(err,ErrPhoneOTPSendUncertain) { t.Fatalf("transport result: %v",err) }
}

type uncertainOTPSender struct{}
func (uncertainOTPSender) SendPhoneOTP(context.Context,uuid.UUID,string,string) error { return ErrPhoneOTPSendUncertain }

func TestOTPDispatchWorkerSuppressesUncertainSendRetry(t *testing.T) {
 store:=redisOTPTestStore(t)
 dispatch,err:=NewRedisOTPDispatch(store.client,bytes.Repeat([]byte{0x98},32))
 if err!=nil { t.Fatal(err) }
 id:=uuid.New()
 if ok,err:=dispatch.Enqueue(context.Background(),id,"+923001234567","123456",time.Minute);err!=nil || !ok { t.Fatalf("enqueue: %v %v",ok,err) }
 worker,err:=NewOTPDispatchWorker(dispatch,uncertainOTPSender{},10*time.Second)
 if err!=nil { t.Fatal(err) }
 if err:=worker.Deliver(context.Background(),id);!errors.Is(err,ErrPhoneOTPDelivery) { t.Fatalf("delivery: %v",err) }
 if err:=worker.Deliver(context.Background(),id);!errors.Is(err,ErrOTPDispatchMissing) { t.Fatalf("unexpected retry: %v",err) }
}

var _ = url.Values{}
