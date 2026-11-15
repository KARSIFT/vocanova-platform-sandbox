-- atlas:txmode file
CREATE TABLE user_discovery_positions (
 user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
 catalog_version text NOT NULL CHECK (length(catalog_version) BETWEEN 1 AND 64),
 band integer NOT NULL CHECK (band BETWEEN 1 AND 100),
 rank integer NOT NULL CHECK (rank BETWEEN 1 AND 10000 AND band=(rank-1)/100+1),
 active boolean NOT NULL,
 revision bigint NOT NULL CHECK (revision>0),
 updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
