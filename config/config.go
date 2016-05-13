package config

import (
	//"encoding/json"
	//"fmt"
	"github.com/pr8kerl/f5er/f5"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path/filepath"
)

type Rule struct {
	Local  string `yaml:"local"`
	Remote string `yaml:"remote"`
}

type Vip struct {
	Name        string `yaml:"name"`
	Destination string `yaml:"destination"`
	Pool        string `yaml:"pool"`
	Rules       []Rule `yaml:"rules"`
}

type Config struct {
	Vips []Vip
}

// Load parses the YAML input s into a Config.
func LoadConfig(s string) (*Config, error) {
	cfg := &Config{}
	err := yaml.Unmarshal([]byte(s), cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadFile parses the given YAML file into a Config.
func LoadConfigFile(filename string) (*Config, error) {
	absPath, _ := filepath.Abs(filename)
	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, err
	}
	cfg, err := LoadConfig(string(content))
	if err != nil {
		return nil, err
	}
	//resolveFilepaths(filepath.Dir(filename), cfg)
	return cfg, nil
}

func GetVirtualServer(vs string, bigip *f5.Device) (Vip, error) {
	err, virtual := bigip.ShowVirtual(vs)
	if err != nil {
		log.Fatalf("Error getting virtual: %s", err)
	}
	var vip Vip
	vip.Name = virtual.FullPath
	vip.Destination = virtual.Destination
	vip.Pool = virtual.Pool
	for _, r := range virtual.Rules {
		var rule Rule
		rule.Remote = r
		rule.Local = ""
		vip.Rules = append(vip.Rules, rule)
	}
	return vip, nil
}

func ShowConfig(bigip *f5.Device) (Config, error) {
	err, virtuals := bigip.ShowVirtuals()
	if err != nil {
		log.Fatalf("Error getting virtuals: %s", err)
	}
	var config Config
	for _, v := range virtuals.Items {
		var vip Vip
		vip.Name = v.FullPath
		vip.Destination = v.Destination
		vip.Pool = v.Pool
		for _, r := range v.Rules {
			var rule Rule
			rule.Remote = r
			rule.Local = ""
			vip.Rules = append(vip.Rules, rule)
		}
		config.Vips = append(config.Vips, vip)
	}
	return config, err
}
