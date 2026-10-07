package pricing
import ("context";"database/sql";"errors";"strings";"github.com/google/uuid")
type PostgresRepository struct {db *sql.DB}
func NewPostgresRepository(db *sql.DB)*PostgresRepository{return &PostgresRepository{db:db}}
const policyColumns="p.id,p.service_code,p.currency,p.version,p.base_fare_minor,p.rate_minor_per_km,p.rate_minor_per_minute,p.minimum_fare_minor,p.rounding_increment_minor,p.calculation_rule,p.effective_from,p.published_at,p.published_by"
type scanner interface{Scan(...any)error}
func scanPolicy(row scanner)(Policy,error){var p Policy;err:=row.Scan(&p.ID,&p.ServiceCode,&p.Currency,&p.Version,&p.BaseFareMinor,&p.RateMinorPerKm,&p.RateMinorPerMinute,&p.MinimumFareMinor,&p.RoundingIncrementMinor,&p.CalculationRule,&p.EffectiveFrom,&p.PublishedAt,&p.PublishedBy);return p,err}
func (r *PostgresRepository) Current(ctx context.Context,service,currency string)(Policy,error){
 var available bool
 if err:=r.db.QueryRowContext(ctx,"SELECT EXISTS(SELECT 1 FROM driver_service_catalog WHERE code=$1 AND is_active AND rider_visible)",service).Scan(&available);err!=nil{return Policy{},err}
 if !available{return Policy{},ErrInvalidService}
 p,err:=scanPolicy(r.db.QueryRowContext(ctx,"SELECT "+policyColumns+" FROM ride_pricing_policy_versions p JOIN ride_pricing_policy_current c ON c.current_policy_id=p.id WHERE c.service_code=$1 AND c.currency=$2",service,currency))
 if errors.Is(err,sql.ErrNoRows){return Policy{},ErrUnavailable};return p,err
}
func lockService(ctx context.Context,tx *sql.Tx,service,mode string,requireActive bool)error{
 var active,visible bool;err:=tx.QueryRowContext(ctx,"SELECT is_active,rider_visible FROM driver_service_catalog WHERE code=$1 FOR "+mode,service).Scan(&active,&visible)
 if errors.Is(err,sql.ErrNoRows){return ErrInvalidService};if err!=nil{return err}
 if requireActive&&(!active||!visible){return ErrInvalidService};return nil
}
// LockCurrent stabilizes service activation and first publication as well as the pointer.
// Callers keep the same transaction open through request and snapshot insertion.
func LockCurrent(ctx context.Context,tx *sql.Tx,service,currency string)(Policy,error){
 if err:=lockService(ctx,tx,service,"SHARE",true);err!=nil{return Policy{},err}
 var id sql.NullString
 err:=tx.QueryRowContext(ctx,"SELECT current_policy_id FROM ride_pricing_policy_current WHERE service_code=$1 AND currency=$2 FOR SHARE",service,currency).Scan(&id)
 if errors.Is(err,sql.ErrNoRows)||err==nil&&!id.Valid{return Policy{},ErrUnavailable};if err!=nil{return Policy{},err}
 return scanPolicy(tx.QueryRowContext(ctx,"SELECT "+policyColumns+" FROM ride_pricing_policy_versions p WHERE p.id=$1",id.String))
}
func (r *PostgresRepository) Publish(ctx context.Context,d Draft,reviewer string)(Policy,error){
 d.ServiceCode=strings.ToLower(strings.TrimSpace(d.ServiceCode));d.Currency=strings.ToUpper(strings.TrimSpace(d.Currency));reviewer=strings.TrimSpace(reviewer)
 if err:=Validate(d);err!=nil{return Policy{},err};if reviewer==""{return Policy{},ErrInvalidPolicy}
 tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return Policy{},err};defer tx.Rollback()
 if err:=lockService(ctx,tx,d.ServiceCode,"UPDATE",true);err!=nil{return Policy{},err}
 var version int64
 if err:=tx.QueryRowContext(ctx,"SELECT COALESCE(MAX(version),0)+1 FROM ride_pricing_policy_versions WHERE service_code=$1 AND currency=$2",d.ServiceCode,d.Currency).Scan(&version);err!=nil{return Policy{},err}
 p,err:=scanPolicy(tx.QueryRowContext(ctx,"INSERT INTO ride_pricing_policy_versions AS p(id,service_code,currency,version,base_fare_minor,rate_minor_per_km,rate_minor_per_minute,minimum_fare_minor,rounding_increment_minor,calculation_rule,published_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "+policyColumns,uuid.New(),d.ServiceCode,d.Currency,version,d.BaseFareMinor,d.RateMinorPerKm,d.RateMinorPerMinute,d.MinimumFareMinor,d.RoundingIncrementMinor,CalculationRule,reviewer))
 if err!=nil{return Policy{},err}
 _,err=tx.ExecContext(ctx,"INSERT INTO ride_pricing_policy_current(service_code,currency,current_policy_id,updated_by) VALUES($1,$2,$3,$4) ON CONFLICT(service_code,currency) DO UPDATE SET current_policy_id=EXCLUDED.current_policy_id,updated_by=EXCLUDED.updated_by,updated_at=clock_timestamp()",d.ServiceCode,d.Currency,p.ID,reviewer)
 if err!=nil{return Policy{},err};if err:=tx.Commit();err!=nil{return Policy{},err};return p,nil
}
func (r *PostgresRepository) Disable(ctx context.Context,service,currency string,expectedPolicyID uuid.UUID,reviewer string)error{
 if strings.TrimSpace(reviewer)==""{return ErrInvalidPolicy}
 tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return err};defer tx.Rollback()
 if err:=lockService(ctx,tx,service,"UPDATE",false);err!=nil{return err}
 result,err:=tx.ExecContext(ctx,"UPDATE ride_pricing_policy_current SET current_policy_id=NULL,updated_by=$4,updated_at=clock_timestamp() WHERE service_code=$1 AND currency=$2 AND current_policy_id=$3",service,currency,expectedPolicyID,reviewer)
 if err!=nil{return err};n,err:=result.RowsAffected();if err!=nil{return err};if n!=1{return ErrPolicyChanged};return tx.Commit()
}
func (r *PostgresRepository) List(ctx context.Context)([]Policy,error){
 rows,err:=r.db.QueryContext(ctx,"SELECT "+policyColumns+" FROM ride_pricing_policy_versions p ORDER BY service_code,currency,version DESC");if err!=nil{return nil,err};defer rows.Close()
 result:=make([]Policy,0);for rows.Next(){p,err:=scanPolicy(rows);if err!=nil{return nil,err};result=append(result,p)};return result,rows.Err()
}
