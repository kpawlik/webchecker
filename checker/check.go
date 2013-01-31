package checker

import (
	"appengine"
	"appengine/datastore"
	"appengine/mail"
	"appengine/urlfetch"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func checkMd5(c appengine.Context, key, url string, emails []string) (err error) {
	var (
		checkSum string
	)
	if pb, e := getPageBody(c, url); err != nil {
		err = e
		return
	} else {
		checkSum = calcMd5(pb)
	}
	result, e := getLastCheckResult(c, key)
	fmt.Println(result)
	if e != nil {
		err = e
		return
	}
	if result.Date == "" {
		k := datastore.NewIncompleteKey(c, key, nil)
		wd := &CheckResult{time.Now().Format(`02-01-2006T15:04:05`), checkSum}
		_, err = datastore.Put(c, k, wd)
		if err != nil {
			return
		}
	} else {
		wd, e := getLastCheckResult(c, key)
		if e != nil {
			err = e
			return
		}
		if wd.Md5 != checkSum {
			subject := fmt.Sprintf("Web page %s was changed!", key)
			message := fmt.Sprintf("Web page %s (%s) was changed!", key, url)
			k := datastore.NewIncompleteKey(c, key, nil)
			wd := &CheckResult{time.Now().Format(`02-01-2006T15:04:05`), checkSum}
			_, err = datastore.Put(c, k, wd)
			if err != nil {
				return
			}
			err = sendMail(c, subject, message, emails)
		}
	}
	return
}

func calcMd5(b []byte) string {
	md5 := md5.New()
	md5.Write(b)
	return fmt.Sprintf("%x", md5.Sum(nil))
}

func getPageBody(c appengine.Context, url string) (body []byte, err error) {
	var (
		resp *http.Response
	)
	client := urlfetch.Client(c)
	if resp, err = client.Get(url); err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	return
}

func sendMail(c appengine.Context, subject, message string, emails []string) (err error) {
	msg := &mail.Message{
		Sender:  "WEB CHECKER <kpawlik78@gmail.com>",
		To:      emails,
		Subject: subject,
		Body:    message,
	}
	err = mail.Send(c, msg)
	if err != nil {
		c.Infof("Send email error: %v", err)
	}
	return
}

func checkPage(c appengine.Context, conf *Config, usr *User) (err error) {
	var (
		checkSum string = ""
		cr       Keyed
	)
	db := NewDB(c)
	if pb, err := getPageBody(c, conf.Url); err != nil {
		return err
	} else {
		checkSum = calcMd5(pb)
	}
	_ = checkSum
	parentKey := conf.Key(c, usr.key(c, nil))
	if cr, err = db.LastRec(nCheckResult, "-Date", parentKey); err != nil {
		return err
	}
	if cr == nil {
		cr = &CheckResult{time.Now().Format(`02-01-2006T15:04:05`), checkSum}
		if _, err = db.Save(cr, parentKey); err != nil {
			return
		}
	}
	return
}
