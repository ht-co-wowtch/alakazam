package config

import (
	"bytes"
	"github.com/spf13/viper"
	"io/ioutil"
)

func Read(path string) (*viper.Viper, error) {
	viper.SetConfigType("yaml")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = viper.ReadConfig(bytes.NewBuffer(b)); err != nil {
		return nil, err
	}
	return viper.GetViper(), nil
}

func ReadTag(path, tag string) (*viper.Viper, error) {
	v, err := Read(path)
	if err != nil {
		return nil, err
	}
	if tag != "" {
		v = v.Sub(tag)
	}
	return v, nil
}
