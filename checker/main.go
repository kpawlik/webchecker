// main.go
package checker

//TODO: as a parameter of dataMap add function to check if page was updated, 
//	check sum is not a good solution
//TODO: move dataMap to datastore or maybe config file
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
	/*confData = [...]*Config{
	&Config{Name: "siis", Url: "https://form.teleinfrastruktura.gov.pl/help/", Emails: []string{"k.pawlik@astec.net", "kpawlik78@gmail.com"}, CheckFuncName: "md5"},
	&Config{Name: "siis_xsd", Url: "https://form.teleinfrastruktura.gov.pl/static/help/siis2.2-8.xsd", Emails: []string{"k.pawlik@astec.net"}, CheckFuncName: "md5"},
	&Config{Name: "Are you fucking coding me", Url: "http://areyoufuckingcoding.me/", Emails: []string{"kpawlik78@gmail.com"}, CheckFuncName: "md5"}}
	*/
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
		err error
	)
	c := appengine.NewContext(r)
	u := user.Current(c)
	if usr := getUser(c, u.String()); usr == nil {
		usr = NewUser(u.String())
		if err = usr.Save(c); err != nil {
			panic(err)
		}
	} else {
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
