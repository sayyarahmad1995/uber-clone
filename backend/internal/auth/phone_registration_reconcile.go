package auth

import (
 "context"
 "errors"

 "github.com/google/uuid"
)

var ErrPhoneRegistrationReconciliation = errors.New("phone registration reconciliation failed")

// RegistrationIdentityVerifier verifies the actual Kratos identity by subject
// and phone through a trusted admin API, never from user-submitted traits.
type RegistrationIdentityVerifier interface {
 VerifyPhoneIdentity(context.Context,string,string)(bool,error)
}

type RegistrationRecoveryRepository interface {
 GetRegistration(context.Context,string)(RegistrationOperation,error)
 RecordRegistrationIdentity(context.Context,uuid.UUID,string)(RegistrationOperation,error)
 CompleteRegistration(context.Context,uuid.UUID)(uuid.UUID,error)
}

// ReconcilePhoneRegistration is a privileged, operator-directed recovery
// action. The operator must locate and review the exact Kratos identity
// independently; this method never guesses subjects or creates identities.
// It must never be exposed as an unauthenticated HTTP endpoint.
func ReconcilePhoneRegistration(ctx context.Context, repo RegistrationRecoveryRepository, verifier RegistrationIdentityVerifier, phone string, operationID uuid.UUID, subject string)(uuid.UUID,error){
 if repo==nil||verifier==nil||subject==""||operationID==uuid.Nil{
  return uuid.Nil,ErrPhoneRegistrationReconciliation
 }
 normalized,err:=(PhoneChallengeRequest{PhoneNumber:phone,Purpose:PhoneChallengeSignInOrRegister}).Validate()
 if err!=nil{return uuid.Nil,ErrPhoneRegistrationReconciliation}
 op,err:=repo.GetRegistration(ctx,normalized.PhoneNumber)
 if err!=nil{return uuid.Nil,err}
 if op.ID!=operationID || op.Phone!=normalized.PhoneNumber{
  return uuid.Nil,ErrPhoneRegistrationReconciliation
 }
 if op.State!="submitting"&&op.State!="identity_created"&&op.State!="completed"{
  return uuid.Nil,ErrPhoneRegistrationReconciliation
 }
 if op.Subject!=""&&op.Subject!=subject{
  return uuid.Nil,ErrPhoneRegistrationReconciliation
 }
 matches,err:=verifier.VerifyPhoneIdentity(ctx,subject,normalized.PhoneNumber)
 if err!=nil{return uuid.Nil,err}
 if !matches{return uuid.Nil,ErrPhoneRegistrationReconciliation}
 if op.State=="submitting"{
  if _,err=repo.RecordRegistrationIdentity(ctx,op.ID,subject);err!=nil{return uuid.Nil,err}
 }
 return repo.CompleteRegistration(ctx,op.ID)
}
