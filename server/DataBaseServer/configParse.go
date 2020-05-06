package main

import (
	"DrillServerGo/public"
)

type Configs struct {
	ini *public.ConfigINI

	serverid   uint32
	listenaddr string

	loglevel    uint32
	logfilename string

	sql_host     string
	sql_port     string
	sql_user     string
	sql_pwd      string
	sql_database string
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

	c.loglevel, err = c.ini.GetUInt32("log", "level")
	if err != nil {
		return err
	}

	c.logfilename, err = c.ini.GetString("log", "filename")
	if err != nil {
		return err
	}

	c.sql_host, err = c.ini.GetString("mysql", "host")
	if err != nil {
		return err
	}

	c.sql_port, err = c.ini.GetString("mysql", "port")
	if err != nil {
		return err
	}

	c.sql_user, err = c.ini.GetString("mysql", "user")
	if err != nil {
		return err
	}

	c.sql_pwd, err = c.ini.GetString("mysql", "pwd")
	if err != nil {
		return err
	}

	c.sql_database, err = c.ini.GetString("mysql", "database")
	if err != nil {
		return err
	}

	return nil
}


