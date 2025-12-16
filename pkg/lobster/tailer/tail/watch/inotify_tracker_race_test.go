// Copyright (c) 2025 Naver Corporation. All rights reserved.

package watch

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

func TestInotifyTrackerRaceFile(t *testing.T) {
	var (
		numOfTests       = 300
		numOfFileWriteOp = 30

		errors = make(chan error, numOfTests*3) // create/write/watch file
	)

	// Step 0. Create temp directory
	tmpDir, err := os.MkdirTemp("", "inotify_tracker_race_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	for i := range numOfTests {
		var (
			wg       = sync.WaitGroup{}
			onlyOnce = sync.Once{}
			fname    = filepath.Join(tmpDir, strconv.Itoa(i))
		)

		wg.Add(1)

		// Step 1. Create file
		if err := os.WriteFile(fname, []byte("test"), 0644); err != nil {
			errors <- err
			return
		}

		// Step 2. Watch file
		errors <- Watch(fname)

		// Step 3. Run receiver to process watch
		watchDone := make(chan struct{})
		go runReceiver(fname, watchDone)

		// Step 4. Destroy receiver to simulate hang
		close(watchDone)

		// Step 5. Write file without the receiver
		for k := range numOfFileWriteOp {
			// Step 5-1. Generate write events
			if err := os.WriteFile(fname, []byte(strconv.Itoa(k)), 0644); err != nil {
				errors <- err
				return
			}

			// Step 5-2. Remove watch while writing events without an active receiver
			onlyOnce.Do(func() {
				go func() {
					defer wg.Done()
					// Step 5-3. Remove watch
					Cleanup(fname)
					// Step 5-4. drain events to speed up deletion in the test
					drainEvents(fname)
				}()
			})
		}

		wg.Wait()

		// Step 6. Verify the events channel is cleaned up to simulate hang recovery(close, drain)
		if Events(fname) != nil {
			t.Errorf("Events channel should be deleted for %s", fname)
			t.Fail()
		}
	}

	close(errors)
	for err := range errors {
		if err != nil {
			t.Errorf("Errors: %v", err)
		}
	}
}

func runReceiver(fname string, watchDone chan struct{}) {
	for {
		select {
		case <-Events(fname):
		case <-watchDone:
		}
	}
}

func drainEvents(fname string) {
	ch := Events(fname)

	if ch == nil {
		return
	}

	for range ch {
	}
}
