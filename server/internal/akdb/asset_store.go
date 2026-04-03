package akdb

import (
	"crypto/sha256"
	"encoding/binary"
)

type AssetStore struct {
	db *DatbaseConn
}

func NewAssetStore(db *DatbaseConn) *AssetStore {
	return &AssetStore{db: db}
}

func ProvideAssetStore(db *DatbaseConn) *AssetStore {
	return NewAssetStore(db)
}

func hashFilePath(filePath string) int64 {
	sum := sha256.Sum256([]byte(filePath))
	return int64(binary.BigEndian.Uint64(sum[:8]))
}

type assetRow struct {
	Data []byte
}

func (s *AssetStore) GetAsset(filePath string) ([]byte, error) {
	var row assetRow
	err := s.db.DefaultDB.Raw(
		"SELECT data FROM asset_files WHERE file_path_hash = ?",
		hashFilePath(filePath),
	).Scan(&row).Error
	return row.Data, err
}
