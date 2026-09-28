package config

func LoadTestConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8081,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "kori_test",
			User:     "test_user",
			Password: "test_password",
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}
}
