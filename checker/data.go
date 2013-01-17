package checker

import (
	"appengine"
	"appengine/datastore"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	configTableName = "Cfg"
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
	key := datastore.NewKey(c, configTableName, cfg.Name, 0, nil)
	_, err = datastore.Put(c, key, cfg)
	return
}

func (cfg *Config) SaveAsNew(c appengine.Context) (err error) {
	if strings.ToLower(cfg.Name) == strings.ToLower(configTableName) {
		err = errors.New(fmt.Sprintf("This name '%s' is reserved", configTableName))
		return
	}
	key := datastore.NewKey(c, configTableName, cfg.Name, 0, nil)
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
	q := datastore.NewQuery(configTableName)
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
