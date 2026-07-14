package httpapi

import (
	"context"
	"sync"
)

type turnKey struct {
	TenantID       string
	UserID         string
	ConversationID string
}

type turnLockEntry struct {
	token chan struct{}
	refs  int
}

type turnLocker struct {
	mu      sync.Mutex
	entries map[turnKey]*turnLockEntry
}

func newTurnLocker() *turnLocker {
	return &turnLocker{entries: make(map[turnKey]*turnLockEntry)}
}

func (l *turnLocker) lock(ctx context.Context, key turnKey) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	l.mu.Lock()
	entry := l.entries[key]
	if entry == nil {
		entry = &turnLockEntry{token: make(chan struct{}, 1)}
		entry.token <- struct{}{}
		l.entries[key] = entry
	}
	entry.refs++
	l.mu.Unlock()

	select {
	case <-ctx.Done():
		l.releaseReference(key, entry)
		return nil, ctx.Err()
	case <-entry.token:
		return func() {
			entry.token <- struct{}{}
			l.releaseReference(key, entry)
		}, nil
	}
}

func (l *turnLocker) releaseReference(key turnKey, entry *turnLockEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry.refs--
	if entry.refs == 0 && l.entries[key] == entry {
		delete(l.entries, key)
	}
}
