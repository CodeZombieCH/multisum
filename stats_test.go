package main

import (
	"log"
	"testing"
)

type scannerMock struct {
	isScanning    bool
	getTotalCount int
}

func (s *scannerMock) IsScanning() bool {
	return s.isScanning
}
func (s *scannerMock) GetTotalCount() int {
	return s.getTotalCount
}

func TestStats(t *testing.T) {

	scanner := scannerMock{}

	stats := newStats(&scanner)

	log.Println()

	scanner.isScanning = true
	stats.PrintStats()

	scanner.isScanning = false
	scanner.getTotalCount = 2000
	stats.PrintStats()
}
