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
	funcMap    = map[string]checkFunc{"cmp": compare}
	tmpls      *template.Template
	whiteChars = [...][]byte{[]byte("\t"), []byte(" "), []byte("\n")}
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
		handlePanic(w, c, usr.Save(c))
	}
	if !usr.Active {
		http.Error(w, fmt.Sprintf("User '%s' is not active!", u), 401)
		return
	}
	url, _ := user.LogoutURL(c, "/")
	dd := struct {
		UserName, LogoutUrl string
	}{u.String(), url}
	handlePanic(w, c, tmpls.ExecuteTemplate(w, "main", dd))
}

func data(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	confData, err := Configs(appengine.NewContext(r))
	handlePanic(w, c, err)
	jsonData, err := json.Marshal(confData)
	handlePanic(w, c, err)
	fmt.Fprintf(w, "%s", jsonData)
}

func check(w http.ResponseWriter, r *http.Request) {
	var (
		body   []byte
		err    error
		usrs   []*User
		result *CheckResult
		ok     bool
	)
	c := appengine.NewContext(r)
	usrs, err = Users(c)
	handlePanic(w, c, err)
	for _, user := range usrs {
		if !user.Active {
			continue
		}
		confData, err := user.Configs(appengine.NewContext(r))
		handlePanic(w, c, err)
		for _, conf := range confData {
			c.Infof("Checking: %s - %v", conf.Name, conf)
			if body, err = getPageBody(c, conf.Url); err != nil {
				handleError(w, c, err)
				continue
			}
			result, err = conf.LastResult(c, user)
			handlePanic(w, c, err)
			if result.Date == "" {
				result.Date = time.Now().Format(dataFormat)
				result.Data = body
				result.Parent = conf.Name
				handlePanic(w, c, result.SaveNew(c, conf, user))
				continue
			}
			fmt.Println("res ", result.Date)
			if ok, err = funcMap[conf.CheckFuncName](c, body, result.Data); ok {
				continue
			}
			handlePanic(w, c, err)
			wd := &CheckResult{time.Now().Format(dataFormat), body, conf.Name}
			handlePanic(w, c, wd.SaveNew(c, conf, user))
			handlePanic(w, c, conf.Notify(c, wd, result))
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

func handlePanic(w http.ResponseWriter, c appengine.Context, err error) {
	if err != nil {
		handleError(w, c, err)
		panic(err)
	}
}

func handleError(w http.ResponseWriter, c appengine.Context, err error) {
	c.Errorf("ERROR: %v", err)
	sendMail(c, "KpaChecker Error", fmt.Sprintf("%v", err), []string{"kpawlik78@gmail.com"}, nil)
	http.Error(w, err.Error(), 500)
}
