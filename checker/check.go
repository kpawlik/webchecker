package checker

import (
	"appengine"
	"appengine/mail"
	"appengine/urlfetch"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func checkMd5(c appengine.Context, conf *Config) (ok bool, err error) {
	var (
		checkSum string
		body     []byte
		result   *CheckResult
	)
	if body, err = getPageBody(c, conf.Url); err != nil {
		return
	}
	checkSum = calcMd5(body)
	if result, err = conf.LastResult(c); err != nil {
		return
	}
	if result.Date == "" {
		result.Date = time.Now().Format(`02-01-2006T15:04:05`)
		result.Result = checkSum
		result.Parent = conf.Name
		err = result.SaveNew(c, conf)
		ok = (err == nil)
		return
	}
	if !result.Equal(checkSum) {
		wd := &CheckResult{time.Now().Format(`02-01-2006T15:04:05`), checkSum, conf.Name}
		if err = wd.SaveNew(c, conf); err != nil {
			return
		}
		ok = true
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
