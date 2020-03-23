package store

import (
	"cloud.google.com/go/firestore"
	"context"
	"log"
)


type Firestore struct {
	client *firestore.Client
}

func (f Firestore) CreateSubscription(email string, allRestaurants bool, specificRestaurants []string,
	subType string, amount string, paymentMethod string) error {

	panic("implement me")
}

var _ Store = &Firestore{}

func NewFirestore() *Firestore {
	projectID := "docfood"

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	return &Firestore{client:client}
}
