package config

type Kafka struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	ClientId string `yaml:"client_id"`
}