package store

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/devdinu/gcloud-client/gcloud"
)

type DB struct {
	*bolt.DB
}

func (db DB) CreateBuckets(names []string) error {
	return db.Update(func(tx *bolt.Tx) error {
		for _, n := range names {
			_, err := tx.CreateBucketIfNotExists([]byte(n))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db DB) Save(ctx context.Context, instances <-chan gcloud.Instance, wg *sync.WaitGroup) error {
	defer wg.Done()

	for inst := range instances {
		err := db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(inst.Project))
			if b == nil {
				return fmt.Errorf("[DB] save instances failed for bucket %s inst: %+v", inst.Project, inst)
			}
			var data bytes.Buffer
			if err := gob.NewEncoder(&data).Encode(inst); err != nil {
				return err
			}
			return b.Put([]byte(inst.Name), data.Bytes())
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (db DB) Search(ctx context.Context, pattern string) ([]gcloud.Instance, error) {
	var insts []gcloud.Instance
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("wonderful-maps")).Cursor()
		for k, v := c.Seek([]byte(pattern)); k != nil && bytes.HasPrefix(k, []byte(pattern)); k, v = c.Next() {
			var data bytes.Buffer
			gob.NewDecoder(&data).Decode(v)
			fmt.Println("found: ", string(k), data.String())
		}
		return nil
	})
	return insts, nil
}

func NewDB() (DB, error) {
	db, err := bolt.Open("hosts.db", 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return DB{}, err
	}
	return DB{db}, nil
}
