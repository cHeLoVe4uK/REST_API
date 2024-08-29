package api

import (
	"net/http"
	"restapi/internal/app/models"
	"strings"

	"github.com/gin-gonic/gin"
)

// Получение информации о стране и гайда по IATA
func (a *API) GetCountryInfoByIATA(c *gin.Context) {
	// Каждый обработчик работает с атомарным счетчиком, по которому мы отследим что все они завершились
	a.Add(1)
	defer a.Done()
	a.logger.Info("Пользователь выполняет 'GET: GetCountryInfoByIATA /info/:iata'")

	// Считываем значение IATA, проверка что это не число и приведение к прописному виду в случае успеха
	iata := c.Params.ByName("iata")
	message, err := a.NotNumb(iata)
	if err == nil {
		c.JSON(http.StatusBadRequest, message)
		return
	}
	iataUP := strings.ToUpper(iata)

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Проверяем есть ли у нас в Кэше редиса элемент с таким IATA
	a.logger.Info("Поиск информация о стране в кэше редиса")
	val, err := a.redisClient.GetValue(iataUP)
	if err != nil {
		a.logger.Info("При получении информации о стране из кэша редиса произошла ошибка")
	} else {
		a.logger.Info("Информация о стране была найдена в кэше редиса")
		a.logger.Info("Поиск гайда в кэше")
		// Проверяем гайд в кэше
		g, ok := a.cache.Get(iataUP)
		if ok {
			guideSlice := make([]*models.Guide, 0, 1)
			Message := Message{
				Message:     "Запрос на получение информации о стране и получение гайда выполнен успешно",
				Guide:       append(guideSlice, g),
				CountryInfo: val,
			}
			c.JSON(http.StatusOK, Message)
			a.logger.Info("Гайд был найден в кэше при получении информации о стране")
			a.logger.Info("Запрос 'GET: GetCountryInfoByIATA /info/:iata' успешно выполнен")
			return
		}
		a.logger.Info("Гайд в кэше не был найден")

		// Ищем гайд в БД
		guide, ok, err := a.storage.Guide().GetGuideByIATA(iataUP)
		if err != nil {
			a.logger.Info("Проблемы c подключением к БД (таблица guides)")
			Message := Message{
				Message: "Извините, возникли проблемы с доступом к базе данных",
			}
			c.JSON(http.StatusInternalServerError, Message)
			return
		}
		// Если гайд не был найден
		if !ok {
			a.logger.Info("Гайд с таким IATA в базе данных не найден")
			Message := Message{
				Message:     "Гайда с таким IATA не существует",
				CountryInfo: val,
			}
			c.JSON(http.StatusOK, Message)
			return
		}

		// Если информация о стране была успешно получена и был найден гайд
		guideSlice := make([]*models.Guide, 0, 1)
		Message := Message{
			Message:     "Запрос на получение информации о стране и получение гайда выполнен успешно",
			Guide:       append(guideSlice, guide),
			CountryInfo: val,
		}
		c.JSON(http.StatusOK, Message)
		a.logger.Info("Запрос 'GET: GetCountryInfoByIATA /info/:iata' успешно выполнен")
		return
	}
	a.logger.Info("Информация о стране не найдена в кэше редиса")

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// Если информация о стране не была найдена в Кэше редиса
	// Проверяем что страна с таким IATA существует
	suggestions, err := a.GetCountryByIATA(iataUP)
	if err != nil {
		a.logger.Info("При получении информации о стране по ее IATA произошла ошибка")
		Message := Message{
			Message: "Извините, в данный момент эта функция недоступна. Попробуйте позже",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}
	// Если все прошло хорошо, но никакой информации о стране не было получено
	if len(suggestions) == 0 {
		a.logger.Info("При получение информации о стране пользователь указал несуществующий IATA")
		Message := Message{
			Message: "Вы указали несуществующий IATA",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Если информация о стране была успешно получена
	a.logger.Info("Информация о стране была успешно получена с сервиса DaData")
	var countryName string
	for _, s := range suggestions {
		countryName = s.Value
	}

	// Дбавляем элемент в Кэш редиса по его IATA
	err = a.redisClient.SetValue(iataUP, countryName)
	if err != nil {
		a.logger.Info("При попытке добавить элемент в кэш редиса произошла ошибка")
	}
	a.logger.Info("Информация о стране добавлена в кэш редиса")

	// Проверяем гайд в кэше
	a.logger.Info("Поиск гайда в кэше")
	g, ok := a.cache.Get(iataUP)
	if ok {
		guideSlice := make([]*models.Guide, 0, 1)
		Message := Message{
			Message:     "Запрос на получение информации о стране и получение гайда выполнен успешно",
			Guide:       append(guideSlice, g),
			CountryInfo: countryName,
		}
		c.JSON(http.StatusOK, Message)
		a.logger.Info("Гайд был найден в кэше при получении информации о стране")
		a.logger.Info("Запрос 'GET: GetCountryInfoByIATA /info/:iata' успешно выполнен")
		return
	}
	a.logger.Info("Гайд в кэше не был найден")

	// Ищем гайд в БД
	guide, ok, err := a.storage.Guide().GetGuideByIATA(iataUP)
	if err != nil {
		a.logger.Info("Проблемы c подключением к БД (таблица guides)")
		Message := Message{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}
	// Если гайд не был найден
	if !ok {
		a.logger.Info("Гайд с таким IATA в базе данных не найден")
		Message := Message{
			Message:     "Гайда с таким IATA не существует",
			CountryInfo: countryName,
		}
		c.JSON(http.StatusOK, Message)
		return
	}

	// Если информация о стране была успешно получена и был найден гайд
	guideSlice := make([]*models.Guide, 0, 1)
	Message := Message{
		Message:     "Запрос на получение информации о стране и получение гайда выполнен успешно",
		Guide:       append(guideSlice, guide),
		CountryInfo: countryName,
	}
	c.JSON(http.StatusOK, Message)

	a.logger.Info("Запрос 'GET: GetCountryInfoByIATA /info/:iata' успешно выполнен")
}
