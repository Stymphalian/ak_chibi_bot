package main

// go run server/tools/ingest_assets/main.go -assetDir static/assets -dry-run
// go run server/tools/ingest_assets/main.go -assetDir static/assets -diff
// DB connection is configured via environment variables (see .envrc)

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/akdb"
)

func hashFilePath(filePath string) int64 {
	sum := sha256.Sum256([]byte(filePath))
	return int64(binary.BigEndian.Uint64(sum[:8]))
}

func gzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func collectFilePaths(assetDir string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(assetDir, func(fullPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Store as forward-slash relative path (matches what clients request)
		relPath, err := filepath.Rel(assetDir, fullPath)
		if err != nil {
			return err
		}
		// Skip the _index.json and _names.json files
		if strings.HasSuffix(relPath, "_index.json") || strings.HasSuffix(relPath, "_names.json") {
			return nil
		}

		paths = append(paths, strings.ReplaceAll(relPath, string(filepath.Separator), "/"))
		return nil
	})
	return paths, err
}

func fetchExistingHashes(dbConn *akdb.DatbaseConn) (map[int64]bool, error) {
	var hashes []int64
	result := dbConn.DefaultDB.Raw("SELECT file_path_hash FROM asset_files").Scan(&hashes)
	if result.Error != nil {
		return nil, result.Error
	}
	existing := make(map[int64]bool, len(hashes))
	for _, h := range hashes {
		existing[h] = true
	}
	return existing, nil
}

func run(assetDir string, dryRun bool, diffOnly bool) error {
	start := time.Now()
	var dbConn *akdb.DatbaseConn
	var err error
	if !dryRun {
		dbConn, err = akdb.ProvideDatabaseConn()
		if err != nil {
			return err
		}
	}

	var existingHashes map[int64]bool
	if diffOnly {
		if dryRun {
			return fmt.Errorf("-diff and -dry-run cannot be used together")
		}
		log.Printf("Fetching existing hashes from DB...")
		existingHashes, err = fetchExistingHashes(dbConn)
		if err != nil {
			return err
		}
		log.Printf("Found %d existing entries in DB", len(existingHashes))
	}

	paths, err := collectFilePaths(assetDir)
	if err != nil {
		return err
	}

	total := len(paths)
	log.Printf("Found %d files to ingest", total)

	ingested := 0
	skipped := 0
	var totalRaw, totalCompressed int64
	for i, relPath := range paths {
		fullPath := filepath.Join(assetDir, filepath.FromSlash(relPath))
		_, err := os.Stat(fullPath)
		if err != nil {
			log.Printf("[%d/%d] SKIP (stat error): %s — %v", i+1, total, relPath, err)
			skipped++
			continue
		}

		raw, err := os.ReadFile(fullPath)
		if err != nil {
			log.Printf("[%d/%d] SKIP (read error): %s — %v", i+1, total, relPath, err)
			skipped++
			continue
		}
		data, err := gzipCompress(raw)
		if err != nil {
			log.Printf("[%d/%d] SKIP (gzip error): %s — %v", i+1, total, relPath, err)
			skipped++
			continue
		}
		hash := hashFilePath(relPath)

		totalRaw += int64(len(raw))
		totalCompressed += int64(len(data))

		if diffOnly && existingHashes[hash] {
			continue
		}

		log.Printf("Processing file [%d]%s with size %d bytes compressed", hash, relPath, len(data))
		if !dryRun {
			result := dbConn.DefaultDB.Exec(
				`INSERT INTO asset_files (file_path_hash, file_path, data)
				VALUES (?, ?, ?)
				ON CONFLICT (file_path_hash) DO UPDATE SET data = EXCLUDED.data`,
				hash, relPath, data,
			)
			if result.Error != nil {
				return result.Error
			}
		}

		ingested++
		if (i+1)%100 == 0 || i+1 == total {
			log.Printf("[%d/%d] ingested", i+1, total)
		}
	}

	var savingsPct float64
	if totalRaw > 0 {
		savingsPct = float64(totalRaw-totalCompressed) / float64(totalRaw) * 100
	}
	log.Printf("Done — ingested: %d, skipped: %d, raw: %.1f MB, compressed: %.1f MB, savings: %.1f%%, elapsed: %s",
		ingested, skipped,
		float64(totalRaw)/1e6, float64(totalCompressed)/1e6,
		savingsPct,
		time.Since(start).Round(time.Millisecond),
	)
	return nil
}

func main() {
	assetDirPtr := flag.String("assetDir", "static/assets", "path to the assets root directory")
	dryRunPtr := flag.Bool("dry-run", false, "log what would be ingested without writing to the DB")
	diffOnlyPtr := flag.Bool("diff", false, "only insert files not already present in the DB")
	flag.Parse()

	log.Printf("-assetDir: %s", *assetDirPtr)
	log.Printf("-dry-run: %v", *dryRunPtr)
	log.Printf("-diff: %v", *diffOnlyPtr)
	if err := run(*assetDirPtr, *dryRunPtr, *diffOnlyPtr); err != nil {
		log.Fatal(err)
	}
}
