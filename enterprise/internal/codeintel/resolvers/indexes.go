package resolvers

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

// IndexesResolver wraps store.GetIndexes so that the underlying function can be
// invoked lazily and its results memoized.
type IndexesResolver struct {
	store store.Store
	opts  store.GetIndexesOptions
	once  sync.Once
	//
	Indexes    []store.Index
	TotalCount int
	NextOffset *int
	err        error
}

// NewIndexesResolver creates a new IndexesResolver which wil invoke store.GetIndexes
// with the given options.
func NewIndexesResolver(store store.Store, opts store.GetIndexesOptions) *IndexesResolver {
	return &IndexesResolver{store: store, opts: opts}
}

// Resolve ensures that store.GetIndexes has been invoked. This function returns the
// error from the invocation, if any. If the error is nil, then the resolver's Indexes,
// TotalCount, and NextOffset fields will be populated.
func (r *IndexesResolver) Resolve(ctx context.Context) error {
	r.once.Do(func() { r.err = r.resolve(ctx) })
	return r.err
}

func (r *IndexesResolver) resolve(ctx context.Context) error {
	indexes, totalCount, err := r.store.GetIndexes(ctx, r.opts)
	if err != nil {
		return err
	}

	r.Indexes = indexes
	r.NextOffset = nextOffset(r.opts.Offset, len(indexes), totalCount)
	r.TotalCount = totalCount
	return nil
}
