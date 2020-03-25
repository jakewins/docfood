package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"github.com/google/uuid"
	"os"
	"path"
)

type FileStore struct {
	Dir string
}

func (m *FileStore) CreateSubscription(sub Subscription) error {
	fmt.Printf("[store] Create subscription: %v\n", sub)
	id := uuid.New().String()

	var out bytes.Buffer
	err := json.NewEncoder(&out).Encode(sub)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path.Join(m.Dir, id), out.Bytes(), 0660)
}

func NewFileStore(dir string) (*FileStore, error) {
	err := os.MkdirAll(dir, 0770)
	if err != nil {
		return nil, err
	}
	return &FileStore{Dir:dir}, nil
}

var _ Store = &FileStore{}