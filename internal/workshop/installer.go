package workshop

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// InstallOptions controls how a Workshop item is installed.
type InstallOptions struct {
	// PublishedFileId is the Steam Workshop file ID (used as directory name).
	PublishedFileId string
	// ZipData is the raw zip archive bytes to extract.
	ZipData []byte
	// InstallDir is the base directory; item is installed to <InstallDir>/<PublishedFileId>/.
	InstallDir string
	// ValidateRuleset runs the server-side validator after extraction.
	ValidateRuleset bool
}

// InstallResult holds the outcome of a successful Workshop item install.
type InstallResult struct {
	ItemDir       string
	Manifest      *WorkshopManifest
	AlreadyCurrent bool
}

// checksumFile stores the SHA-256 of the installed zip for idempotency checks.
const checksumFileName = ".checksum"

// Install extracts a Workshop zip archive to the install directory, validates
// the manifest, optionally validates the ruleset YAML, and returns the result.
//
// Calling Install twice with the same zip data is a no-op (AlreadyCurrent=true).
func Install(opts InstallOptions) (InstallResult, error) {
	if opts.PublishedFileId == "" {
		return InstallResult{}, fmt.Errorf("workshop installer: publishedFileId is required")
	}
	if len(opts.ZipData) == 0 {
		return InstallResult{}, fmt.Errorf("workshop installer: zip data is empty")
	}

	itemDir := filepath.Join(opts.InstallDir, opts.PublishedFileId)

	// Compute checksum of incoming zip.
	h := sha256.Sum256(opts.ZipData)
	checksum := hex.EncodeToString(h[:])

	// Check if already installed with same checksum (idempotency).
	checksumPath := filepath.Join(itemDir, checksumFileName)
	if existing, err := os.ReadFile(checksumPath); err == nil {
		if string(existing) == checksum {
			// Already installed at this version — read manifest and return.
			manifest, err := readManifestFromDir(itemDir)
			if err != nil {
				return InstallResult{}, fmt.Errorf("workshop installer: read cached manifest: %w", err)
			}
			return InstallResult{
				ItemDir:        itemDir,
				Manifest:       manifest,
				AlreadyCurrent: true,
			}, nil
		}
	}

	// Extract zip to itemDir.
	if err := os.MkdirAll(itemDir, 0o755); err != nil {
		return InstallResult{}, fmt.Errorf("workshop installer: create item dir: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(opts.ZipData), int64(len(opts.ZipData)))
	if err != nil {
		return InstallResult{}, fmt.Errorf("workshop installer: parse zip: %w", err)
	}

	for _, f := range zr.File {
		destPath := filepath.Join(itemDir, filepath.FromSlash(f.Name))
		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0o755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return InstallResult{}, fmt.Errorf("workshop installer: mkdir for %q: %w", f.Name, err)
		}
		rc, err := f.Open()
		if err != nil {
			return InstallResult{}, fmt.Errorf("workshop installer: open zip entry %q: %w", f.Name, err)
		}
		dst, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			return InstallResult{}, fmt.Errorf("workshop installer: create %q: %w", destPath, err)
		}
		if _, err := io.Copy(dst, rc); err != nil {
			dst.Close()
			rc.Close()
			return InstallResult{}, fmt.Errorf("workshop installer: extract %q: %w", f.Name, err)
		}
		dst.Close()
		rc.Close()
	}

	// Read manifest.
	manifest, err := readManifestFromDir(itemDir)
	if err != nil {
		return InstallResult{}, fmt.Errorf("workshop installer: %w", err)
	}

	// Optionally validate.
	if opts.ValidateRuleset {
		validResult, err := ValidatePackage(itemDir, false)
		if err != nil {
			return InstallResult{}, fmt.Errorf("workshop installer: validation error: %w", err)
		}
		if !validResult.Valid {
			return InstallResult{}, fmt.Errorf("workshop installer: package invalid: %v", validResult.Errors)
		}
	}

	// Write checksum for idempotency.
	if err := os.WriteFile(checksumPath, []byte(checksum), 0o644); err != nil {
		return InstallResult{}, fmt.Errorf("workshop installer: write checksum: %w", err)
	}

	return InstallResult{
		ItemDir:        itemDir,
		Manifest:       manifest,
		AlreadyCurrent: false,
	}, nil
}

// readManifestFromDir reads and parses manifest.json from a directory.
func readManifestFromDir(dir string) (*WorkshopManifest, error) {
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("manifest.json not found: %w", err)
	}
	var m WorkshopManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest.json: %w", err)
	}
	return &m, nil
}
