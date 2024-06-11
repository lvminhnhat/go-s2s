package main

import (
	"splunkTcp/s2s"
)

var (
	SPLUNK_CONFIG = make(map[string][]s2s.S2S)
)

type S2S_Config struct {
	Endpoint     []string  `yaml:"endpoint_alias"`
	SplunkConfig []s2s.S2S `yaml:"s2s_config"`
}

func parseYAML(configs []S2S_Config) (map[string][]s2s.S2S, error) {
	endpoints := make(map[string][]s2s.S2S)
	for _, config := range configs {
		for _, alias := range config.Endpoint {
			for _, spl := range config.SplunkConfig {
				endpoints[alias] = append(endpoints[alias], spl)
			}
		}
	}
	return endpoints, nil
}
