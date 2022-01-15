package main

import (
	"net/url"
	"sync"
	"time"
)

type Crawler struct {
	Driver *Driver
	RWMute *sync.RWMutex
	WaitPS int
}

func NewCrawler(exec string, port int, works int) *Crawler {
	driver := NewDriver(exec, uint16(port))
	if driver.Run() != nil {
		return nil
	}
	rwmute := &sync.RWMutex{}
	return &Crawler{Driver: driver, RWMute: rwmute, WaitPS: works}
}

func (cl *Crawler) Destory() {
	cl.RWMute.Lock()
	cl.WaitPS = -cl.WaitPS
	cl.RWMute.Unlock()
	for {
		cl.RWMute.RLock()
		if cl.WaitPS == 0 {
			cl.Driver.Stop()
			cl.RWMute.RUnlock()
			return
		}
		cl.RWMute.RUnlock()
	}
}

func (cl *Crawler) Scratch(home string, class func(*Page) bool) *Site {
	if cl.RWMute.RLock(); cl.WaitPS <= 0 {
		cl.RWMute.RUnlock()
		return nil
	} else {
		cl.WaitPS -= 1
		cl.RWMute.RUnlock()
	}

	session, err := NewSession(cl.Driver)
	if err != nil {
		return nil
	}
	defer session.Free()
	base, _ := url.Parse(home)
	site := NewSite(base)
	for link := site.Detach(); link != nil; link = site.Detach() {
		session.Open(link.String())
		time.Sleep(time.Minute)
		html, _ := session.Html()
		text, _ := session.Text()
		if page := NewPage(link, html, text);class == nil || class(page) {
			site.Attach(page)
		}
	}
	return site
}
