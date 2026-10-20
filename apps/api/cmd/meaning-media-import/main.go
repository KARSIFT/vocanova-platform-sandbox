// Command meaning-media-import records an operator-reviewed, already-uploaded
// manifest. It never uploads, downloads, generates or exposes original images.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaningmedia"
	_ "github.com/lib/pq"
)

func run() error {
	manifestFile := flag.String("manifest", "", "Approved metadata manifest file (required)")
	flag.Parse()
	if *manifestFile == "" || flag.NArg() != 0 {
		return fmt.Errorf("usage: meaning-media-import --manifest FILE")
	}
	f, err := os.Open(*manifestFile)
	if err != nil {
		return fmt.Errorf("open manifest")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.Size() > 16*1024*1024 {
		return fmt.Errorf("manifest exceeds the 16 MiB limit")
	}
	decoder := json.NewDecoder(io.LimitReader(f, 16*1024*1024+1))
	decoder.DisallowUnknownFields()
	var manifest meaningmedia.Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return fmt.Errorf("invalid manifest JSON")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return fmt.Errorf("manifest contains trailing data or exceeds limit")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is required in the operator environment")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open database")
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	n, err := meaningmedia.NewService(db).Import(ctx, manifest)
	if err != nil {
		return fmt.Errorf("import failed; no metadata committed: %w", err)
	}
	fmt.Printf("Imported %d new approved meaning images (%d checked).\n", n, len(manifest.Images))
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
