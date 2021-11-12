package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	http.HandleFunc("/", defaultHandler)
	http.ListenAndServe(":8080", nil)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	s := requestPrint(t, r)
	fmt.Fprint(w, s)
	fmt.Println(s)
}

func requestPrint(t time.Time, r *http.Request) string {
	const log_time_format = "2006/01/02 15:04:05"

	s := fmt.Sprintf("Time:\n\t%s\n", t.Format(log_time_format))
	s += fmt.Sprintf("URL:\n\t%s\n", r.URL)
	s += fmt.Sprintf("Method:\n\t%s\n", r.Method)
	s += fmt.Sprintf("Protocol:\n\t%s\n", r.Proto)
	s += fmt.Sprintf("Header:\n\t%s\n", headerToString(r.Header, "\n\t"))
	s += fmt.Sprintf("Body:\n\t%s\n", bodyToString(r, "\n\t"))
	s += fmt.Sprintf("Parameters:\n\t%s\n", parsedParams(r, "\n\t"))

	return s
}

func headerToString(r http.Header, sep string) string {
	s := ""
	for k, v := range r {
		for _, e := range v {
			s += k + ":" + e + sep
		}
	}
	return strings.TrimRight(s, sep)
}

func parsedParams(r *http.Request, sep string) string {
	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}

	s := ""
	for k, v := range r.Form {
		for _, e := range v {
			s += k + ":" + e + sep
		}
	}
	return strings.TrimRight(s, sep)
}

func bodyToString(r *http.Request, sep string) string {
	var b []byte
	var err error

	if r.Body != nil {
		b, err = io.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
	}
	r.Body = io.NopCloser(bytes.NewBuffer(b))

	return string(b)
}
