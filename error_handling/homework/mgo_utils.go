package main

import (
	"strings"
	"sync"
	"time"

	"github.com/globalsign/mgo"
)

const (
	host   = ""
	source = "admin"
	user   = ""
	pass   = ""
)

var globalS *mgo.Session
var initOnce sync.Once

func init() {
	initOnce.Do(func() {
		addrs := strings.Split(host, ",")
		dialInfo := &mgo.DialInfo{
			Addrs:    addrs,
			Timeout:  50 * time.Second,
			Source:   source,
			Username: user,
			Password: pass,
		}
		s, err := mgo.DialWithInfo(dialInfo)
		if err != nil {
			panic("create session error " + err.Error())
		}
		globalS = s
	})
}

func connect(db, collection string) (*mgo.Session, *mgo.Collection) {
	s := globalS.Copy()
	c := s.DB(db).C(collection)
	return s, c
}

func FindOne(db, collection string, query, result interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()
	return c.Find(query).One(result)
}

func InsertOne(db, collection string, query interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()
	return c.Insert(query)
}
