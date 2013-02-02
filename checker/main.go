// main.go
package checker

import (
	"appengine"
	"appengine/user"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

var (
	funcMap = map[string]checkFunc{"md5": checkMd5}
	tmpls   *template.Template
)

func init() {
	var (
		err error
	)
	tmpls = template.New("tmpls").Funcs(template.FuncMap{})
	if tmpls, err = tmpls.ParseFiles("templates/index.html"); err != nil {
		panic(err)
	}
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/check", check)
	http.HandleFunc("/data", data)
	http.HandleFunc("/del", del)
	http.HandleFunc("/save", save)
	http.HandleFunc("/add", add)
	http.HandleFunc("/", index)
}

type checkFunc func(appengine.Context, *Config) (bool, error)

func index(w http.ResponseWriter, r *http.Request) {
	var (
		usr *User
	)
	c := appengine.NewContext(r)
	u := user.Current(c)
	if usr = getUser(c, u.String()); usr == nil {
		usr = NewUser(u.String())
		handlePanic(w, usr.Save(c))
	}
	if !usr.Active {
		http.Error(w, fmt.Sprintf("User '%s' is not active!", u), 401)
		return
	}
	url, _ := user.LogoutURL(c, "/")
	dd := struct {
		UserName, LogoutUrl string
	}{u.String(), url}
	handlePanic(w, tmpls.ExecuteTemplate(w, "main", dd))
}

func data(w http.ResponseWriter, r *http.Request) {
	confData, err := Configs(appengine.NewContext(r))
	handlePanic(w, err)
	jsonData, err := json.Marshal(confData)
	handlePanic(w, err)
	fmt.Fprintf(w, "%s", jsonData)
}

func check(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		usrs []*User
	)
	c := appengine.NewContext(r)
	usrs, err = Users(c)
	handlePanic(w, err)
	for _, user := range usrs {
		if !user.Active {
			continue
		}
		confData, err := user.Configs(appengine.NewContext(r))
		handlePanic(w, err)
		for _, conf := range confData {
			c.Infof("Check: %s - %v", conf.Name, conf)
			ok, err := funcMap[conf.CheckFuncName](c, conf)
			if ok {
				err = conf.Notify(c)
			}
			handlePanic(w, err)
		}
	}
	fmt.Fprintf(w, "OK!")
}

func save(w http.ResponseWriter, r *http.Request) {
	cfg := NewConfigFromRequest(r)
	if err := cfg.Save(appengine.NewContext(r)); err != nil {
		fmt.Fprintf(w, "%v", err)
	} else {
		fmt.Fprintf(w, "")
	}
}

func add(w http.ResponseWriter, r *http.Request) {
	cfg := NewConfigFromRequest(r)
	if err := cfg.SaveAsNew(appengine.NewContext(r)); err != nil {
		fmt.Fprintf(w, "%v", err)
	} else {
		fmt.Fprintf(w, "")
	}
}

func del(w http.ResponseWriter, r *http.Request) {
	cfg := NewConfigFromRequest(r)
	if err := cfg.Delete(appengine.NewContext(r)); err != nil {
		fmt.Fprintf(w, "%v", err)
	} else {
		fmt.Fprintf(w, "")
	}
}

func handlePanic(w http.ResponseWriter, err error) {
	if err != nil {
		handleError(w, err)
		panic(err)
	}
}

func handleError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), 500)
}
