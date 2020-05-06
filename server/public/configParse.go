package public

import (
	"container/list"
	"fmt"
	"github.com/ini"
	"strconv"
	"strings"
)

type ServerHostConfig struct {
	Id   uint32
	Host string
}

type ConfigINI struct {
	fd            *ini.File
}

func NewConfigINI() *ConfigINI {
	return &ConfigINI{
		fd: &ini.File{},
	}
}

func (c *ConfigINI) Load(file string) error {
	conf, err := ini.Load(file)
	if err != nil {
		return err
	}
	c.fd = conf
	return nil
}

func (c* ConfigINI) GetString(section string, key string) (string, error) {
	s := c.fd.Section(section)
	if s == nil {
		return "", fmt.Errorf("can not get section:%s", section)
	}

	d := s.Key(key).String()
	if len(d) <= 0 {
		return "", fmt.Errorf("can not get key:%s", key)
	}

	return d, nil
}

func (c* ConfigINI) GetUInt32(section string, key string) (uint32, error) {
	s := c.fd.Section(section)
	if s == nil {
		return 0, fmt.Errorf("can not get section:%s", section)
	}

	d, err := s.Key(key).Int()
	if err != nil {
		return 0, fmt.Errorf("can not get key:%s:%s", key, err)
	}

	return uint32(d), nil
}

func (c* ConfigINI) GetBool(section string, key string) (bool, error) {
	s := c.fd.Section(section)
	if s == nil {
		return false, fmt.Errorf("can not get section:%s", section)
	}

	d, err := s.Key(key).Bool()
	if err != nil {
		return false, fmt.Errorf("can not get key:%s:%s", key, err)
	}

	return d, nil
}

func (c* ConfigINI) DealHosts(str string) (list.List, error) {
	str = PreString(str)
	lists := strings.Split(str, ";")
	var ret list.List
	for i := 0; i < len(lists); i++ {
		cfg := strings.Split(lists[i], "-")
		if len(cfg) == 2 {
			id, _ := strconv.Atoi(cfg[0])
			host := cfg[1]
			one := ServerHostConfig{uint32(id), host}
			ret.PushBack(one)
		}
	}
	return ret, nil
}

func PreString(str string) string {
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "\n", "", -1)
	str = strings.Replace(str, "\r", "", -1)
	return str
}
