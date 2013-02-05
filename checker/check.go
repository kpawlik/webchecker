package checker

import (
	"appengine"
	"appengine/mail"
	"appengine/urlfetch"
	"archive/zip"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func checkMd5(c appengine.Context, resbody, newbody []byte) (ok bool, err error) {
	return calcMd5(newbody) == calcMd5(resbody), nil
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

func sendMail(c appengine.Context, subject, message string, emails []string, attachments []mail.Attachment) (err error) {
	msg := &mail.Message{
		Sender:      "WEB CHECKER <kpawlik78@gmail.com>",
		To:          emails,
		Subject:     subject,
		Body:        message,
		Attachments: attachments,
	}
	err = mail.Send(c, msg)
	if err != nil {
		c.Infof("Send email error: %v", err)
	}
	return
}

func createArchive(newResult, oldResult *CheckResult) (res []byte, err error) {
	var (
		f1, f2 io.Writer
		fn     = "%s.txt"
	)
	b := new(bytes.Buffer)
	z := zip.NewWriter(b)
	if f1, err = z.Create(fmt.Sprintf(fn, newResult.Date)); err != nil {
		return
	}
	f1.Write(newResult.Data)
	if f2, err = z.Create(fmt.Sprintf(fn, oldResult.Date)); err != nil {
		return
	}
	f2.Write(oldResult.Data)
	z.Close()
	res = b.Bytes()
	return
}
