package config

// MySQLConfig MySQLの設定を管理する構造体
type MySQLConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
}
