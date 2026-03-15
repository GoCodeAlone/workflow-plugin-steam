package workshop

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// VersionRecord tracks the installed state of a Workshop item.
type VersionRecord struct {
	PublishedFileId string    `json:"publishedFileId"`
	LastUpdatedAt   time.Time `json:"lastUpdatedAt"`
	InstalledAt     time.Time `json:"installedAt"`
	Version         string    `json:"version"`
	ItemDir         string    `json:"itemDir"`
}

// VersionDB is a JSON-backed store of Workshop item version records.
// It is stored as a single JSON file at the configured path.
type VersionDB struct {
	path    string
	mu      sync.Mutex
	records map[string]VersionRecord
}

// NewVersionDB creates a VersionDB backed by a JSON file at dbPath.
// Existing data is loaded lazily on first access.
func NewVersionDB(dbPath string) *VersionDB {
	return &VersionDB{
		path:    dbPath,
		records: nil,
	}
}

// Get returns the VersionRecord for the given publishedFileId, if it exists.
func (db *VersionDB) Get(publishedFileId string) (VersionRecord, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if err := db.loadLocked(); err != nil {
		return VersionRecord{}, false
	}
	r, ok := db.records[publishedFileId]
	return r, ok
}

// Set inserts or updates a VersionRecord.
func (db *VersionDB) Set(r VersionRecord) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if err := db.loadLocked(); err != nil {
		return err
	}
	db.records[r.PublishedFileId] = r
	return db.saveLocked()
}

// ListAll returns all VersionRecords.
func (db *VersionDB) ListAll() ([]VersionRecord, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if err := db.loadLocked(); err != nil {
		return nil, err
	}
	out := make([]VersionRecord, 0, len(db.records))
	for _, r := range db.records {
		out = append(out, r)
	}
	return out, nil
}

// Remove deletes the VersionRecord for the given publishedFileId.
func (db *VersionDB) Remove(publishedFileId string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if err := db.loadLocked(); err != nil {
		return err
	}
	delete(db.records, publishedFileId)
	return db.saveLocked()
}

// loadLocked reads the JSON file into memory. Must be called with mu held.
func (db *VersionDB) loadLocked() error {
	if db.records != nil {
		return nil
	}
	db.records = map[string]VersionRecord{}
	data, err := os.ReadFile(db.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("versiondb: read %q: %w", db.path, err)
	}
	if err := json.Unmarshal(data, &db.records); err != nil {
		return fmt.Errorf("versiondb: parse %q: %w", db.path, err)
	}
	return nil
}

// saveLocked writes db.records to the JSON file. Must be called with mu held.
func (db *VersionDB) saveLocked() error {
	data, err := json.MarshalIndent(db.records, "", "  ")
	if err != nil {
		return fmt.Errorf("versiondb: marshal: %w", err)
	}
	if err := os.WriteFile(db.path, data, 0o644); err != nil {
		return fmt.Errorf("versiondb: write %q: %w", db.path, err)
	}
	return nil
}
