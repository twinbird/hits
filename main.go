package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
)

var (
	defaultPort           int
	defaultResponseStatus int
	defaultResponseText   string
	outputFilePath        string
	outputFile            *os.File
	responseFilePath      string
	setContentTypeToJson  bool
	basicAuthUser         string
	basicAuthPassword     string
)

func main() {
	flag.IntVar(&defaultPort, "p", 8080, "listen port")
	flag.IntVar(&defaultResponseStatus, "s", 200, "response status")
	flag.StringVar(&defaultResponseText, "r", "", "response text")
	flag.StringVar(&outputFilePath, "o", "", "log output file path")
	flag.StringVar(&responseFilePath, "f", "", "response contents file path")
	flag.BoolVar(&setContentTypeToJson, "j", false, "set content type to json(application/json; charset=utf-8)")
	flag.StringVar(&basicAuthUser, "u", "", "basic authentication user name")
	flag.StringVar(&basicAuthPassword, "P", "", "basic authentication password")
	flag.Parse()

	if outputFilePath != "" {
		f, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		outputFile = f
		defer outputFile.Close()
	}

	if defaultResponseText == "" {
		defaultResponseText = readDefaultResponseText()
	}

	http.HandleFunc("/", defaultHandler)

	port := strconv.Itoa(defaultPort)

	log.Printf("listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	if basicAuthUser != "" {
		if doBasicAuth(w, r) == false {
			return
		}
	}

	if setContentTypeToJson {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	w.WriteHeader(defaultResponseStatus)

	s := buildLogString(r)

	if defaultResponseText != "" {
		fmt.Fprint(w, defaultResponseText)
	} else {
		fmt.Fprint(w, s)
	}

	logging(s)
}

func logging(s string) {
	t := time.Now()
	const log_time_format = "2006/01/02 15:04:05"
	ts := fmt.Sprintf("Time:\n\t%s\n", t.Format(log_time_format))

	if outputFile != nil {
		outputFile.WriteString(ts + s)
	} else {
		fmt.Println(ts + s)
	}
}

func buildLogString(r *http.Request) string {
	var s string
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

func readDefaultResponseText() string {
	if responseFilePath != "" {
		fbytes, err := os.ReadFile(responseFilePath)
		if err != nil {
			log.Fatal(err)
		}
		return string(fbytes)
	}

	if !isatty.IsTerminal(os.Stdin.Fd()) {
		body, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		return string(body)
	}

	return ""
}

func doBasicAuth(w http.ResponseWriter, r *http.Request) bool {
	if user, pass, ok := r.BasicAuth(); !ok || user != basicAuthUser || pass != basicAuthPassword {
		logging(fmt.Sprintf("Basic auth not authorized.\n\taccepted user: %s\n\taccepted password: %s\n", user, pass))

		w.Header().Add("WWW-Authenticate", `Basic realm="secret area"`)
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return false
	}
	return true
}
