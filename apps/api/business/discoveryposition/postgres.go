package discoveryposition

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
)

type PostgreSQLRepository struct{ db *sql.DB }

func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db} }

const readPosition = `SELECT p.catalog_version,p.band,p.rank,p.active,p.revision FROM user_discovery_positions p WHERE p.user_id=$1`

func scanPosition(row interface{ Scan(...any) error }) (*Position, error) {
	var p Position
	err := row.Scan(&p.CatalogVersion, &p.Band, &p.Rank, &p.Active, &p.Revision)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}
func (r *PostgreSQLRepository) Get(ctx context.Context, user uuid.UUID) (*Position, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var exists bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL)`, user).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}
	p, err := scanPosition(tx.QueryRowContext(ctx, readPosition, user))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return p, nil
}
func (r *PostgreSQLRepository) Put(ctx context.Context, user uuid.UUID, u Update) (*Position, error) {
	if err := u.Validate(); err != nil {
		return nil, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// Serialize creation and CAS with each other and account deletion.
	var owner uuid.UUID
	err = tx.QueryRowContext(ctx, `SELECT id FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL FOR UPDATE`, user).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var exists bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM knowledge_catalog_entries kc JOIN canonical_words cw ON cw.id=kc.word_id AND cw.status='active' JOIN word_meanings wm ON wm.id=kc.meaning_id AND wm.status='active' WHERE kc.rank=$1)`, u.Rank).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrInvalid
	}
	p, err := scanPosition(tx.QueryRowContext(ctx, readPosition, user))
	if err != nil {
		return nil, err
	}
	// Safe retry after a lost response is a read-only success. Different stale
	// intent never acquires a new revision and cannot replace a confirmed position.
	if p != nil && p.CatalogVersion == u.CatalogVersion && p.Band == u.Band && p.Rank == u.Rank && p.Active == u.Active {
		return p, tx.Commit()
	}
	revision := int64(0)
	if p != nil {
		revision = p.Revision
	}
	if revision != u.ExpectedRevision {
		return nil, ErrConflict
	}
	p, err = scanPosition(tx.QueryRowContext(ctx, `INSERT INTO user_discovery_positions(user_id,catalog_version,band,rank,active,revision,updated_at) VALUES($1,$2,$3,$4,$5,1,now()) ON CONFLICT(user_id) DO UPDATE SET catalog_version=EXCLUDED.catalog_version,band=EXCLUDED.band,rank=EXCLUDED.rank,active=EXCLUDED.active,revision=user_discovery_positions.revision+1,updated_at=now() RETURNING catalog_version,band,rank,active,revision`, user, u.CatalogVersion, u.Band, u.Rank, u.Active))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return p, nil
}
