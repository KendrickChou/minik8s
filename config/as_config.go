package config

type EtcdConfigType struct {
	EtcdAddr string
	EtcdPort int
}

var EtcdConfig = EtcdConfigType{EtcdPort: 2379, EtcdAddr: "127.0.0.1"}
