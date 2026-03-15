package workshop

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	defaultMaxTotalBytes   = 100 * 1024 * 1024 // 100 MB
	defaultMaxPreviewBytes = 1 * 1024 * 1024   // 1 MB
)

// PackageOptions controls the workshop package build.
type PackageOptions struct {
	// SourceDir is the directory containing manifest.json and item files.
	SourceDir string
	// OutputPath is the destination .steamworkshop zip file path.
	OutputPath string
	// MaxTotalBytes is the maximum allowed total uncompressed size (default 100 MB).
	MaxTotalBytes int64
	// MaxPreviewBytes is the maximum allowed size for the preview image (default 1 MB).
	MaxPreviewBytes int64
}

// BuildPackage validates a workshop item directory and produces a .steamworkshop zip archive.
func BuildPackage(opts PackageOptions) error {
	if opts.MaxTotalBytes <= 0 {
		opts.MaxTotalBytes = defaultMaxTotalBytes
	}
	if opts.MaxPreviewBytes <= 0 {
		opts.MaxPreviewBytes = defaultMaxPreviewBytes
	}

	// 1. Read and validate manifest.json.
	manifestPath := filepath.Join(opts.SourceDir, "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("workshop packager: read manifest.json: %w", err)
	}
	var manifest WorkshopManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("workshop packager: parse manifest.json: %w", err)
	}
	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("workshop packager: invalid manifest: %w", err)
	}

	// 2. Check preview image size if specified.
	if manifest.PreviewImagePath != "" {
		previewFull := filepath.Join(opts.SourceDir, manifest.PreviewImagePath)
		info, err := os.Stat(previewFull)
		if err == nil { // only check if preview exists
			if info.Size() > opts.MaxPreviewBytes {
				return fmt.Errorf("workshop packager: preview image %q is %d bytes, maximum is %d",
					manifest.PreviewImagePath, info.Size(), opts.MaxPreviewBytes)
			}
		}
	}

	// 3. Walk the source directory, check total size, build zip.
	var totalSize int64
	type fileEntry struct {
		relPath  string
		fullPath string
	}
	var entries []fileEntry

	err = filepath.Walk(opts.SourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(opts.SourceDir, path)
		if err != nil {
			return err
		}
		totalSize += info.Size()
		if totalSize > opts.MaxTotalBytes {
			return fmt.Errorf("workshop packager: total content size exceeds maximum of %d bytes", opts.MaxTotalBytes)
		}
		entries = append(entries, fileEntry{relPath: rel, fullPath: path})
		return nil
	})
	if err != nil {
		return err
	}

	// 4. Write zip archive.
	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0o755); err != nil {
		return fmt.Errorf("workshop packager: create output dir: %w", err)
	}
	f, err := os.Create(opts.OutputPath)
	if err != nil {
		return fmt.Errorf("workshop packager: create output file: %w", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	for _, entry := range entries {
		w, err := zw.Create(filepath.ToSlash(entry.relPath))
		if err != nil {
			return fmt.Errorf("workshop packager: zip create entry %q: %w", entry.relPath, err)
		}
		src, err := os.Open(entry.fullPath)
		if err != nil {
			return fmt.Errorf("workshop packager: open %q: %w", entry.fullPath, err)
		}
		if _, err := io.Copy(w, src); err != nil {
			src.Close()
			return fmt.Errorf("workshop packager: write %q to zip: %w", entry.relPath, err)
		}
		src.Close()
	}

	return nil
}
