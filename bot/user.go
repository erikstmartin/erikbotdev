package bot

import (
	"encoding/json"
	"fmt"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/nicklaw5/helix"
	"go.etcd.io/bbolt"
)

func init() {
	users, _ = lru.New(100)
	twitchUsers, _ = lru.New(100)
}

var users *lru.Cache
var twitchUsers *lru.Cache

var userID string
var mainChannel string

type User struct {
	ID          string         `json:"id"`
	DisplayName string         `json:"displayName"`
	Color       string         `json:"color"`
	Badges      map[string]int `json:"badges"`
	Points      uint64         `json:"points"`
	New         bool           `json:"-"`
	IsFollower  bool           `json:"isFollower"`
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
	return db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(USER_BUCKET)
		if err := bucket.Put([]byte(u.ID), jsonUser1); err != nil {
			return err
		}
		if err := bucket.Put([]byte(u2.ID), jsonUser2); err != nil {
			return err
		}
		return nil
	})
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

	return db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(USER_BUCKET)
		err := bucket.Put([]byte(u.ID), buf)
		return err
	})
}

func GetUser(id string) (*User, error) {
	var u User

	if u, ok := users.Get(id); ok {
		return u.(*User), nil
	}

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(USER_BUCKET)
		v := b.Get([]byte(id))
		if len(v) == 0 {
			u.ID = id
			u.New = true
			return nil
		}

		return json.Unmarshal(v, &u)
	})

	if err == nil {
		users.Add(id, &u)
	}

	return &u, err
}

func GetUserByName(name string) (*User, error) {
	if u, ok := twitchUsers.Get(name); ok {
		return GetUser(u.(helix.User).ID)
	}

	resp, err := helixClient.GetUsers(&helix.UsersParams{
		Logins: []string{name},
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Data.Users) == 0 {
		return nil, fmt.Errorf("User with name '%s' was not found.", name)
	}

	twitchUsers.Add(name, resp.Data.Users[0])
	return GetUser(resp.Data.Users[0].ID)
}

func UpdateFollowers() error {
	fmt.Println("Update of followers started.")
	defer fmt.Println("Update of followers finished.")

	err := db.Update(func(tx *bbolt.Tx) error {
		if err := tx.DeleteBucket(FOLLOWER_BUCKET); err != nil {
			return err
		}
		if _, err := tx.CreateBucket(FOLLOWER_BUCKET); err != nil {
			return err
		}
		followers := tx.Bucket(FOLLOWER_BUCKET)

		cursor := ""
		for {
			resp, err := helixClient.GetUsersFollows(&helix.UsersFollowsParams{After: cursor, First: 100, ToID: getUserID()})
			if err != nil {
				return err
			}

			for _, f := range resp.Data.Follows {
				j, err := json.Marshal(f)
				if err != nil {
					return err
				}

				if err := followers.Put([]byte(f.FromID), j); err != nil {
					return err
				}
			}

			if len(resp.Data.Follows) < 100 {
				break
			}
			cursor = resp.Data.Pagination.Cursor
		}

		return nil
	})

	err = db.Update(func(tx *bbolt.Tx) error {
		users := tx.Bucket(USER_BUCKET)
		followers := tx.Bucket(FOLLOWER_BUCKET)
		users.ForEach(func(id, v []byte) error {
			var u User
			err := json.Unmarshal(v, &u)
			if err != nil {
				return nil
			}

			// Check Followers bucket to see if this id exists
			u.IsFollower = len(followers.Get(id)) > 0
			buf, err := json.Marshal(u)
			if err != nil {
				return err
			}
			return users.Put([]byte(u.ID), buf)
		})
		return nil
	})

	twitchUsers.Purge()
	users.Purge()

	return err
}

func getMainChannel() string {
	if mainChannel != "" {
		return mainChannel
	}

	type twitchConfig struct {
		MainChannel string `json:"mainChannel"`
	}

	if c, ok := config.ModuleConfig["twitch"]; ok {
		var tc twitchConfig
		if err := json.Unmarshal(c, &tc); err == nil {
			mainChannel = tc.MainChannel
		}
	}

	return mainChannel
}

func getUserID() string {
	if userID != "" {
		return userID
	}

	if u, err := GetUserByName(getMainChannel()); err == nil {
		userID = u.ID
	}

	return userID
}
