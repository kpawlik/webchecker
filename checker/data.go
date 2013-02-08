package checker

import (
	"appengine"
	"appengine/datastore"
	"appengine/mail"
	"appengine/user"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	configTableName      = "Cfg"
	userTabName          = "Usr"
	checkResultTableName = "Result"
	attachmentName       = "files.7z"
	messageSubject       = "Web page %s was changed!"
	messageText          = `Greetings user!

Web page %s (%s) was changed.
	
Compare attached files to find more details.
	

---
Best regards
kpachecker

(message send from: http://kpachecker.appspot.com/)`
)

/*************************************
Config
*************************************/
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
	if !includeKey(funcMap, cfg.CheckFuncName) {
		ks := keys(funcMap)
		err = errors.New(fmt.Sprintf("Incorect check method name: '%s'!\nAllowed names %v", cfg.CheckFuncName, ks))
		return
	}
	u := getUserFromContext(c)
	key := datastore.NewKey(c, configTableName, cfg.Name, 0, u.Key(c))
	_, err = datastore.Put(c, key, cfg)
	return
}

func (cfg *Config) SaveAsNew(c appengine.Context) (err error) {
	if !includeKey(funcMap, cfg.CheckFuncName) {
		ks := keys(funcMap)
		err = errors.New(fmt.Sprintf("Incorect check method name: '%s'!\nAllowed names %v", cfg.CheckFuncName, ks))
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
	var keys []*datastore.Key
	if keys, err = cfg.ResultsKeys(c); err != nil {
		return
	}
	if err = datastore.DeleteMulti(c, keys); err != nil {
		return
	}
	err = datastore.Delete(c, cfg.Key(c, nil))
	return
}

func (cfg *Config) Key(c appengine.Context, parent *datastore.Key) *datastore.Key {
	if parent == nil {
		u := user.Current(c)
		parent = datastore.NewKey(c, userTabName, u.String(), 0, nil)
	}
	return datastore.NewKey(c, configTableName, cfg.Name, 0, parent)
}

func (cfg *Config) LastResult(c appengine.Context, parent *User) (result *CheckResult, err error) {
	result = &CheckResult{}
	q := datastore.NewQuery(checkResultTableName).
		Limit(1).
		Order("-Date").Ancestor(cfg.Key(c, parent.Key(c)))
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

func (cfg *Config) ResultsKeys(c appengine.Context) (keys []*datastore.Key, err error) {
	var (
		key *datastore.Key
	)
	q := datastore.NewQuery(checkResultTableName).Ancestor(cfg.Key(c, nil)).KeysOnly()
	for t := q.Run(c); ; {
		key, err = t.Next(nil)
		if err == datastore.Done {
			err = nil
			break
		}
		if err != nil {
			return
		}
		keys = append(keys, key)
	}
	return
}

func (cfg *Config) Notify(c appengine.Context, newResult, oldResult *CheckResult) error {
	var (
		data []byte
		err  error
	)
	if data, err = createArchive(newResult, oldResult); err != nil {
		return err
	}
	attachs := []mail.Attachment{mail.Attachment{Name: attachmentName, Data: data}}
	subject := fmt.Sprintf(messageSubject, cfg.Name)
	message := fmt.Sprintf(messageText, cfg.Name, cfg.Url)
	return sendMail(c, subject, message, cfg.Emails, attachs)
}

func Configs(c appengine.Context) ([]*Config, error) {
	return getUserFromContext(c).Configs(c)
}

/*************************************
CheckResult
*************************************/
// CheckResult - record to store shortcut or change to check and information
// when last change was notoced.
type CheckResult struct {
	Date   string
	Data   []byte
	Parent string
}

func (cr *CheckResult) NewKey(c appengine.Context, parent *Config) *datastore.Key {
	return datastore.NewIncompleteKey(c, checkResultTableName, parent.Key(c, nil))
}

func (cr *CheckResult) Save(c appengine.Context, parent *Config) error {
	_, err := datastore.Put(c, cr.NewKey(c, parent), cr)
	return err
}

func (cr *CheckResult) SaveNew(c appengine.Context, cfg *Config, user *User) error {
	k := datastore.NewIncompleteKey(c, checkResultTableName, cfg.Key(c, user.Key(c)))
	_, err := datastore.Put(c, k, cr)
	return err
}

/*************************************
User
*************************************/
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

func (u *User) Configs(c appengine.Context) (cfgs []*Config, err error) {
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

/*************************************
Functions
*************************************/
func includeKey(m map[string]checkFunc, key string) bool {
	for k, _ := range m {
		if k == key {
			return true
		}
	}
	return false
}

func keys(m map[string]checkFunc) []string {
	var keys []string
	for k, _ := range m {
		keys = append(keys, k)
	}
	return keys
}
