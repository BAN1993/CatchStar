package main

import (
	"DrillServerGo/public"
	"container/list"
)

type Configs struct {
	ini *public.ConfigINI

	serverid    uint32
	listenaddr  string
	mode		string
	loglevel    uint32
	logfilename string

	dbs list.List // ServerHostConfig
	gs  list.List // ServerHostConfig
}

func (c* Configs) Load(path string) error {
	c.ini = public.NewConfigINI()
	err := c.ini.Load(path)
	if err != nil {
		return err
	}

	c.serverid, err = c.ini.GetUInt32("server", "serverid")
	if err != nil {
		return err
	}

	c.listenaddr, err = c.ini.GetString("server", "listen")
	if err != nil {
		return err
	}

	c.mode, err = c.ini.GetString("server","mode")
	if err != nil {
		c.mode = "tcp"
	}

	dbhosts, err := c.ini.GetString("hosts", "dbs")
	if err != nil {
		return err
	}
	c.dbs, err = c.ini.DealHosts(dbhosts)
	if err != nil {
		return err
	}

	gshost, err := c.ini.GetString("hosts", "gs")
	if err != nil {
		return err
	}
	c.gs, err = c.ini.DealHosts(gshost)
	if err != nil {
		return err
	}

	c.loglevel, err = c.ini.GetUInt32("log", "level")
	if err != nil {
		c.loglevel = 3
	}

	c.logfilename, err = c.ini.GetString("log", "filename")
	if err != nil {
		c.logfilename = "unknown"
	}

	return nil
}
