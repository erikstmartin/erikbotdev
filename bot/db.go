package bot

import (
	"os"
	"time"

	"github.com/boltdb/bolt"
)

var db *bolt.DB
var USER_BUCKET = []byte("Users")

func InitDatabase(file string, mode os.FileMode) error {
	var err error
	db, err = bolt.Open(file, mode, &bolt.Options{Timeout: 1 * time.Second})
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

	// Commit the transaction and check for error.
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
