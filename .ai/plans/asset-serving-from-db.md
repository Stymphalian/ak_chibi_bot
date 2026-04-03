# Plan: Serve Assets from PostgreSQL

Move binary asset files (`.atlas`, `.skel`, `.png`, `.jpg`, etc.) from the Docker
image filesystem into a new `asset_files` Postgres table. The Go `/image/assets/`
route replaces `http.FileServer` with a DB-backed handler. `*_index.json` files
remain on the filesystem. The `COPY ./static/assets` step is removed from the
`Dockerfile`, shrinking the production image.

**Decisions**
- DB bind-mount: **already done** by user (`data/.db-data`) — only `.gitignore` remains
- Migration files: created via the `golang-migrate` CLI per `db/instructions.txt`
- Ingestion tool: Go, under `server/tools/ingest_assets/main.go` (not Python)
- `*_index.json` files: stay on filesystem, not stored in DB
- Go caching: none — raw Postgres query per request
- Content-Type: derived from extension at serve time (`mime.TypeByExtension`)
- Lookup key: first 8 bytes of SHA-256 of `file_path` stored as `BIGINT` (`INT8`) — native integer primary key for fastest index comparison; `file_path TEXT` retained for debugging

---

## Phase 1 — .gitignore (bind-mount already done)

**Step 1.** Add `data/.db-data/` to `.gitignore`.

---

## Phase 2 — DB Migration

**Step 2.** Use the golang-migrate CLI to generate the migration file pair:
```
docker run -it -v ./db/migrations:/migrations migrate/migrate \
  create -ext sql -dir /migrations asset_store
```
This creates `db/migrations/<timestamp>_asset_store.up.sql` and `.down.sql`.

**Step 3.** Fill in the generated up file:
```sql
BEGIN;
CREATE TABLE IF NOT EXISTS asset_files (
    file_path_hash  BIGINT    PRIMARY KEY,   -- first 8 bytes of SHA-256(file_path) as int64
    file_path       TEXT      NOT NULL,
    data            BYTEA     NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
GRANT SELECT, INSERT, UPDATE, DELETE ON asset_files TO akdb_role;
COMMIT;
```
`BIGINT` (8-byte signed integer) gives native integer comparison on the B-tree primary
key index — faster than either `TEXT` or `CHAR(64)` string comparisons. The value is
the first 8 bytes of SHA-256 interpreted as a big-endian `uint64` cast to `int64`.
At application scale (thousands of asset files) the 64-bit truncated hash has an
essentially zero collision probability.

**Step 4.** Fill in the generated down file:
```sql
BEGIN;
DROP TABLE IF EXISTS asset_files;
COMMIT;
```

---

## Phase 3 — Go Ingestion Tool

Create `server/tools/ingest_assets/main.go`, following the same pattern as
`server/tools/index/create_asset_index.go`.

**Step 5.** CLI flags:
- `-assetDir` — path to the assets root (e.g. `static/assets`)
- DB connection via the existing env vars read by `akdb.ProvideDatabaseConn()`:
  `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_DB`,
  `DATABASE_PASSWORD` / `DATABASE_PASSWORD_FILE`

**Step 6.** Logic:
1. Load three `operator.SpineAssetMap`s via `Load(assetDir, "characters")`,
   `Load(assetDir, "custom")`, `Load(assetDir, "enemies")`
2. Iterate each map using `SpineAssetMap.Iterate(callback)` — inside the callback
   collect the five `PlatformIndie*` filepath fields from each `*SpineData`:
   - `PlaformIndieAtlasFilepath`
   - `PlaformIndieSkelFilepath`
   - `PlaformIndieSkelJsonFilepath`
   - `PlaformIndiePngFilepath`
   - `PlatformIndieSpritesheetDataFilepath`
   (skip empty strings; deduplicate across calls)
3. For each unique file path, read `{assetDir}/{path}` as binary bytes
4. Compute the integer hash of the file path:
   ```go
   sum := sha256.Sum256([]byte(filePath))           // crypto/sha256
   hash := int64(binary.BigEndian.Uint64(sum[:8]))  // encoding/binary
   ```
5. Upsert into `asset_files`:
   ```sql
   INSERT INTO asset_files (file_path_hash, file_path, data)
   VALUES ($1, $2, $3)
   ON CONFLICT (file_path_hash) DO UPDATE SET data = EXCLUDED.data
   ```
   Use `db.DefaultDB.Exec(...)` — keep it simple, no model needed
5. Print running progress (`N/total`) and a final summary

