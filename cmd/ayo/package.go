package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/build"
)

func newPackageCmd() *cobra.Command {
	var format string
	var version string

	cmd := &cobra.Command{
		Use:   "package <directory>",
		Short: "Create distributable archives of built agent binaries",
		Long: `Package agent binaries into distributable archives.

The package command:
1. Finds built agent binaries (from dist/ or build output)
2. Creates platform-specific archives (tar.gz or zip)
3. Generates SHA256 checksums for verification
4. Places packages in a releases/ directory

Usage:
  ayo package <directory>              Package with latest git tag version
  ayo package <directory> --version 1.0.0  Use specific version
  ayo package <directory> --format tar.gz  Specify archive format

Archives are named: {agent-name}-{version}-{os}-{arch}.{ext}
Example: myagent-1.0.0-linux-amd64.tar.gz`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			return runPackage(dir, format, version)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "auto", "Archive format: tar.gz, zip, or auto (default)")
	cmd.Flags().StringVarP(&version, "version", "v", "", "Version string (defaults to git tag)")

	return cmd
}

func runPackage(dir, format, version string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("error resolving directory: %w", err)
	}

	cfg, _, err := build.LoadConfigFromDir(absDir)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	agentName := cfg.Agent.Name

	if version == "" {
		if cfg.Agent.Version != "" {
			version = cfg.Agent.Version
		} else {
			version, err = getGitVersion(absDir)
			if err != nil {
				version = "0.0.0-dev"
			}
		}
	}

	binaries, err := findBuiltBinaries(absDir, agentName)
	if err != nil {
		return fmt.Errorf("error finding binaries: %w", err)
	}

	if len(binaries) == 0 {
		return fmt.Errorf("no built binaries found. Run 'ayo build %s' first", dir)
	}

	releasesDir := filepath.Join(absDir, "releases")
	if err := os.MkdirAll(releasesDir, 0755); err != nil {
		return fmt.Errorf("error creating releases directory: %w", err)
	}

	checksums := make(map[string]string)

	fmt.Printf("Packaging %s v%s...\n", agentName, version)

	for _, binary := range binaries {
		archiveFormat := format
		if archiveFormat == "auto" {
			if strings.Contains(binary.Path, "windows") {
				archiveFormat = "zip"
			} else {
				archiveFormat = "tar.gz"
			}
		}

		archiveName := fmt.Sprintf("%s-%s-%s-%s", agentName, version, binary.OS, binary.Arch)
		if archiveFormat == "zip" {
			archiveName += ".zip"
		} else {
			archiveName += ".tar.gz"
		}
		archivePath := filepath.Join(releasesDir, archiveName)

		fmt.Printf("  Creating %s...\n", archiveName)

		if err := createArchive(binary.Path, archivePath, archiveFormat); err != nil {
			return fmt.Errorf("error creating archive for %s: %w", binary.Path, err)
		}

		checksum, err := computeFileHash(archivePath)
		if err != nil {
			return fmt.Errorf("error computing checksum: %w", err)
		}
		checksums[archiveName] = checksum
	}

	checksumPath := filepath.Join(releasesDir, fmt.Sprintf("%s-%s.sha256", agentName, version))
	if err := writeChecksums(checksumPath, checksums); err != nil {
		return fmt.Errorf("error writing checksums: %w", err)
	}

	fmt.Printf("\n✓ Created %d packages in %s/\n", len(binaries), releasesDir)
	fmt.Printf("  Checksums: %s\n", filepath.Base(checksumPath))

	return nil
}

type BinaryInfo struct {
	Path string
	OS   string
	Arch string
}

func findBuiltBinaries(dir, agentName string) ([]BinaryInfo, error) {
	var binaries []BinaryInfo

	distDir := filepath.Join(dir, "dist")
	if _, err := os.Stat(distDir); err == nil {
		entries, err := os.ReadDir(distDir)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()

			if strings.HasPrefix(name, agentName+"-") {
				binaries = append(binaries, BinaryInfo{
					Path: filepath.Join(distDir, name),
					OS:   "",
					Arch: "",
				})
			}
		}
	}

	if len(binaries) > 0 {
		for i := range binaries {
			name := filepath.Base(binaries[i].Path)
			name = strings.TrimPrefix(name, agentName+"-")
			name = strings.TrimSuffix(name, filepath.Ext(name))
			if strings.Contains(name, ".") {
				name = strings.Split(name, ".")[0]
			}

			parts := strings.Split(name, "-")
			if len(parts) >= 2 {
				binaries[i].OS = parts[len(parts)-2]
				binaries[i].Arch = parts[len(parts)-1]
			}
		}
		return binaries, nil
	}

	buildDir := filepath.Join(dir, ".build", "bin")
	if _, err := os.Stat(buildDir); err == nil {
		entries, err := os.ReadDir(buildDir)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()

			if strings.HasPrefix(name, agentName) || !strings.Contains(name, "-") {
				binaries = append(binaries, BinaryInfo{
					Path: filepath.Join(buildDir, name),
					OS:   "",
					Arch: "",
				})
			}
		}
	}

	return binaries, nil
}

func createArchive(sourcePath, destPath, format string) error {
	switch format {
	case "tar.gz", "tgz":
		return createTarGz(sourcePath, destPath)
	case "zip":
		return createZip(sourcePath, destPath)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func createTarGz(sourcePath, destPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	gzw := gzip.NewWriter(dest)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	info, err := source.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = filepath.Base(sourcePath)

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, source)
	return err
}

func createZip(sourcePath, destPath string) error {
	archive, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	info, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	header, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:               filepath.Base(sourcePath),
		UncompressedSize64: uint64(info.Size()),
		UncompressedSize:   uint32(info.Size()),
		Modified:           info.ModTime(),
		Method:             zip.Deflate,
	})
	if err != nil {
		return err
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = io.Copy(header, source)
	return err
}

func computeFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func writeChecksums(path string, checksums map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for name, checksum := range checksums {
		if _, err := fmt.Fprintf(file, "%s  %s\n", checksum, name); err != nil {
			return err
		}
	}

	return nil
}

func getGitVersion(dir string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
