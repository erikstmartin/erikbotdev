package bot

import (
	"encoding/json"
	"sync"

	"github.com/boltdb/bolt"
)

// TODO: Use an LRU
var users = map[string]*User{}

type User struct {
	ID          string         `json:"id"`
	DisplayName string         `json:"displayName"`
	Color       string         `json:"color"`
	Badges      map[string]int `json:"badges"`
	Points      uint64         `json:"points"`
	New         bool           `json:"-"`
	lock        sync.RWMutex
}

func (u *User) GivePoints(points uint64) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.Points = u.Points + points
	return updateUser(u)
}

func (u *User) Save() error {
	u.lock.Lock()
	defer u.lock.Unlock()

	return updateUser(u)
}

func updateUser(u *User) error {
	u.New = false

	buf, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(USER_BUCKET)
		err := bucket.Put([]byte(u.ID), buf)
		return err
	})
}

func GetUser(id string) (*User, error) {
	var u User

	if u, ok := users[id]; ok {
		return u, nil
	}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(USER_BUCKET)
		v := b.Get([]byte(id))
		if len(v) == 0 {
			u.New = true
			return nil
		}
		return json.Unmarshal(v, &u)
	})

	if err == nil {
		users[id] = &u
	}

	return &u, err
}
