package checker

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	configTableName = "Cfg"
	userTabName     = "Usr"
)

var (
	reservedNames = []string{configTableName, userTabName}
)

type Config struct {
	Name, Url, CheckFuncName string
	Emails                   []string
}

func NewConfig(name, url, chkFunc string, emails []string) *Config {
	return &Config{Name: name,
		Url:           url,
		CheckFuncName: chkFunc,
		Emails:        emails}
}

func NewConfigFromRequest(r *http.Request) *Config {
	name := r.FormValue("Name")
	url := r.FormValue("Url")
	chkFunc := r.FormValue("CheckFuncName")
	emails := strings.Split(r.FormValue("Emails"), ",")
	return NewConfig(name, url, chkFunc, emails)

}

func (cfg *Config) Save(c appengine.Context) (err error) {
	u := getUserFromContext(c)
	key := datastore.NewKey(c, configTableName, cfg.Name, 0, u.Key(c))
	_, err = datastore.Put(c, key, cfg)
	return
}

func (cfg *Config) SaveAsNew(c appengine.Context) (err error) {
	if isReservedName(cfg.Name) {
		err = errors.New(fmt.Sprintf("This name '%s' is reserved", configTableName))
		return
	}
	u := getUserFromContext(c)
	key := datastore.NewKey(c, configTableName, cfg.Name, 0, u.Key(c))
	err = datastore.Get(c, key, cfg)
	if err == datastore.ErrNoSuchEntity {
		_, err = datastore.Put(c, key, cfg)
		return
	}
	err = errors.New(fmt.Sprintf("Record with key '%s'' already exists", cfg.Name))
	return
}

func (cfg *Config) Delete(c appengine.Context) (err error) {
	key := datastore.NewKey(c, configTableName, cfg.Name, 0, nil)
	err = datastore.Delete(c, key)
	return
}

func Configs(c appengine.Context) (cfgs []*Config, err error) {
	u := getUserFromContext(c)
	cfgs, err = ConfigsForUser(c, u)
	return
}

func ConfigsForUser(c appengine.Context, u *User) (cfgs []*Config, err error) {
	q := datastore.NewQuery(configTableName).Ancestor(u.Key(c))
	for t := q.Run(c); ; {
		result := &Config{}
		_, err = t.Next(result)
		if err == datastore.Done {
			err = nil
			break
		}
		if err != nil {
			return
		}
		cfgs = append(cfgs, result)
	}
	return
}

// CheckResult - record to store shortcut or change to check and information
// when last change was notoced.
type CheckResult struct {
	Date string
	Md5  string
}

func getLastCheckResult(c appengine.Context, key string) (result *CheckResult, err error) {
	result = &CheckResult{}
	q := datastore.NewQuery(key).
		Limit(1).
		Order("-Date")
	for t := q.Run(c); ; {
		_, err = t.Next(result)
		if err == datastore.Done {
			err = nil
			break
		}
		if err != nil {
			return
		}
	}
	return
}

type User struct {
	Name   string
	Active bool
}

func NewUser(name string) *User {
	return &User{name, true}
}

func (u User) Key(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, userTabName, u.Name, 0, nil)
}

func (u *User) Save(c appengine.Context) error {
	_, err := datastore.Put(c, u.Key(c), u)
	return err
}

func getUser(c appengine.Context, name string) (u *User) {
	u = &User{}
	key := datastore.NewKey(c, userTabName, name, 0, nil)
	if datastore.Get(c, key, u) == datastore.ErrNoSuchEntity {
		u = nil
	}
	return
}

func getUserFromContext(c appengine.Context) (u *User) {
	cusr := user.Current(c)
	return getUser(c, cusr.String())
}
func isReservedName(name string) bool {
	lname := strings.ToLower(name)
	for i := 0; i < len(reservedNames); i++ {
		if lname == strings.ToLower(reservedNames[i]) {
			return true
		}
	}
	return false
}

func Users(c appengine.Context) (usrs []*User, err error) {
	q := datastore.NewQuery(userTabName)
	for t := q.Run(c); ; {
		result := &User{}
		_, err = t.Next(result)
		if err == datastore.Done {
			err = nil
			break
		}
		if err != nil {
			return
		}
		usrs = append(usrs, result)
	}
	return
}
