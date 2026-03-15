package workshop_test

import (
	"testing"
	"time"

	"github.com/GoCodeAlone/workflow-plugin-steam/internal/workshop"
)

func TestVersionDB_GetMissing(t *testing.T) {
	db := workshop.NewVersionDB(t.TempDir() + "/.versions.json")
	_, found := db.Get("nonexistent")
	if found {
		t.Error("expected found=false for unknown publishedFileId")
	}
}

func TestVersionDB_SetAndGet(t *testing.T) {
	db := workshop.NewVersionDB(t.TempDir() + "/.versions.json")

	now := time.Now().UTC().Truncate(time.Second)
	record := workshop.VersionRecord{
		PublishedFileId: "12345",
		LastUpdatedAt:   now,
		InstalledAt:     now,
		Version:         "1.0.0",
		ItemDir:         "/tmp/workshop/12345",
	}
	if err := db.Set(record); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, found := db.Get("12345")
	if !found {
		t.Fatal("expected found=true after Set")
	}
	if got.PublishedFileId != "12345" {
		t.Errorf("PublishedFileId = %q, want 12345", got.PublishedFileId)
	}
	if got.Version != "1.0.0" {
		t.Errorf("Version = %q, want 1.0.0", got.Version)
	}
}

func TestVersionDB_ListAll(t *testing.T) {
	db := workshop.NewVersionDB(t.TempDir() + "/.versions.json")

	now := time.Now().UTC()
	db.Set(workshop.VersionRecord{PublishedFileId: "aaa", LastUpdatedAt: now, InstalledAt: now, Version: "1.0.0", ItemDir: "/tmp/a"})
	db.Set(workshop.VersionRecord{PublishedFileId: "bbb", LastUpdatedAt: now, InstalledAt: now, Version: "2.0.0", ItemDir: "/tmp/b"})

	all, err := db.ListAll()
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("len(all) = %d, want 2", len(all))
	}
}

func TestVersionDB_Remove(t *testing.T) {
	db := workshop.NewVersionDB(t.TempDir() + "/.versions.json")

	now := time.Now().UTC()
	db.Set(workshop.VersionRecord{PublishedFileId: "xyz", LastUpdatedAt: now, InstalledAt: now, Version: "1.0.0", ItemDir: "/tmp/xyz"})

	if err := db.Remove("xyz"); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	_, found := db.Get("xyz")
	if found {
		t.Error("expected found=false after Remove")
	}
}
