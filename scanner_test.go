package main

import (
	"context"
	"testing"
	"time"
)

func TestScanner(t *testing.T) {
	s := scanner{}
	s.StartScan(context.TODO(), "/home/mab/go")

	time.Sleep(100 * time.Millisecond)

	//t.Logf("total count: %d", s.totalCount)
	s.StopScan()
	time.Sleep(2 * time.Second)
	t.Logf("total count: %d", s.totalCount)

}
