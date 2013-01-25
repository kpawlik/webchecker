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
	http.HandleFunc("/check", check)
	http.HandleFunc("/data", data)
	http.HandleFunc("/del", del)
	http.HandleFunc("/save", save)
	http.HandleFunc("/add", add)
	http.HandleFunc("/", index)
}

type checkFunc func(appengine.Context, string, string, []string) error

func index(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		kusr Keyed
	)
	c := appengine.NewContext(r)
	db := NewDB(c)
	u := user.Current(c)
	usr := NewUser(u.String())
	if kusr, err = db.Get(usr, nil); kusr == nil {
		if err = db.Save(usr, nil); err != nil {
			panic(err)
		}
	} else {
		usr, _ = kusr.(*User)
		if !usr.Active {
			fmt.Fprintf(w, "User '%s' is not active!", u)
			return
		}
	}

	url, _ := user.LogoutURL(c, "/")
	dd := struct {
		UserName, LogoutUrl string
	}{u.String(), url}
	if err = tmpls.ExecuteTemplate(w, "main", dd); err != nil {
		panic(err)
	}
}

func data(w http.ResponseWriter, r *http.Request) {
	confData, err := Configs(appengine.NewContext(r))
	if err != nil {
		panic(err)
	}
	if jsonData, err := json.Marshal(confData); err != nil {
		panic(err)
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
		panic(err)
	}
	for _, user := range usrs {
		confData, err := ConfigsForUser(appengine.NewContext(r), user)
		if err != nil {
			panic(err)
		}
		for _, conf := range confData {
			c.Infof("Check: %s - %v", conf.Name, conf)
			err := funcMap[conf.CheckFuncName](c, conf.Name, conf.Url, conf.Emails)
			if err != nil {
				fmt.Fprintf(w, fmt.Sprintf("ERROR: %s", err))
				subject := fmt.Sprintf("Webchecker error %s", conf.Name)
				message := fmt.Sprintf("Error: %s", err)
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
