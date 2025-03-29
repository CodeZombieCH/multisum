package main

import (
	"bufio"
	"context"
	"crypto"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"slices"
	"strings"
)

type printMode int

const (
	ModeText printMode = iota
	ModeBinary
)

type ChecksumCalculatorConfig struct {
	// Checksum print mode
	printMode printMode

	// Hashes to compute
	hashes []crypto.Hash

	// Path to directory of files to be checksum'ed
	sourceDirectoryPath string
	// Path to directory where to create the checksum files
	targetDirectoryPath string
}

func (c *ChecksumCalculatorConfig) Print() {
	hashes := make([]string, 0)
	for _, h := range c.hashes {
		hashes = append(hashes, h.String())
	}

	var mode string
	switch c.printMode {
	case ModeText:
		mode = "text"
	case ModeBinary:
		mode = "binary"

	}

	log.Printf("creating checksum files (%s) in %s mode for directory \"%s\" writing to \"%s\"",
		strings.Join(hashes, ","),
		mode,
		c.sourceDirectoryPath,
		c.targetDirectoryPath)
}

// Create a single hash file per hash and file
type ChecksumCalculator struct {
	cfg     ChecksumCalculatorConfig
	scanner *scanner
	stats   stats
}

func NewChecksumCalculator(cfg ChecksumCalculatorConfig) *ChecksumCalculator {
	return &ChecksumCalculator{
		cfg: cfg,
	}
}

// Path to directory of files to be checksum'ed
func (s *ChecksumCalculator) Calculate(cfg ChecksumCalculatorConfig, sourceDirectoryPath string) error {
	// TODO: No longer needed
	if err := s.ensureIsGitSumRepo(cfg.targetDirectoryPath); err != nil {
		return err
	}

	// TODO: No full clean needed anymore, only check if files exist that would be overwritten
	// Delete repo workspace
	if err := s.resetChecksums(cfg.targetDirectoryPath); err != nil {
		return err
	}

	s.scanner = newScanner()
	s.scanner.StartScan(context.TODO(), sourceDirectoryPath)

	s.stats = newStats(s.scanner)
	s.stats.StartUpdate()

	checksumWriters, err := s.createWriters(cfg)
	if err != nil {
		return fmt.Errorf("failed to create checksum writers: %w", err)
	}

	// Create all checksum files
	if err := s.calcChecksums(sourceDirectoryPath, checksumWriters); err != nil {
		return err
	}

	s.scanner.StopScan()

	s.stats.StopUpdate()
	s.stats.PrintStats()

	return nil
}

func (s *ChecksumCalculator) createWriters(cfg ChecksumCalculatorConfig) ([]*checksumWriter, error) {

	checksumWriters := make([]*checksumWriter, 0)

	for _, h := range cfg.hashes {
		switch h {
		case crypto.MD5:
			checksumWriters = append(checksumWriters, newChecksumWriter(cfg.targetDirectoryPath, md5.New()))
		case crypto.SHA1:
			checksumWriters = append(checksumWriters, newChecksumWriter(cfg.targetDirectoryPath, sha1.New()))
		case crypto.SHA256:
			checksumWriters = append(checksumWriters, newChecksumWriter(cfg.targetDirectoryPath, sha256.New()))
		case crypto.SHA512:
			checksumWriters = append(checksumWriters, newChecksumWriter(cfg.targetDirectoryPath, sha512.New()))
		default:
			return nil, fmt.Errorf("unknown hash value %s", h)
		}
	}

	return checksumWriters, nil
}

