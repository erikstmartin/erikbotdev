package bot

import (
	"encoding/json"
	"os"
	"time"

	"go.etcd.io/bbolt"
)

var db *bbolt.DB
var USER_BUCKET = []byte("Users")
var FOLLOWER_BUCKET = []byte("Followers")
var COUNTER_BUCKET = []byte("Counters")

func IncrementCounter(counter string) (current uint64) {
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(COUNTER_BUCKET)
		v := b.Get([]byte(counter))

		if len(v) >= 1 {
			if err := json.Unmarshal(v, &current); err != nil {
				return err
			}
		}
		current++

		j, err := json.Marshal(current)
		if err != nil {
			return err
		}
		b.Put([]byte(counter), j)

		return nil
	})

	return
}

func ListCounters() []string {
	counters := make([]string, 0)

	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(COUNTER_BUCKET)

		b.ForEach(func(k, v []byte) error {
			counters = append(counters, string(k))
			return nil
		})
		return nil
	})

	return counters
}

func InitDatabase(file string, mode os.FileMode) error {
	var err error
	db, err = bbolt.Open(file, mode, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	// Initialize (create any needed buckets, ensure they don't exists first)
	// Start a writable transaction.
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Use the transaction...
	_, err = tx.CreateBucketIfNotExists(USER_BUCKET)
	if err != nil {
		return err
	}

	_, err = tx.CreateBucketIfNotExists(COUNTER_BUCKET)
	if err != nil {
		return err
	}

	_, err = tx.CreateBucketIfNotExists(FOLLOWER_BUCKET)
	if err != nil {
		return err
	}

	// Commit the transaction and check for error.
	if err := tx.Commit(); err != nil {
		return err
	}

	go func() {
		UpdateFollowers()
		t := time.NewTicker(5 * time.Minute)
		for range t.C {
			UpdateFollowers()
		}
	}()

	return nil
}
