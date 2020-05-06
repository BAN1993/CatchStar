package main

import (
	"DrillServerGo/public"
)

type Configs struct {
	ini *public.ConfigINI

	serverhost  string
	mode		string
	loglevel    uint32
	logfilename string
}

func (c* Configs) Load(path string) error {
	c.ini = public.NewConfigINI()
	err := c.ini.Load(path)
	if err != nil {
		return err
	}

	c.serverhost, err = c.ini.GetString("client", "serverhost")
	if err != nil {
		return err
	}

	c.mode, err = c.ini.GetString("client", "mode")
	if err != nil {
		c.mode = "tcp"
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
