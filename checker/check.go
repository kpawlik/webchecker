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

func compare(c appengine.Context, resbody, newbody []byte) (ok bool, err error) {
	return bytes.Equal(remWitheChars(resbody), remWitheChars(newbody)), nil
}

func calcMd5(s []byte) string {
	md5 := md5.New()
	md5.Write(s)
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
		Sender:      "KpaChecker <kpawlik78+wchecker@gmail.com>",
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
	b := new(bytes.Buffer)
	z := zip.NewWriter(b)
	if err = pack(z, oldResult); err != nil {
		return
	}
	if err = pack(z, newResult); err != nil {
		return
	}
	z.Close()
	res = b.Bytes()
	return
}

func pack(z *zip.Writer, result *CheckResult) (err error) {
	var (
		f         io.Writer
		fnPattern = "%s.txt"
	)
	if f, err = z.Create(fmt.Sprintf(fnPattern, result.Date)); err != nil {
		return
	}
	f.Write(result.Data)
	return
}

func remWitheChars(s []byte) []byte {
	empty := []byte("")
	tmpS := s
	for i := 0; i < len(whiteChars); i++ {
		tmpS = bytes.Replace(tmpS, whiteChars[i], empty, -1)
	}
	return tmpS
}
