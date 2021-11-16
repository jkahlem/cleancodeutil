package utils

import "testing"

func TestRevision(t *testing.T) {
	rev := NewRevision()
	for i := 0; i < 5; i++ {
		Subscribe(rev)
	}
	rev.SetOutdated()
}

// Helper functions

func Subscribe(rev Revision) {
	go rev.WaitUntilOutdated()
}
