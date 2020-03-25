package store

import (
	"cloud.google.com/go/firestore"
	"context"
	"log"
)


type Firestore struct {
	client *firestore.Client
}

func (f *Firestore) CreateSubscription(sub Subscription) error {
	_, _, err := f.client.Collection("subscriptions").Add(context.Background(), sub)
	return err
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
