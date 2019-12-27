package cache

import (
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"strings"
	"sync"
)

type UserCache struct {
	nickByID      map[int]string
	nickByIDMutex sync.RWMutex

	idByNick      map[string]int
	idByNickMutex sync.RWMutex
}

func NewUserCache() UserCache {
	return UserCache{
		nickByID: make(map[int]string),
		idByNick: make(map[string]int),
	}
}

func (u *UserCache) GetIDByNick(nick string) (int, error) {
	u.idByNickMutex.RLock()
	id, ok := u.idByNick[strings.ToLower(nick)]
	u.idByNickMutex.RUnlock()
	if !ok {
		return 0, consts.ErrNotFound
	}
	return id, nil
}

func (u *UserCache) GetNickByID(id int) (string, error) {
	u.nickByIDMutex.RLock()
	nick, ok := u.nickByID[id]
	u.nickByIDMutex.RUnlock()
	if !ok {
		return "", consts.ErrNotFound
	}
	return nick, nil
}

func (u *UserCache) GetNickCaseInsensitive(nick string) (string, error) {
	u.idByNickMutex.RLock()
	id, ok := u.idByNick[strings.ToLower(nick)]
	u.idByNickMutex.RUnlock()
	if !ok {
		return "", consts.ErrNotFound
	}
	return u.GetNickByID(id)
}

func (u *UserCache) Add(id int, nick string) {
	u.addNick(nick, id)
	u.addID(id, nick)
}

func (u *UserCache) addNick(nick string, id int) {
	u.idByNickMutex.Lock()
	u.idByNick[strings.ToLower(nick)] = id
	u.idByNickMutex.Unlock()
}

func (u *UserCache) addID(id int, nick string) {
	u.nickByIDMutex.Lock()
	u.nickByID[id] = nick
	u.nickByIDMutex.Unlock()
}