**Step 7.** Add a usage comment at the top of the file:
```
// go run server/tools/ingest_assets/main.go -assetDir static/assets
```
Add `crypto/sha256` and `encoding/binary` to imports.

---

## Phase 4 — Go Backend: Custom Asset Handler

> Replace the `http.FileServer` at `/image/assets/` with a DB-backed handler.

**Step 8.** Add `AssetStore` to `server/internal/akdb/`:
```go
type AssetStore struct {
    db *DatbaseConn
}

func NewAssetStore(db *DatbaseConn) *AssetStore { ... }

func hashFilePath(filePath string) int64 {
    sum := sha256.Sum256([]byte(filePath))          // crypto/sha256
    return int64(binary.BigEndian.Uint64(sum[:8]))  // encoding/binary
}

func (s *AssetStore) GetAsset(filePath string) ([]byte, error) {
    var data []byte
    err := s.db.DefaultDB.Raw(
        "SELECT data FROM asset_files WHERE file_path_hash = ?",
        hashFilePath(filePath),
    ).Scan(&data).Error
    return data, err
}

func ProvideAssetStore(db *DatbaseConn) *AssetStore { ... }
```
Imports needed: `crypto/sha256`, `encoding/binary`.

**Step 9.** In `server/internal/server/server.go` (lines 136–143), replace:
```go
mux.Handle("/image/assets/",
    http.TimeoutHandler(
        http.StripPrefix("/image/assets/", http.FileServer(http.Dir(s.args.ImageAssetDir))),
        DEFAULT_TIMEOUT, "",
    ),
)
```
with:
```go
mux.Handle("/image/assets/",
    http.TimeoutHandler(
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            filePath := strings.TrimPrefix(r.URL.Path, "/image/assets/")
            data, err := s.assetStore.GetAsset(filePath)
            if err != nil || len(data) == 0 {
                http.NotFound(w, r)
                return
            }
            ct := mime.TypeByExtension(filepath.Ext(filePath))
            if ct == "" {
                ct = "application/octet-stream"
            }
            w.Header().Set("Content-Type", ct)
            w.WriteHeader(http.StatusOK)
            w.Write(data)
        }),
        DEFAULT_TIMEOUT, "",
    ),
)
```
Add import `"mime"` and `"path/filepath"` if not already present. Remove
`"net/http"` `FileServer` usage for this route (it may still be used for others).

**Step 10.** Add `assetStore *akdb.AssetStore` to `MainServer` struct and wire it
in via `NewMainServer(...)` arguments + a wire provider, following the same pattern
as `DatbaseConn` in `server/internal/akdb/db.go` and the existing wire graph
(look at how other providers are registered in `server/wire*.go`).

---

## Phase 5 — Dockerfile Cleanup

**Step 11.** In the `base` stage of `Dockerfile`, remove `assets` from the `mkdir`
and remove the `COPY ./static/assets /ak_chibi_assets/assets` line.

**Step 12.** In the `production` stage `ENTRYPOINT`, remove the `-image_assetdir`
flag — the new handler no longer reads from disk.

---

## Relevant Files

| File | Change |
|---|---|
| `.gitignore` | Add `data/.db-data/` |
| `db/migrations/<timestamp>_asset_store.up.sql` | New (via migrate CLI) |
| `db/migrations/<timestamp>_asset_store.down.sql` | New (via migrate CLI) |
| `server/tools/ingest_assets/main.go` | New — Go ingestion tool |
| `server/internal/operator/assets.go` | Reference: `SpineAssetMap`, `Iterate()`, `PlatformIndie*` fields on `SpineData` |
| `server/internal/akdb/db.go` | Add `AssetStore` + `ProvideAssetStore` here |
| `server/internal/server/server.go` (lines 136–143) | Replace FileServer with DB handler |
| `Dockerfile` | Remove `COPY ./static/assets` + update ENTRYPOINT |

## Verification

1. Run migration: `docker compose up migrations` → confirm `asset_files` table exists
2. Set env vars and run:
   `go run server/tools/ingest_assets/main.go -assetDir static/assets`
   → confirm row count matches expected file count
3. `curl -v http://localhost:8080/image/assets/characters/char_002_amiya/default/base/Front/build_char_002_amiya.png`
   → expect `200 OK` with `Content-Type: image/png`
4. Load the web app, open a room, set an operator → confirm spine animation renders
5. `docker compose build` → confirm production image is smaller (no `static/assets` baked in)

## Out of Scope
- In-memory caching layer
- Migrating `*_index.json` files to the DB
- Migrating `public/`, `spine/`, or `web_app/` static assets
- Serving compressed DDS/ASTC textures (separate task)