func (s *ChecksumCalculator) calcChecksums(storePath string, checksumWriters []*checksumWriter) error {

	writers := make([]io.Writer, 0)
	for _, w := range checksumWriters {
		err := w.Open()
		if err != nil {
			return fmt.Errorf("failed to open checksum writer: %w", err)
		}
		writers = append(writers, w.hash)
	}

	writer := io.MultiWriter(writers...)

	fileSystem := os.DirFS(storePath)

	err := fs.WalkDir(fileSystem, ".", func(localPath string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("warning: %v", err)
			return fs.SkipDir
		}

		if d.IsDir() {
			if localPath == "." {
				// Nothing to do with the store directory itself
				return nil
			}
			if localPath == ".git" {
				// We do not wan't a `.git` directory in our checksum repo and overwrite it when generating checksums
				log.Printf("warning: ignoring git repository %v", path.Join(storePath, localPath))
				return fs.SkipDir
			}

			return nil
		}

		if !d.Type().IsRegular() {
			log.Printf("skipping non regular file %s", localPath)
			return nil
		}

		// WARNING
		// fails with symlinks!!!!!!!!!!!!!!!!!!!!!!!!!!!
		// WARNING
		// e.g. /home/mab/Dropbox/Projects/SunriseStats/venv/include/python3.4m

		absolutePath := path.Join(storePath, localPath)
		err = s.checksumFile(absolutePath, writer)
		if err != nil {
			return fmt.Errorf("calcChecksums: failed to calculate checksums: %w", err)
		}

		for _, w := range checksumWriters {
			err := w.WriteChecksum(localPath)
			if err != nil {
				return fmt.Errorf("failed to write checksum with checksum writer: %w", err)
			}
		}

		s.stats.IncrementActualCount()

		return nil
	})

	for _, w := range checksumWriters {
		err := w.Close()
		if err != nil {
			return fmt.Errorf("failed to close checksum writer: %w", err)
		}
	}

	return err
}

func (s *ChecksumCalculator) checksumFile(path string, writer io.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(writer, f); err != nil {
		return err
	}

	return nil
}

// Ensure the repo is either empty or contains only hashes
func (s *ChecksumCalculator) ensureIsGitSumRepo(repoPath string) error {
	repoFS := os.DirFS(repoPath)

	// Dry run
	err := fs.WalkDir(repoFS, ".", func(localPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("ensureIsGitSumRepo: failed to walk directory: %w", err)
		}
		if d.IsDir() {
			if localPath == ".git" {
				// Do not check the git repository
				return fs.SkipDir
			}

			// regular directory, should be checked
			return nil
		} else {
			if localPath == ".gitattributes" {
				// Do not delete the .gitattributes file
				return nil
			}

			// Dirty hack
			allowed := []string{"md5sums", "sha256sums"}
			if !slices.Contains(allowed, path.Base(path.Join(repoPath, localPath))) {
				return fmt.Errorf("ensureIsGitSumRepo: unexpected file %s found in repository", localPath)
			}

			return nil
		}
	})

	return err
}

func (s *ChecksumCalculator) resetChecksums(repoPath string) error {
	repoFS := os.DirFS(repoPath)

	err := fs.WalkDir(repoFS, ".", func(localPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("resetChecksums: failed to walk directory: %w", err)
		}
		if d.IsDir() {
			if localPath == "." {
				// Do not delete the repo directory itself
				return nil
			}
			if localPath == ".git" {
				// Do not delete the git repository
				return fs.SkipDir
			}

			err := os.RemoveAll(path.Join(repoPath, localPath))
			if err != nil {
				return fmt.Errorf("resetChecksums: failed to remove directory: %w", err)
			}
			return fs.SkipDir
		}

		if localPath == ".gitattributes" {
			// Do not delete the .gitattributes file
			return nil
		}

		if err := os.Remove(path.Join(repoPath, localPath)); err != nil {
			return fmt.Errorf("resetChecksums: failed to remove file: %w", err)
		}

		return nil
	})

	return err
}

type checksumWriter struct {
	hash     hash.Hash
	filename string
	// Writer for the checksum file
	writer *bufio.Writer
	file   *os.File
}

func newChecksumWriter(repoPath string, h hash.Hash) *checksumWriter {
	rawName := fmt.Sprintf("%T", h)
	hashName := strings.Split(strings.TrimPrefix(rawName, "*"), ".")[0]

	filename := path.Join(repoPath, strings.ToUpper(hashName)+"SUMS")

	return &checksumWriter{
		hash:     h,
		filename: filename,
	}
}

func (w *checksumWriter) Open() error {
	var err error
	w.file, err = os.Create(w.filename)
	if err != nil {
		return err
	}

	w.writer = bufio.NewWriter(w.file)

	return nil
}

func (w *checksumWriter) WriteChecksum(filePath string) error {
	_, err := w.writer.Write(fmt.Appendf(nil, "%x *%s\n", w.hash.Sum(nil), filePath)) // * for binary mode
	if err != nil {
		return err
	}
	// Prepare for next write
	w.hash.Reset()

	return nil
}

func (w *checksumWriter) Close() error {
	err := w.writer.Flush()
	if err != nil {
		return err
	}
	return w.file.Close()
}
