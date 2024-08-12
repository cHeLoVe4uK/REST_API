package api

import (
	"log/slog"
	"net/http"
	"restapi/internal/app/cache"
	"restapi/internal/app/config"
	red "restapi/internal/app/redis"
	"restapi/storage"

	"github.com/ekomobile/dadata/v2/api/suggest"
	"github.com/gin-gonic/gin"
)

// Инстанс нашего сервера
type API struct {
	// Поля неэкспортируемые (конфендициальная информация)
	srv         *http.Server     // можно сказать это поле проводник между нашим апи и встроенным сервером в пакете http, сделано это, чтобы реализовать gracefull shutdown
	config      *config.Config   // конфиг, который будет использоваться для настройки сервера
	logger      *slog.Logger     // логер который будет использоваться в процессе работы сервера
	router      *gin.Engine      // роутер который будет использоваться в процессе работы сервера (в нашем случае используем фреймворк gin)
	storage     *storage.Storage // БД, которая будет использоваться в процессе работы сервера
	countryAPI  *suggest.Api     // сторонняя API, которая будет использоваться для получения данных о странах
	cache       *cache.Cache     // Кэш нашего приложения
	redisClient *red.RedClient   // Кэш redis для нашего приложения
}

// Конструктор, возвращающий инстанс нашего сервера
func New(config *config.Config) *API {
	return &API{
		config: config,
	}
}

// Метод для получения нашего проводника
func (api *API) GetSrv() *http.Server {
	return api.srv
}

// Метод, стартующий наш сервер
func (api *API) Start() error {
	// Настройка поля логгер
	err := api.configureLoggerField()
	if err != nil {
		return err
	}
	api.logger.Info("Логгер сервера успешно настроен")

	// Настройка поля роутер
	api.configRouterField()
	api.logger.Info("Маршрутизатор готов к работе")

	// Настройка поля с хранилищем
	err = api.configureStorageField()
	if err != nil {
		return err
	}
	api.logger.Info("Соединение с базой данных успешно установлено")

	// Настройка поля стороннего API
	api.configureCountryAPIField()
	api.logger.Info("API для получения информации о странах успешно настроен")

	// Настройка поля srv
	srv := api.configureSrvField()
	api.logger.Info("Проводник для реализации плавного завершения сервера успешно настроен")

	// Настройка Кэша для приложения
	api.configureCacheField()
	api.logger.Info("Кэш для приложения успешно создан")

	// Настройка Редис кэша для приложения
	api.configureRedisClientField()
	api.logger.Info("Редис кэш для приложения успешно создан")

	// Запуск очистки кэша по интервалам
	go api.cache.StartGC()
	api.logger.Info("Горутина для контроля устаревших элементов запущена")

	// Сигнал о том, что настройка прошла успешно и запуск нашего сервера
	api.logger.Info("Стартуем", "Порт", api.config.BindAddr)

	return srv.ListenAndServe()
}
