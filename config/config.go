package config

type Config struct {
	MysqlDsn    string
	NodeURL     string
	NetworkNode string
}

func DefaultConfig() *Config {
	return &Config{
		MysqlDsn:    "admin:_Admin123@(127.0.0.1:3306)/subspace?parseTime=true&loc=Local",
		NodeURL:     "ws://127.0.0.1:9944",
		NetworkNode: "polkadot",
	}
}
