package config

type ASConfigType struct {
	EtcdAddr string
	EtcdPort int

	HttpListenPort int
}

var ASConfig = ASConfigType{
	EtcdAddr: "127.0.0.1",
	EtcdPort: 2379,

	HttpListenPort: 8080,
}
