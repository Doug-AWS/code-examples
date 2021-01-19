package main

import (
	"testing"
	"time"
)

func TestDummy(t *testing.T) {
	thisTime := time.Now()
	nowString := thisTime.Format("2006-01-02 15:04:05 Monday")
	t.Log("Starting unit test at " + nowString)

	t.Log("Dummy test complete")
}