package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type Setting struct {
	Routes []Route
}

func (s *Setting) String() string {
	if s == nil {
		return ""
	}

	var v string
	for _, r := range s.Routes {
		v += fmt.Sprintf("%s\n", r.String())
	}
	return v
}

type Route struct {
	Path              string   `json:"path"`
	Headers           []string `json:"headers"`
	Status            int      `json:"status"`
	BodyString        string   `json:"bodyString"`
	BodyFilePath      string   `json:"bodyFilePath"`
	ServeDirPath      string   `json:"serveDirPath"`
	BasicAuthUser     string   `json:"basicAuthUser"`
	BasicAuthPassword string   `json:"basicAuthPassword"`
}

func (r *Route) String() string {
	var s string
	s += "path:\n\t" + r.Path + "\n"
	s += "headers:\n"
	for _, h := range r.Headers {
		s += "\t" + h + "\n"
	}
	s += "status:\n\t" + strconv.Itoa(r.Status) + "\n"
	s += "bodyString:\n\t" + r.BodyString + "\n"
	s += "bodyFilePath:\n\t" + r.BodyFilePath + "\n"
	s += "serveDirPath:\n\t" + r.ServeDirPath + "\n"
	s += "basicAuthUser:\n\t" + r.BasicAuthUser + "\n"
	s += "basicAuthPassword:\n\t" + r.BasicAuthPassword + "\n"

	return s
}

func loadSettings(filepath []string) (*Setting, error) {
	if len(filepath) == 0 {
		return nil, nil
	}

	var ret *Setting
	for _, p := range filepath {
		s, err := loadSetting(p)
		if err != nil {
			return nil, err
		}
		if ret == nil {
			ret = s
		} else {
			ret.Routes = append(ret.Routes, s.Routes...)
		}
	}
	return ret, nil
}

func loadSetting(filepath string) (*Setting, error) {
	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var setting Setting
	if err := json.Unmarshal(b, &setting); err != nil {
		return nil, err
	}

	return &setting, nil
}

func genSettingTemplate(filepath string) error {
	var s Setting

	r1 := &Route{
		Path:       "/v1/messages",
		Headers:    []string{"Content-Type: application/json"},
		Status:     200,
		BodyString: `{"messages":[{id:1, "message":"Hi!"},{id:2, "message":"bye!"}]}`,
	}
	s.Routes = append(s.Routes, *r1)

	r2 := &Route{
		Path:              "/v1/export",
		Headers:           []string{"Content-Type: text/csv"},
		Status:            200,
		BodyFilePath:      "sample.csv",
		BasicAuthUser:     "user",
		BasicAuthPassword: "secret",
	}
	s.Routes = append(s.Routes, *r2)

	r3 := &Route{
		Path:         "/assets/",
		Headers:      []string{"Content-Type: text/html"},
		Status:       200,
		ServeDirPath: "assets",
	}
	s.Routes = append(s.Routes, *r3)

	str, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath, []byte(str), 0666); err != nil {
		return err
	}

	return nil
}
