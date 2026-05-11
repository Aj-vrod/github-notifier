package storagev0

import (
	"Aj-vrod/github-notifier/types"
	"sync"
	"testing"
	"time"
)

// BUG DOCUMENTATION: The Storage struct does not protect its registry map with a mutex.
// This causes race conditions when the poller goroutine calls GetAllSubscriptions()
// while the API server goroutine calls Subscribe() or Unsubscribe() simultaneously.
// Run with `go test -race` to detect the race condition.

func TestNewStorage(t *testing.T) {
	storage := NewStorage()
	if storage == nil {
		t.Fatal("NewStorage() returned nil")
	}
	if storage.registry == nil {
		t.Fatal("NewStorage() created storage with nil registry")
	}
}

func TestSubscribe(t *testing.T) {
	storage := NewStorage()
	prInfo := &types.PRInfo{
		URL:    "https://github.com/owner/repo/pull/1",
		Owner:  "owner",
		Repo:   "repo",
		Number: 1,
	}
	prState := types.PRState{
		Body:     "Test PR body",
		Comments: []types.Comment{},
		Commits:  []types.CommitNode{},
	}

	storage.Subscribe(prInfo, prState)

	subs := storage.GetAllSubscriptions()
	if len(subs) != 1 {
		t.Errorf("Expected 1 subscription, got %d", len(subs))
	}

	storedState, exists := subs[prInfo.URL]
	if !exists {
		t.Errorf("Subscription not found for URL %s", prInfo.URL)
	}
	if storedState.Body != prState.Body {
		t.Errorf("Expected body %s, got %s", prState.Body, storedState.Body)
	}
}

func TestUnsubscribe(t *testing.T) {
	storage := NewStorage()
	prInfo := &types.PRInfo{
		URL:    "https://github.com/owner/repo/pull/1",
		Owner:  "owner",
		Repo:   "repo",
		Number: 1,
	}
	prState := types.PRState{
		Body:     "Test PR body",
		Comments: []types.Comment{},
		Commits:  []types.CommitNode{},
	}

	storage.Subscribe(prInfo, prState)
	storage.Unsubscribe(prInfo)

	subs := storage.GetAllSubscriptions()
	if len(subs) != 0 {
		t.Errorf("Expected 0 subscriptions after unsubscribe, got %d", len(subs))
	}
}

func TestGetAllSubscriptions(t *testing.T) {
	storage := NewStorage()

	// Add multiple subscriptions
	for i := 1; i <= 3; i++ {
		prInfo := &types.PRInfo{
			URL:    "https://github.com/owner/repo/pull/" + string(rune(i)),
			Owner:  "owner",
			Repo:   "repo",
			Number: i,
		}
		prState := types.PRState{
			Body: "Test PR " + string(rune(i)),
		}
		storage.Subscribe(prInfo, prState)
	}

	subs := storage.GetAllSubscriptions()
	if len(subs) != 3 {
		t.Errorf("Expected 3 subscriptions, got %d", len(subs))
	}
}

// TestConcurrentSubscribeAndGetAll simulates the race condition between
// the API server calling Subscribe() and the poller calling GetAllSubscriptions().
// BUG: This test will fail with -race flag due to missing mutex protection.
func TestConcurrentSubscribeAndGetAll(t *testing.T) {
	storage := NewStorage()
	var wg sync.WaitGroup

	// Simulate API server subscribing PRs concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			prInfo := &types.PRInfo{
				URL:    "https://github.com/owner/repo/pull/" + string(rune(i)),
				Owner:  "owner",
				Repo:   "repo",
				Number: i,
			}
			prState := types.PRState{
				Body: "Concurrent test",
			}
			storage.Subscribe(prInfo, prState)
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Simulate poller reading all subscriptions concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = storage.GetAllSubscriptions()
			time.Sleep(1 * time.Millisecond)
		}
	}()

	wg.Wait()
}

