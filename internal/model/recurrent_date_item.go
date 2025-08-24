package model

type ConfigItem struct {
	Name         string `yaml:"name"`
	EveryWhenDay int    `yaml:"everyWhenDay"`
}

type Config struct {
	Items []ConfigItem `toml:"items" yaml:"items"`
}
