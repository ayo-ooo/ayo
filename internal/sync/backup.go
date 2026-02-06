package sync

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/version"
)

// BackupManifest contains metadata about a backup.
type BackupManifest struct {
	Version   string    `json:"version"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Machine   string    `json:"machine,omitempty"`
	Type      string    `json:"type"` // "manual" or "auto"
	Files     int       `json:"files"`
	Size      int64     `json:"size"`
}

// BackupInfo provides information about a backup file.
type BackupInfo struct {
	Name      string
	Path      string
	CreatedAt time.Time
	Size      int64
	Type      string
}

// BackupDir returns the backup storage directory.
func BackupDir() string {
	return filepath.Join(paths.DataDir(), "backups")
}

// ManualBackupDir returns the manual backup directory.
func ManualBackupDir() string {
	return filepath.Join(BackupDir(), "manual")
}

// AutoBackupDir returns the auto backup directory.
func AutoBackupDir() string {
	return filepath.Join(BackupDir(), "auto")
}

// CreateBackup creates a new backup with the given name.
// If name is empty, generates a timestamped name.
func CreateBackup(name string) (*BackupInfo, error) {
	if name == "" {
		name = time.Now().Format("2006-01-02-150405")
	}

	backupDir := ManualBackupDir()
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("create backup directory: %w", err)
	}

	backupPath := filepath.Join(backupDir, name+".tar.gz")

	// Create backup file
	file, err := os.Create(backupPath)
	if err != nil {
		return nil, fmt.Errorf("create backup file: %w", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	var totalFiles int
	var totalSize int64

	// Backup sandbox directory
	sandboxDir := SandboxDir()
	if _, err := os.Stat(sandboxDir); err == nil {
		files, size, err := addDirToTar(tarWriter, sandboxDir, "sandbox")
		if err != nil {
			return nil, fmt.Errorf("backup sandbox: %w", err)
		}
		totalFiles += files
		totalSize += size
	}

	// Backup config directory
	configDir := paths.ConfigDir()
	if _, err := os.Stat(configDir); err == nil {
		files, size, err := addDirToTar(tarWriter, configDir, "config")
		if err != nil {
			return nil, fmt.Errorf("backup config: %w", err)
		}
		totalFiles += files
		totalSize += size
	}

	// Backup data directory (excluding sandbox and backups)
	dataDir := paths.DataDir()
	if _, err := os.Stat(dataDir); err == nil {
		files, size, err := addDirToTarExcluding(tarWriter, dataDir, "data", []string{"sandbox", "backups"})
		if err != nil {
			return nil, fmt.Errorf("backup data: %w", err)
		}
		totalFiles += files
		totalSize += size
	}

	// Write manifest
	hostname, _ := os.Hostname()
	manifest := BackupManifest{
		Version:   version.Version,
		Name:      name,
		CreatedAt: time.Now(),
		Machine:   hostname,
		Type:      "manual",
		Files:     totalFiles,
		Size:      totalSize,
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal manifest: %w", err)
	}

	header := &tar.Header{
		Name:    "manifest.json",
		Mode:    0644,
		Size:    int64(len(manifestData)),
		ModTime: time.Now(),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("write manifest header: %w", err)
	}
	if _, err := tarWriter.Write(manifestData); err != nil {
		return nil, fmt.Errorf("write manifest: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		return nil, fmt.Errorf("stat backup: %w", err)
	}

	return &BackupInfo{
		Name:      name,
		Path:      backupPath,
		CreatedAt: time.Now(),
		Size:      fileInfo.Size(),
		Type:      "manual",
	}, nil
}

// ListBackups returns all available backups sorted by creation time (newest first).
func ListBackups() ([]BackupInfo, error) {
	var backups []BackupInfo

	// List manual backups
	manualDir := ManualBackupDir()
	if entries, err := os.ReadDir(manualDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			name := strings.TrimSuffix(entry.Name(), ".tar.gz")
			backups = append(backups, BackupInfo{
				Name:      name,
				Path:      filepath.Join(manualDir, entry.Name()),
				CreatedAt: info.ModTime(),
				Size:      info.Size(),
				Type:      "manual",
			})
		}
	}

	// List auto backups
	autoDir := AutoBackupDir()
	if entries, err := os.ReadDir(autoDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			name := strings.TrimSuffix(entry.Name(), ".tar.gz")
			backups = append(backups, BackupInfo{
				Name:      name,
				Path:      filepath.Join(autoDir, entry.Name()),
				CreatedAt: info.ModTime(),
				Size:      info.Size(),
				Type:      "auto",
			})
		}
	}

	// Sort by creation time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// GetBackup returns information about a specific backup.
func GetBackup(name string) (*BackupInfo, error) {
	backups, err := ListBackups()
	if err != nil {
		return nil, err
	}

	for _, b := range backups {
		if b.Name == name {
			return &b, nil
		}
	}

	return nil, fmt.Errorf("backup not found: %s", name)
}

// GetManifest reads the manifest from a backup file.
func GetManifest(backupPath string) (*BackupManifest, error) {
	file, err := os.Open(backupPath)
	if err != nil {
		return nil, fmt.Errorf("open backup: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("decompress backup: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read backup: %w", err)
		}

		if header.Name == "manifest.json" {
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("read manifest: %w", err)
			}

			var manifest BackupManifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				return nil, fmt.Errorf("parse manifest: %w", err)
			}

			return &manifest, nil
		}
	}

	return nil, fmt.Errorf("manifest not found in backup")
}

// RestoreBackup restores from a backup.
// Creates a safety backup first if createSafetyBackup is true.
func RestoreBackup(name string, createSafetyBackup bool) error {
	backup, err := GetBackup(name)
	if err != nil {
		return err
	}

	// Create safety backup first
	if createSafetyBackup {
		safetyName := "pre-restore-" + time.Now().Format("2006-01-02-150405")
		if _, err := CreateBackup(safetyName); err != nil {
			return fmt.Errorf("create safety backup: %w", err)
		}
	}

	file, err := os.Open(backup.Path)
	if err != nil {
		return fmt.Errorf("open backup: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("decompress backup: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	// Map of prefix to target directory
	prefixMap := map[string]string{
		"sandbox/": SandboxDir(),
		"config/":  paths.ConfigDir(),
		"data/":    paths.DataDir(),
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read backup: %w", err)
		}

		// Skip manifest
		if header.Name == "manifest.json" {
			continue
		}

		// Find target directory
		var targetPath string
		for prefix, dir := range prefixMap {
			if strings.HasPrefix(header.Name, prefix) {
				relPath := strings.TrimPrefix(header.Name, prefix)
				targetPath = filepath.Join(dir, relPath)
				break
			}
		}

		if targetPath == "" {
			continue // Unknown prefix, skip
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("create parent directory: %w", err)
			}
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("create file %s: %w", targetPath, err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("write file %s: %w", targetPath, err)
			}
			outFile.Close()
		}
	}

	return nil
}

// ExportBackup exports a backup to a portable archive at the given path.
func ExportBackup(name, destPath string) error {
	backup, err := GetBackup(name)
	if err != nil {
		return err
	}

	// Copy the backup file
	src, err := os.Open(backup.Path)
	if err != nil {
		return fmt.Errorf("open backup: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create export file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copy backup: %w", err)
	}

	return nil
}

// ImportBackup imports a backup from an external archive.
func ImportBackup(srcPath string) (*BackupInfo, error) {
	// Read manifest to get name
	manifest, err := GetManifest(srcPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	// Create import directory
	backupDir := ManualBackupDir()
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("create backup directory: %w", err)
	}

	// Generate unique name if exists
	name := manifest.Name
	destPath := filepath.Join(backupDir, name+".tar.gz")
	for i := 1; ; i++ {
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			break
		}
		name = fmt.Sprintf("%s-%d", manifest.Name, i)
		destPath = filepath.Join(backupDir, name+".tar.gz")
	}

	// Copy file
	src, err := os.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("open source: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("create dest: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("copy backup: %w", err)
	}

	fileInfo, err := os.Stat(destPath)
	if err != nil {
		return nil, fmt.Errorf("stat backup: %w", err)
	}

	return &BackupInfo{
		Name:      name,
		Path:      destPath,
		CreatedAt: time.Now(),
		Size:      fileInfo.Size(),
		Type:      "manual",
	}, nil
}

// PruneAutoBackups removes old auto backups, keeping only the most recent count.
func PruneAutoBackups(keepCount int) (int, error) {
	autoDir := AutoBackupDir()
	entries, err := os.ReadDir(autoDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read auto backup dir: %w", err)
	}

	var backups []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tar.gz") {
			backups = append(backups, entry)
		}
	}

	if len(backups) <= keepCount {
		return 0, nil
	}

	// Sort by mod time (oldest first)
	sort.Slice(backups, func(i, j int) bool {
		infoI, _ := backups[i].Info()
		infoJ, _ := backups[j].Info()
		if infoI == nil || infoJ == nil {
			return false
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// Remove oldest
	toRemove := len(backups) - keepCount
	removed := 0
	for i := 0; i < toRemove; i++ {
		path := filepath.Join(autoDir, backups[i].Name())
		if err := os.Remove(path); err == nil {
			removed++
		}
	}

	return removed, nil
}

// Helper functions

func addDirToTar(tw *tar.Writer, srcDir, prefix string) (int, int64, error) {
	return addDirToTarExcluding(tw, srcDir, prefix, nil)
}

func addDirToTarExcluding(tw *tar.Writer, srcDir, prefix string, exclude []string) (int, int64, error) {
	var files int
	var size int64

	excludeSet := make(map[string]bool)
	for _, e := range exclude {
		excludeSet[e] = true
	}

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return nil
		}

		// Check exclusions (top-level only)
		if excludeSet[relPath] {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create tar header
		tarPath := filepath.Join(prefix, relPath)
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return nil
		}
		header.Name = tarPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return nil // Skip files we can't open
			}
			defer file.Close()

			written, err := io.Copy(tw, file)
			if err != nil {
				return err
			}
			size += written
			files++
		}

		return nil
	})

	return files, size, err
}