// TestConcurrentUnsubscribeAndGetAll simulates the race condition between
// the API server calling Unsubscribe() and the poller calling GetAllSubscriptions().
// BUG: This test will fail with -race flag due to missing mutex protection.
func TestConcurrentUnsubscribeAndGetAll(t *testing.T) {
	storage := NewStorage()

	// Pre-populate with subscriptions
	for i := 0; i < 50; i++ {
		prInfo := &types.PRInfo{
			URL:    "https://github.com/owner/repo/pull/" + string(rune(i)),
			Owner:  "owner",
			Repo:   "repo",
			Number: i,
		}
		prState := types.PRState{
			Body: "Test",
		}
		storage.Subscribe(prInfo, prState)
	}

	var wg sync.WaitGroup

	// Simulate API server unsubscribing PRs concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			prInfo := &types.PRInfo{
				URL:    "https://github.com/owner/repo/pull/" + string(rune(i)),
				Owner:  "owner",
				Repo:   "repo",
				Number: i,
			}
			storage.Unsubscribe(prInfo)
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Simulate poller reading all subscriptions concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = storage.GetAllSubscriptions()
			time.Sleep(1 * time.Millisecond)
		}
	}()

	wg.Wait()
}

// TestMultipleConcurrentSubscribes simulates multiple API handlers calling Subscribe()
// simultaneously for different PRs.
// BUG: This test will fail with -race flag due to missing mutex protection.
func TestMultipleConcurrentSubscribes(t *testing.T) {
	storage := NewStorage()
	var wg sync.WaitGroup

	numGoroutines := 10
	subscribesPerGoroutine := 20

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < subscribesPerGoroutine; i++ {
				prInfo := &types.PRInfo{
					URL:    "https://github.com/owner/repo/pull/" + string(rune(goroutineID*100+i)),
					Owner:  "owner",
					Repo:   "repo",
					Number: goroutineID*100 + i,
				}
				prState := types.PRState{
					Body: "Concurrent subscribe test",
				}
				storage.Subscribe(prInfo, prState)
			}
		}(g)
	}

	wg.Wait()

	// Verify all subscriptions were added (non-deterministic due to race)
	subs := storage.GetAllSubscriptions()
	if len(subs) != numGoroutines*subscribesPerGoroutine {
		t.Logf("Warning: Expected %d subscriptions, got %d (race condition may have caused data loss)",
			numGoroutines*subscribesPerGoroutine, len(subs))
	}
}

// TestConcurrentMixedOperations simulates all operations happening simultaneously:
// Subscribe, Unsubscribe, and GetAllSubscriptions.
// BUG: This test will fail with -race flag due to missing mutex protection.
func TestConcurrentMixedOperations(t *testing.T) {
	storage := NewStorage()

	// Pre-populate
	for i := 0; i < 20; i++ {
		prInfo := &types.PRInfo{
			URL:    "https://github.com/owner/repo/pull/" + string(rune(i)),
			Owner:  "owner",
			Repo:   "repo",
			Number: i,
		}
		prState := types.PRState{
			Body: "Initial",
		}
		storage.Subscribe(prInfo, prState)
	}

	var wg sync.WaitGroup

	// Subscribing new PRs
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 20; i < 40; i++ {
			prInfo := &types.PRInfo{
				URL:    "https://github.com/owner/repo/pull/" + string(rune(i)),
				Owner:  "owner",
				Repo:   "repo",
				Number: i,
			}
			prState := types.PRState{
				Body: "New subscription",
			}
			storage.Subscribe(prInfo, prState)
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Unsubscribing existing PRs
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			prInfo := &types.PRInfo{
				URL:    "https://github.com/owner/repo/pull/" + string(rune(i)),
				Owner:  "owner",
				Repo:   "repo",
				Number: i,
			}
			storage.Unsubscribe(prInfo)
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Reading subscriptions (poller)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			_ = storage.GetAllSubscriptions()
			time.Sleep(500 * time.Microsecond)
		}
	}()

	wg.Wait()
}
