// main.go
package checker

import (
	"appengine"
	"appengine/user"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
)

var (
	funcMap = map[string]checkFunc{"md5": checkMd5}
	tmpls   *template.Template
	p       = fmt.Println
	spf     = fmt.Sprintf
)

func init() {
	var (
		err error
	)
	tmpls = template.New("tmpls").Funcs(template.FuncMap{})
	if tmpls, err = tmpls.ParseFiles("templates/index.html"); err != nil {
		panic(err)
	}
	http.HandleFunc("/check", check)
	http.HandleFunc("/data", data)
	http.HandleFunc("/del", del)
	http.HandleFunc("/save", save)
	http.HandleFunc("/add", add)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/", index)

}

type checkFunc func(appengine.Context, string, string, []string) error

func index(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		kusr Keyed
		ok   bool
		usr  *User
	)
	c := appengine.NewContext(r)
	db := NewDB(c)
	u := user.Current(c)
	if kusr, err = db.SaveNew(NewUser(u.String()), nil); err != nil {
		panicError(c, w, err)
	}
	if usr, ok = kusr.(*User); !ok || (ok && !usr.Active) {
		serveError(c, w, errors.New(spf("User '%s' is not active!", u)))
		return
	}
	url, _ := user.LogoutURL(c, "/")
	//annonymous struct
	dd := struct {
		UserName, LogoutUrl string
	}{usr.Name, url}
	if err = tmpls.ExecuteTemplate(w, "main", dd); err != nil {
		panicError(c, w, err)
	}
}

func data(w http.ResponseWriter, r *http.Request) {
	var (
		confData []*Config
		err      error
	)
	c := appengine.NewContext(r)
	if confData, err = Configs(c); err != nil {
		panicError(c, w, err)
	}
	if jsonData, err := json.Marshal(confData); err != nil {
		panicError(c, w, err)
	} else {
		fmt.Fprintf(w, "%s", jsonData)
	}
}

func check(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		usrs []*User
	)
	c := appengine.NewContext(r)
	if usrs, err = Users(c); err != nil {
		panicError(c, w, err)
	}
	for _, user := range usrs {
		confData, err := ConfigsForUser(appengine.NewContext(r), user)
		if err != nil {
			panicError(c, w, err)
		}
		for _, conf := range confData {
			c.Infof("Check: %s - %v", conf.Name, conf)
			err := funcMap[conf.CheckFuncName](c, conf.Name, conf.Url, conf.Emails)
			if err != nil {
				fmt.Fprintf(w, spf("ERROR: %s", err))
				subject := spf("Webchecker error %s", conf.Name)
				message := spf("Error: %s", err)
				sendMail(c, subject, message, conf.Emails)
				return
			}
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

func serveError(c appengine.Context, w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Internal Server Error\n%v", err)
}

func panicError(c appengine.Context, w http.ResponseWriter, err error) {
	serveError(c, w, err)
	panic(err)
}
