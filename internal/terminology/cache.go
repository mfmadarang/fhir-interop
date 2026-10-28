package terminology

import (
	"context"
	"time"
)

// where cached answers are kept, the store package has the Postgres version
type CacheStore interface {
	GetCached(system, code string) (res *Result, checkedAt time.Time, found bool, err error)
	PutCached(system, code string, res *Result) error
}

type codeValidator interface {
	ValidateCode(ctx context.Context, system, code string) (*Result, error)
}

// checks the cache before asking tx.fhir.org, and saves answers it gets back
type CachedClient struct {
	next  codeValidator
	store CacheStore
	ttl   time.Duration
	now   func() time.Time
}

func NewCachedClient(next codeValidator, store CacheStore, ttl time.Duration) *CachedClient {
	return &CachedClient{next: next, store: store, ttl: ttl, now: time.Now}
}

func (c *CachedClient) ValidateCode(ctx context.Context, system, code string) (*Result, error) {
	// a broken cache just means we ask tx.fhir.org like before
	if res, checkedAt, found, err := c.store.GetCached(system, code); err == nil && found {
		if c.now().Sub(checkedAt) < c.ttl {
			return res, nil
		}
	}

	res, err := c.next.ValidateCode(ctx, system, code)
	if err != nil {
		// errors are never cached, so the next call tries the network again
		return nil, err
	}

	c.store.PutCached(system, code, res)
	return res, nil
}
