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
	defaultPort            int
	defaultResponseStatus  int
	customResponseText     string
	outputFilePath         string
	outputFile             *os.File
	responseFilePath       string
	setContentTypeToJson   bool
	basicAuthUser          string
	basicAuthPassword      string
	documentServeDir       string
	customResponseHeaders  responseHeaders
	outputTemplateFilePath string
	routingSetting         *Setting
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage:
  %s [OPTIONS] [ROUTING FILES]...
Options`+"\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.IntVar(&defaultPort, "p", 8080, "listen port")
	flag.IntVar(&defaultResponseStatus, "s", 200, "response status")
	flag.StringVar(&customResponseText, "r", "", "response text")
	flag.StringVar(&outputFilePath, "o", "", "log output file path")
	flag.StringVar(&responseFilePath, "f", "", "response contents file path")
	flag.BoolVar(&setContentTypeToJson, "j", false, "set content type to json(application/json; charset=utf-8)")
	flag.StringVar(&basicAuthUser, "u", "", "basic authentication user name")
	flag.StringVar(&basicAuthPassword, "P", "", "basic authentication password")
	flag.StringVar(&documentServeDir, "d", "", "documents serve directory path")
	flag.Var(&customResponseHeaders, "H", "response header(ex: 'Content-Type: text/csv')")
	flag.StringVar(&outputTemplateFilePath, "g", "", "generate a routing file template to the specified path(ex: -g routing.json)")
	flag.Parse()

	if outputTemplateFilePath != "" {
		if err := genSettingTemplate(outputTemplateFilePath); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if setting, err := loadSettings(flag.Args()); err != nil {
		log.Fatal(err)
	} else {
		routingSetting = setting
	}

	if outputFilePath != "" {
		f, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		outputFile = f
		defer outputFile.Close()
	}

	if customResponseText == "" {
		customResponseText = readDefaultResponseText()
	}

	if documentServeDir != "" {
		http.HandleFunc("/", wrapBasicAuth(basicAuthUser, basicAuthPassword, func(w http.ResponseWriter, r *http.Request) {
			http.FileServer(http.Dir(documentServeDir)).ServeHTTP(w, r)
		}))
	} else {
		http.HandleFunc("/", wrapBasicAuth(basicAuthUser, basicAuthPassword, defaultHandler))
	}

	if err := setHandlersFromSetting(); err != nil {
		log.Fatal(err)
	}

	port := strconv.Itoa(defaultPort)

	log.Printf("listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	if setContentTypeToJson {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	for _, h := range customResponseHeaders {
		w.Header().Set(h.Field, h.Value)
	}
	w.WriteHeader(defaultResponseStatus)

	s := buildLogString(r)

	if customResponseText != "" {
		fmt.Fprint(w, customResponseText)
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

func wrapBasicAuth(user string, pass string, hf http.HandlerFunc) http.HandlerFunc {
	if basicAuthUser != "" {
		return func(w http.ResponseWriter, r *http.Request) {
			if u, p, ok := r.BasicAuth(); !ok || u != user || p != pass {
				logging(fmt.Sprintf("Basic auth not authorized.\n\taccepted user: %s\n\taccepted password: %s\n", u, p))

				w.Header().Add("WWW-Authenticate", `Basic realm="secret area"`)
				http.Error(w, "Not authorized", http.StatusUnauthorized)
			} else {
				hf(w, r)
			}
		}
	}
	return hf
}

func setHandlersFromSetting() error {
	if routingSetting == nil {
		return nil
	}

	for _, route := range routingSetting.Routes {
		if err := setHandlerFromRoute(&route); err != nil {
			return err
		}
	}
	return nil
}

func headerFromRoute(route *Route) (map[string]string, error) {
	m := make(map[string]string)

	if route.Headers != nil {
		for _, h := range route.Headers {
			f, v, err := parseResponseHeader(h)
			if err != nil {
				return nil, err
			}
			m[f] = v
		}
	}
	return m, nil
}

func setHandlerFromRoute(route *Route) error {
	headers, err := headerFromRoute(route)
	if err != nil {
		return err
	}

	response, err := responseTextFromRoute(route)
	if err != nil {
		return err
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		for f, v := range headers {
			w.Header().Set(f, v)
		}
		w.WriteHeader(route.Status)

		s := buildLogString(r)

		if response == "" {
			fmt.Fprint(w, s)
		} else {
			fmt.Fprint(w, response)
		}

		logging(s)
	}

	if route.ServeDirPath != "" {
		if !strings.HasSuffix(route.Path, "/") {
			route.Path += "/"
		}
		fn = func(w http.ResponseWriter, r *http.Request) {
			http.StripPrefix(route.Path, http.FileServer(http.Dir(route.ServeDirPath))).ServeHTTP(w, r)
		}
	}

	if route.BasicAuthUser != "" {
		http.HandleFunc(route.Path, wrapBasicAuth(route.BasicAuthUser, route.BasicAuthPassword, fn))
	} else {
		http.HandleFunc(route.Path, fn)
	}

	return nil
}

func responseTextFromRoute(route *Route) (string, error) {
	if route.BodyFilePath != "" {
		b, err := os.ReadFile(route.BodyFilePath)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	if route.BodyString != "" {
		return route.BodyString, nil
	}

	return "", nil
}
