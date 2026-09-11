package password

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
)

// Cleaner removes short-lived registration hashes and proof records once they
// can no longer be used. Each pass is bounded so retention cannot monopolize DB
// locks. It runs through the existing auth cleanup loop, including at startup.
type Cleaner struct {
	db    *sql.DB
	clock clock.Clock
}

func NewCleaner(db *sql.DB, clk clock.Clock) *Cleaner {
	return &Cleaner{db: db, clock: clk}
}

func (c *Cleaner) Cleanup(ctx context.Context) error {
	now := c.clock.Now().UTC()
	// Names are fixed application constants, never supplied by a caller.
	for _, table := range []string{"password_registration_links", "password_reset_links"} {
		_, err := c.db.ExecContext(ctx, `WITH expired AS (
   SELECT id FROM `+table+`
   WHERE expires_at <= $1 OR consumed_at IS NOT NULL OR revoked_at IS NOT NULL
   ORDER BY expires_at LIMIT 1000 FOR UPDATE SKIP LOCKED
  ) DELETE FROM `+table+` WHERE id IN (SELECT id FROM expired)`, now)
		if err != nil {
			return fmt.Errorf("clean expired password proofs: %w", err)
		}
	}
	return nil
}
