package driver

import (
    "context"
    "github.com/google/uuid"
)

// Presence uses the marketplace freshness deadline. All presence writers lock
// the profile first; recheck timestamps after acquiring that lock.
func (r PostgresRepository) ExpirePresence(ctx context.Context) error {
    return r.expirePresence(ctx,uuid.Nil)
}

func (r PostgresRepository) expirePresence(ctx context.Context,id uuid.UUID) error {
    tx,err:=r.db.BeginTx(ctx,nil)
    if err!=nil { return err }
    defer tx.Rollback()
    rows,err:=tx.QueryContext(ctx,`SELECT user_id FROM driver_profiles
        WHERE is_online AND ($1::uuid='00000000-0000-0000-0000-000000000000' OR user_id=$1)
        FOR UPDATE SKIP LOCKED`,id)
    if err!=nil { return err }
    ids:=[]uuid.UUID{}
    for rows.Next() {
        var locked uuid.UUID
        if err:=rows.Scan(&locked);err!=nil { rows.Close();return err }
        ids=append(ids,locked)
    }
    err=rows.Err()
    rows.Close()
    if err!=nil { return err }
    for _,locked:=range ids {
        _,err=tx.ExecContext(ctx,`UPDATE driver_profiles p
            SET is_online=FALSE,updated_at=statement_timestamp()
            WHERE p.user_id=$1 AND p.is_online AND NOT EXISTS (
                SELECT 1 FROM driver_locations l WHERE l.driver_user_id=p.user_id
                AND l.updated_at BETWEEN statement_timestamp()-($2 * INTERVAL '1 second') AND statement_timestamp())`,locked,MarketplaceLocationMaxAge.Seconds())
        if err!=nil { return err }
    }
    return tx.Commit()
}
