package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"runtime"
	"strings"
	"text/template"
	"time"
)

type flags struct {
	url        url.URL
	caption    string
	from_Gmail string
	to_mail    string
	password   string
}

var Dat *flags

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU()*2 + 1)

	log.Printf("Informer is running...")

	f, err := getFlags()
	if err != nil {
		log.Fatalf("flags parsing fail: %v", err)
	}

	http.HandleFunc("/message", messageHandler)

	err = http.ListenAndServe(getPort(f.url), nil)
	if err != nil {
		log.Fatalf("ListenAndServe: ", err)
	}
}

func messageHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		logf("err: r.ParseForm: %v", err)
		return
	}

	fmt.Printf("%v ms, %v\n",
		time.Now().UnixNano()/1000000,
		r.FormValue("message"))

	go sendGMail(r.FormValue("message"))
}

func getFlags() (flags, error) {

	// parse
	u := flag.String("u", "http://localhost:8082", "informer url")
	c := flag.String("c", "Informing", "caption")
	f := flag.String("f", "sender@gmail.com", "gmail sender")
	t := flag.String("t", "receiver@example.com", "email receiver")
	p := flag.String("p", "123456", "gmail password")

	flag.Parse()

	// url
	ur, err := url.Parse(*u)
	if err != nil {
		return flags{}, err
	}

	// caption
	ca := *c

	// from_Gmail
	fr := *f

	// to_mail
	to := *t

	//password
	pw := *p

	Dat = &flags{*ur, ca, fr, to, pw}

	return *Dat, nil
}

func sendGMail(message string) {

	auth := smtp.PlainAuth(
		"",
		Dat.from_Gmail,
		Dat.password,
		"smtp.gmail.com",
	)

	type SmtpTemplateData struct {
		From    string
		To      string
		Subject string
		Body    string
	}

	const emailTemplate = `From: {{.From}}
To: {{.To}}
Subject: {{.Subject}}

{{.Body}}
`

	var err error
	var doc bytes.Buffer

	context := &SmtpTemplateData{
		Dat.from_Gmail,
		Dat.to_mail,
		Dat.caption + " " + message,
		time.Now().Format("01/02 15:04:05"),
	}

	t := template.New("emailTemplate")
	t, err = t.Parse(emailTemplate)
	if err != nil {
		log.Printf("error trying to parse mail template")
		return
	}
	err = t.Execute(&doc, context)
	if err != nil {
		log.Printf("error trying to execute mail template")
		return
	}

	err = smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		Dat.from_Gmail,
		[]string{Dat.to_mail},
		doc.Bytes(),
	)
	if err != nil {
		log.Printf("smtp.SendMail err: " + err.Error())
		return
	}
}

func getPort(u url.URL) string {

	r := u.Host

	if n := strings.Index(r, ":"); n != -1 {
		r = r[n:]
	} else {
		r = ":8082"
	}

	return r
}

func logf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	log.Printf(s)
}
