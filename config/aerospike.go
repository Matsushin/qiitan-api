package config

// Aerospike 設定情報の構造体
type AerospikeConfig struct {
	Server        ServerConfig
	LikeRankingDB DBConfig
}

// ServerConfig Config Aerospike clientのConfig
type ServerConfig struct {
	Host string
	Port int
}

// DBConfig Aerospike Setのconfig
type DBConfig struct {
	Namespace string
	Set       string
	Key       string
}
