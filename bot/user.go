package bot

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/nicklaw5/helix"
)

// TODO: Use an LRU
var users = map[string]*User{}
var twitchUsers = map[string]helix.User{}

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

func (u *User) TakePoints(points uint64) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.Points = u.Points - points
	return updateUser(u)
}

func (u *User) TransferPoints(points uint64, userID string) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	u2, err := GetUser(userID)
	if err != nil {
		return err
	}

	// If insufficient balance. Transfer remaining balance.
	if u.Points < points {
		points = u.Points
	}
	u.Points = u.Points - points

	u2.Points = u2.Points + points

	jsonUser1, err := json.Marshal(u)
	if err != nil {
		return err
	}

	jsonUser2, err := json.Marshal(u2)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(USER_BUCKET)
		if err := bucket.Put([]byte(u.ID), jsonUser1); err != nil {
			return err
		}
		if err := bucket.Put([]byte(u2.ID), jsonUser2); err != nil {
			return err
		}
		return nil
	})

	return nil
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

func GetTwitchUserByName(name string) (helix.User, error) {
	var u helix.User

	if u, ok := twitchUsers[name]; ok {
		return u, nil
	}

	resp, err := helixClient.GetUsers(&helix.UsersParams{
		Logins: []string{name},
	})
	if err != nil {
		return u, err
	}

	if len(resp.Data.Users) == 0 {
		return u, fmt.Errorf("User with name '%s' was not found.", name)
	}

	twitchUsers[name] = resp.Data.Users[0]
	return resp.Data.Users[0], nil
}
