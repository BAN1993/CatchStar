package main

import (
	"DrillServerGo/public"
	"container/list"
)

type Configs struct {
	ini *public.ConfigINI

	serverid      uint32
	listenaddr    string
	frameinterval uint32

	loglevel    uint32
	logfilename string

	dbs list.List // ServerHostConfig
}

func (c* Configs) Load(path string) error {
	c.ini = public.NewConfigINI()
	err := c.ini.Load(path)
	if err != nil {
		return err
	}

	id, err := c.ini.GetUInt32("server", "serverid")
	if err != nil {
		return err
	}
	c.serverid = id

	listen, err := c.ini.GetString("server", "listen")
	if err != nil {
		return err
	}
	c.listenaddr = listen

	frame, err := c.ini.GetUInt32("server", "frameinterval")
	if err != nil {
		return err
	}
	c.frameinterval = frame

	dbhosts, err := c.ini.GetString("hosts", "dbs")
	if err != nil {
		return err
	}
	c.dbs, err = c.ini.DealHosts(dbhosts)
	if err != nil {
		return err
	}

	level, err := c.ini.GetUInt32("log", "level")
	if err != nil {
		return err
	}
	c.loglevel = level

	filename, err := c.ini.GetString("log", "filename")
	if err != nil {
		return err
	}
	c.logfilename = filename

	return nil
}
