-- Distinguish safe reservation from possibly dispatched Kratos creation.
ALTER TABLE phone_registration_operations
 DROP CONSTRAINT phone_registration_operations_state_check;
ALTER TABLE phone_registration_operations
 ADD CONSTRAINT phone_registration_operations_state_check
 CHECK (state IN ('reserved','submitting','identity_created','completed','needs_reconciliation'));
ALTER TABLE phone_registration_operations
 DROP CONSTRAINT phone_registration_state_fields;
ALTER TABLE phone_registration_operations
 ADD CONSTRAINT phone_registration_state_fields CHECK (
  (state IN ('reserved','submitting') AND identity_subject IS NULL AND user_id IS NULL) OR
  (state IN ('identity_created','needs_reconciliation') AND identity_subject IS NOT NULL AND user_id IS NULL) OR
  (state='completed' AND identity_subject IS NOT NULL AND user_id IS NOT NULL)
 );
