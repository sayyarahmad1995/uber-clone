package main

import (
 "context"
 "encoding/base64"
 "errors"
 "fmt"
 "net/http"
 "time"

 "github.com/redis/go-redis/v9"
 "github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

func buildPhoneSMSWorker(cfg config) (auth.OTPDispatchWorker, func(), error) {
 if !cfg.PhoneSMSWorkerEnabled {
  return auth.OTPDispatchWorker{}, nil, nil
 }
 // The key must be the same stable 32-byte AES key used for atomic
 // issuance. Do not reuse the OTP HMAC secret for encryption.
 key, err := base64.StdEncoding.DecodeString(cfg.PhoneSMSDispatchKey)
 if err != nil || len(key) != 32 {
  return auth.OTPDispatchWorker{}, nil, fmt.Errorf("PHONE_SMS_DISPATCH_ENCRYPTION_KEY must be base64 of exactly 32 bytes")
 }
 if (cfg.TwilioMessagingServiceSID == "") == (cfg.TwilioFromNumber == "") {
  return auth.OTPDispatchWorker{}, nil, errors.New("configure exactly one Twilio sender identity")
 }
 senderID, useService := cfg.TwilioFromNumber, false
 if cfg.TwilioMessagingServiceSID != "" {
  senderID, useService = cfg.TwilioMessagingServiceSID, true
 }
 sender, err := auth.NewTwilioOTPSender(&http.Client{Timeout: 10 * time.Second},
  cfg.TwilioAccountSID, cfg.TwilioAPIKeySID, cfg.TwilioAPIKeySecret, senderID, useService)
 if err != nil {
  return auth.OTPDispatchWorker{}, nil, fmt.Errorf("invalid Twilio SMS configuration: %w", err)
 }
 options, err := redis.ParseURL(cfg.RedisURL)
 if err != nil {
  return auth.OTPDispatchWorker{}, nil, fmt.Errorf("invalid SMS Redis URL: %w", err)
 }
 client := redis.NewClient(options)
 closeClient := func() { _ = client.Close() }
 ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
 defer cancel()
 if err := client.Ping(ctx).Err(); err != nil {
  closeClient()
  return auth.OTPDispatchWorker{}, nil, fmt.Errorf("SMS worker Redis unavailable: %w", err)
 }
 dispatch, err := auth.NewRedisOTPDispatch(client, key)
 if err != nil {
  closeClient()
  return auth.OTPDispatchWorker{}, nil, err
 }
 worker, err := auth.NewOTPDispatchWorker(dispatch, sender, 20*time.Second)
 if err != nil {
  closeClient()
  return auth.OTPDispatchWorker{}, nil, err
 }
 return worker, closeClient, nil
}
