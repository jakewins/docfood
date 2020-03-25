package store

import (
	"context"
	"fmt"
)

type Subscription struct {
	Email string
	AllRestaurants bool
	SpecificRestaurants []string
	SubscriptionType string
	Amount string
	PaymentMethod string
}

type Store interface {
	CreateSubscription(sub Subscription) error
}

type MemStore struct {
}

func (m *MemStore) CreateSubscription(sub Subscription) error {
	fmt.Printf("[store] Create subscription: %v\n", sub)
	return nil
}

var _ Store = &MemStore{}

type key int
var storeKey key

func NewContext(ctx context.Context, s Store) context.Context {
	return context.WithValue(ctx, storeKey, s)
}

func FromContext(ctx context.Context) (Store, bool) {
	s, ok := ctx.Value(storeKey).(Store)
	return s, ok
}

func NewMemStore() *MemStore {
	return &MemStore{}
}
