// main.go
package checker

import (
	"appengine"
	"appengine/user"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

const (
	dataFormat = `02-01-2006T15-04-05`
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

type checkFunc func(appengine.Context, []byte, []byte) (bool, error)

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
		body   []byte
		err    error
		usrs   []*User
		result *CheckResult
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
			body, err = getPageBody(c, conf.Url)
			handlePanic(w, err)
			result, err = conf.LastResult(c)
			handlePanic(w, err)
			if result.Date == "" {
				result.Date = time.Now().Format(dataFormat)
				result.Data = body
				result.Parent = conf.Name
				handlePanic(w, result.SaveNew(c, conf))
				continue
			}
			fmt.Println("res ", result.Date)
			if ok, err := funcMap[conf.CheckFuncName](c, body, result.Data); ok {
				continue
			} else {
				handlePanic(w, err)
			}
			wd := &CheckResult{time.Now().Format(dataFormat), body, conf.Name}
			err = wd.SaveNew(c, conf)
			handlePanic(w, err)
			fmt.Println("wd ", wd.Date)
			err = conf.Notify(c, wd, result)
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
