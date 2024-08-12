package storage

import (
	"database/sql"
	"restapi/internal/app/config"

	_ "github.com/lib/pq"
)

// Инстанс хранилища для приложения
type Storage struct {
	// Поля неэкспортируемые (конфендициальная информация)
	config          *config.StorageConfig
	db              *sql.DB          // Сущность, представляющая собой мост между нашим приложением и БД
	guideRepository *GuideRepository // Модельный репозиторий, через который будет проводиться работа с БД
}

// Конструктор, возвращающий инстанс нашего сервера
func New(config *config.StorageConfig) *Storage {
	return &Storage{
		config: config,
	}
}

// Функция, открывающая соединение между нашим приложением и БД
func (storage *Storage) Open() error {
	db, err := sql.Open(storage.config.DriverName, storage.config.DataBaseURI)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	storage.db = db
	return nil
}

// Функция, закрывающая наше соединение с БД
func (storage *Storage) Close() {
	storage.db.Close()
}

// Функция, создающая публичный репозиорий для Guide
func (storage *Storage) Guide() *GuideRepository {
	if storage.guideRepository != nil {
		return storage.guideRepository
	}

	storage.guideRepository = &GuideRepository{
		storage: storage,
	}

	return storage.guideRepository
}
