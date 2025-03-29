package main

import (
	"log"
	"os"
	"sync/atomic"
	"time"
)

// Scanner for the total count of files
type totalCountScanner interface {
	IsScanning() bool
	GetTotalCount() int
}

type stats struct {
	actualCount atomic.Int32
	scanner     totalCountScanner
	quit        chan struct{}

	log *log.Logger
}

func newStats(scanner totalCountScanner) stats {
	return stats{
		actualCount: atomic.Int32{},
		scanner:     scanner,
		log:         log.New(os.Stderr, "", 0),
	}
}

func (s *stats) GetActualCount() int32 {
	return s.actualCount.Load()
}

func (s *stats) IncrementActualCount() {
	s.actualCount.Add(1)
}

func (s *stats) PrintStats() {
	actualCount := s.actualCount.Load()

	if s.scanner.IsScanning() {
		s.log.Printf("\033[1A\033[Kstatus: %d/?, scan in progress...", actualCount)
	} else {
		percentage := 100.0 / float32(s.scanner.GetTotalCount()) * float32(actualCount)
		s.log.Printf("\033[1A\033[Kstatus: %d/%d (%.2f%%)", actualCount, s.scanner.GetTotalCount(), percentage)
	}
}

// Start periodic update of progress
func (s *stats) StartUpdate() {
	ticker := time.NewTicker(100 * time.Millisecond)
	s.quit = make(chan struct{})

	// Initial print
	log.Println()
	s.PrintStats()

	go func() {
		for {
			select {
			case <-ticker.C:
				s.PrintStats()
			case <-s.quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *stats) StopUpdate() {
	close(s.quit)
}
