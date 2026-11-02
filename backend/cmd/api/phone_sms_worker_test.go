package main

import (
 "encoding/base64"
 "strings"
 "testing"
)

func TestPhoneSMSWorkerDisabledByDefault(t *testing.T) {
 worker, closeFn, err := buildPhoneSMSWorker(config{})
 if err != nil || closeFn != nil {
  t.Fatalf("disabled SMS worker must not initialize Redis or Twilio: %v", err)
 }
 _ = worker
}

func TestPhoneSMSWorkerRejectsInvalidConfigurationBeforeNetwork(t *testing.T) {
 validKey := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32)))
 for _, tc := range []struct {
  name string
  config config
 }{
  {"missing encryption key", config{PhoneSMSWorkerEnabled:true}},
  {"invalid sender", config{PhoneSMSWorkerEnabled:true, PhoneSMSDispatchKey:validKey}},
  {"both senders", config{PhoneSMSWorkerEnabled:true,PhoneSMSDispatchKey:validKey, TwilioMessagingServiceSID:"MG"+strings.Repeat("a",32),TwilioFromNumber:"+15551234567"}},
  {"missing credentials", config{PhoneSMSWorkerEnabled:true,PhoneSMSDispatchKey:validKey,TwilioFromNumber:"+15551234567"}},
 } {
  t.Run(tc.name, func(t *testing.T) {
   _, closeFn, err := buildPhoneSMSWorker(tc.config)
   if err == nil || closeFn != nil {
    t.Fatalf("expected configuration failure without network access: %v",err)
   }
  })
 }
}

func TestLoadConfigSMSWorkerDefaultDisabled(t *testing.T) {
 t.Setenv("PHONE_SMS_WORKER_ENABLED","")
 cfg := loadConfig()
 if cfg.PhoneSMSWorkerEnabled { t.Fatal("worker enabled without opt-in") }
}
