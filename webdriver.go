package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Driver struct {
	Execute string
	Portnum uint16
	Service *os.Process
}

func NewDriver(path string, port uint16) *Driver {
	return &Driver{Execute: path, Portnum: port, Service: nil}
}

func (dr *Driver) Run() error {
	argv := []string{dr.Execute, fmt.Sprintf("--port=%d", dr.Portnum)}
	attr := os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}}
	proc, err := os.StartProcess(dr.Execute, argv, &attr)
	if err == nil {
		dr.Service = proc
	}
	return err
}

func (dr *Driver) Stop() error {
	return dr.Service.Kill()
}

type Session struct {
	ID string
	DR *Driver
}

func NewSession(dr *Driver) (*Session, error) {
	loc := fmt.Sprintf("http://127.0.0.1:%d/session", dr.Portnum)
	par := strings.NewReader(`{"capabilities":{"browserExecute":"chrome","alwaysMatch":{"goog:chromeOptions":{"args":["--headless"]}}}}`)
	val, err := Exec("POST", loc, par)
	if err != nil {
		return nil, err
	}
	obj := struct {
		Capabilities json.RawMessage `json:"capabilities"`
		SessionId    string          `json:"sessionId"`
	}{}
	if err := json.Unmarshal(val, &obj); err != nil {
		return nil, err
	}
	return &Session{ID: obj.SessionId, DR: dr}, nil
}

func (se *Session) Free() error {
	loc := fmt.Sprintf("http://127.0.0.1:%d/session/%s", se.DR.Portnum, se.ID)
	_, err := Exec("DELETE", loc, nil)
	return err
}

func (se *Session) Open(url string) error {
	loc := fmt.Sprintf("http://127.0.0.1:%d/session/%s/url", se.DR.Portnum, se.ID)
	par := strings.NewReader(fmt.Sprintf(`{"url":"%s"}`, url))
	_, err := Exec("POST", loc, par)
	return err
}

func (se *Session) Html() (string, error) {
	loc := fmt.Sprintf("http://127.0.0.1:%d/session/%s/source", se.DR.Portnum, se.ID)
	val, err := Exec("GET", loc, nil)
	return string(val), err
}

func (se *Session) Text() (string, error) {
	loc := fmt.Sprintf("http://127.0.0.1:%d/session/%s/element", se.DR.Portnum, se.ID)
	par := strings.NewReader(`{"using":"tag name","value":"html"}`)
	val, err := Exec("POST", loc, par)
	if err != nil {
		return "", err
	}
	ems := make(map[string]string)
	err = json.Unmarshal(val, &ems)
	for _, eid := range ems {
		loc := fmt.Sprintf("http://127.0.0.1:%d/session/%s/element/%s/text", se.DR.Portnum, se.ID, eid)
		val, err := Exec("GET", loc, nil)
		return string(val), err
	}
	return "", err
}

func Exec(method string, url string, par io.Reader) ([]byte, error) {
	req, _ := http.NewRequest(method, url, par)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	obj := struct {
		Value json.RawMessage `json:"Value"`
	}{}
	if err = json.Unmarshal(body, &obj); err != nil {
		return nil, err
	}
	return obj.Value, nil
}
