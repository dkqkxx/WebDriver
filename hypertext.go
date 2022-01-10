package main

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Page struct {
	Base *url.URL
	Html string
	Text string
}

func NewPage(base *url.URL, html string, text string) *Page {
	if unq, err := strconv.Unquote(html); err == nil {
		html = unq
	}
	if unq, err := strconv.Unquote(text); err == nil {
		text = unq
	}
	return &Page{Base: base, Html: html, Text: text}
}

func (pg *Page) PickURL() []*url.URL {
	reg := regexp.MustCompile(`<a .*?href=['"]([^'"]*?)['"].*?>`)
	mats := reg.FindAllStringSubmatch(pg.Html, -1)
	grep := make(map[string]bool)
	res := make([]*url.URL, 0, len(grep))
	for i := range mats {
		url, err := url.Parse(mats[i][1])
		if err != nil {
			continue
		}
		url = pg.Base.ResolveReference(url)
		if url.Host == pg.Base.Host {
			url.Scheme = pg.Base.Scheme
		}
		url.Fragment = ""
		if !grep[url.String()] {
			grep[url.String()] = true
			res = append(res, url)
		}
	}
	return res
}

func (pg *Page) Content() string {
	reg := regexp.MustCompile(`[^[:alnum:]\p{Han}]+`)
	text := reg.ReplaceAllString(pg.Text, " ")
	cont := strings.TrimSpace(text)
	return cont
}

type Site struct {
	Home *url.URL
	Todo map[string]*url.URL
	Remo map[string]*url.URL
	Done map[string]string
}

func NewSite(home *url.URL) *Site {
	done := make(map[string]string)
	todo := make(map[string]*url.URL)
	remo := make(map[string]*url.URL)
	todo[home.String()] = home
	return &Site{Home: home, Done: done, Todo: todo, Remo: remo}
}

func (st *Site) Attach(page *Page) {
	st.Done[page.Base.String()] = page.Content()
	for _, url := range page.PickURL() {
		surl := url.String()
		if _, ok := st.Done[surl]; ok {
			continue
		} else if url.Host == st.Home.Host {
			st.Todo[surl] = url
		} else {
			st.Remo[surl] = url
		}
	}
}

func (st *Site) Detach() *url.URL {
	for surl, url := range st.Todo {
		delete(st.Todo, surl)
		return url
	}
	return nil
}
