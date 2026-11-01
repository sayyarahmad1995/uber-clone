package auth

import (
 "context"
 "errors"
 "testing"

 "github.com/google/uuid"
)

type coordinatorProof struct{ calls int; phone string }
func(p *coordinatorProof) ConsumeProof(_ context.Context,_ string,purpose PhoneChallengePurpose)(PhoneVerificationProof,error){
 p.calls++
 return PhoneVerificationProof{PhoneNumber:p.phone,Purpose:purpose},nil
}
type coordinatorKratos struct{ calls int; err error }
func(k *coordinatorKratos) RegisterPhoneIdentity(_ context.Context,_,_ string)(PhoneRegistrationIdentity,error){
 k.calls++
 if k.err!=nil{return PhoneRegistrationIdentity{},k.err}
 return PhoneRegistrationIdentity{IdentitySubject:"kratos-test-subject"},nil
}
func TestRegistrationCoordinatorSingleSubmissionAndReplay(t *testing.T){
 store:=phoneChallengeTestStore(t)
 repo,err:=NewPostgresPhoneAccountResolver(store.db,"primary-identity-v1")
 if err!=nil{t.Fatal(err)}
 proofs:=&coordinatorProof{phone:"+923001234592"}
 kratos:=&coordinatorKratos{}
 service:=PhoneRegistrationCoordinator{Proofs:proofs,Identities:kratos,Operations:repo}
 first,err:=service.Register(context.Background(),"token","password")
 if err!=nil||first==uuid.Nil{t.Fatalf("register: %s %v",first,err)}
 if kratos.calls!=1{t.Fatalf("kratos called %d times",kratos.calls)}
 if _,err=service.Register(context.Background(),"other-token","password");!errors.Is(err,ErrPhoneRegistrationPending){
  t.Fatalf("replay: %v",err)
 }
 if kratos.calls!=1{t.Fatal("replay called Kratos")}
}
func TestRegistrationCoordinatorAmbiguousKratosFailureDoesNotRetry(t *testing.T){
 store:=phoneChallengeTestStore(t)
 repo,err:=NewPostgresPhoneAccountResolver(store.db,"primary-identity-v1")
 if err!=nil{t.Fatal(err)}
 proofs:=&coordinatorProof{phone:"+923001234593"}
 kratos:=&coordinatorKratos{err:errors.New("timeout")}
 service:=PhoneRegistrationCoordinator{Proofs:proofs,Identities:kratos,Operations:repo}
 if _,err=service.Register(context.Background(),"token","password");!errors.Is(err,ErrPhoneRegistrationPending){
  t.Fatalf("ambiguous: %v",err)
 }
 if _,err=service.Register(context.Background(),"other-token","password");!errors.Is(err,ErrPhoneRegistrationPending){
  t.Fatalf("retry: %v",err)
 }
 if kratos.calls!=1{t.Fatalf("duplicated identity creation: %d",kratos.calls)}
 operation,err:=repo.GetRegistration(context.Background(),proofs.phone)
 if err!=nil||operation.State!="submitting"{t.Fatalf("state: %+v %v",operation,err)}
}
