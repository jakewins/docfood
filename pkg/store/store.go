package store

import (
	"context"
	"fmt"
)

type Store interface {
	CreateSubscription(email string, allRestaurants bool, specificRestaurants []string, subType string, amount string,
		paymentMethod string) error
}

type MemStore struct {
}

func (m *MemStore) CreateSubscription(email string, allRestaurants bool, specificRestaurants []string,
	subType string, amount string, paymentMethod string) error {
	fmt.Printf("[store] Create subscription: email=%s, allRestaurants=%v, specificRestaurants=%v, subType=%s, " +
		"amount=%s, paymentMethod=%s\n",
		email, allRestaurants, specificRestaurants, subType, amount, paymentMethod)
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
