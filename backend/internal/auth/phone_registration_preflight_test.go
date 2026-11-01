package auth

import (
 "context"
 "errors"
 "testing"

 "github.com/google/uuid"
)

type failedReserveRepository struct{}
func (failedReserveRepository) ReserveRegistration(context.Context,string)(RegistrationOperation,error){return RegistrationOperation{},errors.New("postgres unavailable")}
func (failedReserveRepository) ClaimRegistrationSubmission(context.Context,uuid.UUID)(bool,error){panic("unexpected claim")}
func (failedReserveRepository) RecordRegistrationIdentity(context.Context,uuid.UUID,string)(RegistrationOperation,error){panic("unexpected record")}
func (failedReserveRepository) CompleteRegistration(context.Context,uuid.UUID)(uuid.UUID,error){panic("unexpected completion")}

func TestRegistrationCoordinatorDoesNotConsumeProofWhenReservationFails(t *testing.T){
 proof:=&coordinatorProof{phone:"+923001234595"}
 creator:=&coordinatorKratos{}
 svc:=PhoneRegistrationCoordinator{Proofs:proof,Identities:creator,Operations:failedReserveRepository{}}
 _,err:=svc.Register(context.Background(),"verified-token","secure-password")
 if err==nil{t.Fatal("expected PostgreSQL failure")}
 if proof.calls!=0{t.Fatalf("consumed proof before durable reservation: %d",proof.calls)}
 if creator.calls!=0{t.Fatal("called Kratos without durable reservation")}
}
