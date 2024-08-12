package config

import (
	"log/slog"

	"github.com/BurntSushi/toml"
)

// Пакет с конфигами для нашего приложения

// Конфиг для сервера
type Config struct {
	BindAddr    string `toml:"bind_addr"`
	LoggerLevel int    `toml:"logger_level"`
	Storage     *StorageConfig
	OutsideAPI  *OutsideConfigAPI
	Redis       *RedisConfig
}

// Конструктор конфига для сервера
func NewConfigAPI(storage *StorageConfig, outside *OutsideConfigAPI, redis *RedisConfig) *Config {
	return &Config{
		Storage:    storage,
		OutsideAPI: outside,
		Redis:      redis,
	}
}

// Конфиг для БД
type StorageConfig struct {
	DataBaseURI string `toml:"database_uri"`
	DriverName  string `toml:"driver_name"`
	TableName   string `toml:"table_name"`
}

// Конструктор конфига для БД
func NewStorageConfig() *StorageConfig {
	return &StorageConfig{}
}

// Конфиг для сторонней API
type OutsideConfigAPI struct {
	ApiKeyValue    string `toml:"api_key"`
	SecretKeyValue string `toml:"secret_key"`
}

// Конструктор конфига для стороннего API
func NewOutsideConfigAPI() *OutsideConfigAPI {
	return &OutsideConfigAPI{}
}

type RedisConfig struct {
	Addr     string `json:"redis_addr"`
	Password string `json:"redis_password"`
}

func NewRedisConfig() *RedisConfig {
	return &RedisConfig{}
}

// Функция для настройки конфигов приложения
func AllApiConfig(configPathAPI *string) (*Config, error) {
	// конфиг для БД
	storConfig := NewStorageConfig() // конфиг для БД
	_, err := toml.DecodeFile(*configPathAPI, storConfig)
	if err != nil {
		slog.Info("Сервер не будет запущен, т.к. не найден файл с конфигурацией")
		return nil, err
	}

	// конфиг для стороннего API
	outsideAPIConfig := NewOutsideConfigAPI() // конфиг для стороннего API
	_, err = toml.DecodeFile(*configPathAPI, outsideAPIConfig)
	if err != nil {
		slog.Info("Сервер не будет запущен, т.к. не найден файл с конфигурацией")
		return nil, err
	}

	// Конфиг для Редиса
	redisConfig := NewRedisConfig() // конфиг для Редиса
	_, err = toml.DecodeFile(*configPathAPI, redisConfig)
	if err != nil {
		slog.Info("Redis не удалось настроить, т.к. не найден файл с конфигурацией", "ошибка:", err)
	}

	// конфиг для нашего API
	apiConfig := NewConfigAPI(storConfig, outsideAPIConfig, redisConfig) // конфиг для нашего API
	_, err = toml.DecodeFile(*configPathAPI, apiConfig)
	if err != nil {
		slog.Info("Сервер не будет запущен, т.к. не найден файл с конфигурацией")
		return nil, err
	}

	return apiConfig, nil
}
