package main

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path"
)

type scanner struct {
	cancel     context.CancelFunc
	totalCount int
}

func newScanner() *scanner {
	return &scanner{
		totalCount: -1,
	}
}

func (s *scanner) IsScanning() bool {
	return s.cancel != nil && s.totalCount == -1
}

func (s *scanner) GetTotalCount() int {
	return s.totalCount
}

func (s *scanner) StartScan(ctx context.Context, storePath string) {
	cancelCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		defer cancel()
		err := s.scan(cancelCtx, storePath)
		if err != nil {
			log.Printf("scan failed: %v\n", err)
		}
	}()
}

func (s *scanner) StopScan() {
	s.cancel()
}

func (s *scanner) scan(ctx context.Context, storePath string) error {
	fileSystem := os.DirFS(storePath)

	counter := 0
	err := fs.WalkDir(fileSystem, ".", func(localPath string, d fs.DirEntry, err error) error {
		// TODO: Check if cancellation request
		if err := ctx.Err(); err != nil {
			log.Printf("cancelled")
			return err
		}

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
			return nil
		}

		counter++
		return nil
	})

	if err != nil {
		return err
	}

	s.totalCount = counter

	return nil
}
