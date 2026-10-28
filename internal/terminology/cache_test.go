package terminology

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeCacheStore struct {
	res       *Result
	checkedAt time.Time
	found     bool
	getErr    error
	puts      int
}

func (f *fakeCacheStore) GetCached(system, code string) (*Result, time.Time, bool, error) {
	return f.res, f.checkedAt, f.found, f.getErr
}

func (f *fakeCacheStore) PutCached(system, code string, res *Result) error {
	f.puts++
	f.res, f.checkedAt, f.found = res, time.Now(), true
	return nil
}

type fakeNext struct {
	calls int
	res   *Result
	err   error
}

func (f *fakeNext) ValidateCode(ctx context.Context, system, code string) (*Result, error) {
	f.calls++
	return f.res, f.err
}

func TestCachedClientUsesCache(t *testing.T) {
	store := &fakeCacheStore{res: &Result{Valid: true}, checkedAt: time.Now(), found: true}
	next := &fakeNext{res: &Result{Valid: false}}
	c := NewCachedClient(next, store, time.Hour)

	res, err := c.ValidateCode(context.Background(), SystemLOINC, "8867-4")
	if err != nil || !res.Valid {
		t.Fatalf("got %+v, %v, want the cached valid result", res, err)
	}
	if next.calls != 0 {
		t.Fatalf("network was called %d time(s), want 0", next.calls)
	}
}

func TestCachedClientExpired(t *testing.T) {
	store := &fakeCacheStore{res: &Result{Valid: false}, checkedAt: time.Now().Add(-2 * time.Hour), found: true}
	next := &fakeNext{res: &Result{Valid: true}}
	c := NewCachedClient(next, store, time.Hour)

	res, _ := c.ValidateCode(context.Background(), SystemLOINC, "8867-4")
	if !res.Valid || next.calls != 1 || store.puts != 1 {
		t.Fatalf("expected a fresh lookup that gets saved: res=%+v calls=%d puts=%d", res, next.calls, store.puts)
	}
}

func TestCachedClientMissSavesAnswer(t *testing.T) {
	store := &fakeCacheStore{}
	next := &fakeNext{res: &Result{Valid: true, Display: "Heart rate"}}
	c := NewCachedClient(next, store, time.Hour)

	c.ValidateCode(context.Background(), SystemLOINC, "8867-4")
	c.ValidateCode(context.Background(), SystemLOINC, "8867-4")
	if next.calls != 1 {
		t.Fatalf("network was called %d time(s), want 1", next.calls)
	}
}

func TestCachedClientDoesNotCacheErrors(t *testing.T) {
	store := &fakeCacheStore{}
	next := &fakeNext{err: errors.New("timeout")}
	c := NewCachedClient(next, store, time.Hour)

	if _, err := c.ValidateCode(context.Background(), SystemLOINC, "8867-4"); err == nil {
		t.Fatal("expected the network error to come back")
	}
	if store.puts != 0 {
		t.Fatal("an error should not be cached")
	}
}

func TestCachedClientBrokenCache(t *testing.T) {
	store := &fakeCacheStore{getErr: errors.New("db down")}
	next := &fakeNext{res: &Result{Valid: true}}
	c := NewCachedClient(next, store, time.Hour)

	res, err := c.ValidateCode(context.Background(), SystemLOINC, "8867-4")
	if err != nil || !res.Valid || next.calls != 1 {
		t.Fatalf("a broken cache should fall through to the network: res=%+v err=%v calls=%d", res, err, next.calls)
	}
}
