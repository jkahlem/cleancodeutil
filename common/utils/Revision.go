package utils

import "sync"

type revision struct {
	subscriber int
	onOutdated chan bool
	outdated   bool
	mutex      sync.Mutex
}

type Revision interface {
	// Wait until the revision is outdated.
	WaitUntilOutdated()
	// Sets the revision outdated.
	SetOutdated()
}

func NewRevision() Revision {
	rev := revision{}
	rev.onOutdated = make(chan bool)
	return &rev
}

// Blocks the thread until the version is outdated
func (rev *revision) WaitUntilOutdated() {
	rev.mutex.Lock()
	if !rev.outdated {
		rev.subscriber++
		rev.mutex.Unlock()
		<-rev.onOutdated
	} else {
		rev.mutex.Unlock()
	}
}

// Releases all subscribers and sets the version as outdated
func (rev *revision) SetOutdated() {
	rev.mutex.Lock()
	defer rev.mutex.Unlock()

	rev.outdated = true
	for i := 0; i < rev.subscriber; i++ {
		rev.onOutdated <- true
	}
	rev.subscriber = 0
}
