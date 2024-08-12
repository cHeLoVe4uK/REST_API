package api

import (
	"context"
	"fmt"
	"net/http"
	"restapi/internal/app/models"
	"strconv"
	"strings"

	"github.com/ekomobile/dadata/v2/api/suggest"
	"github.com/gin-gonic/gin"
)

// Структура для предоставления информации пользователю
type Message struct {
	Message     string          `json:"message"`
	Guide       []*models.Guide `json:"guide,omitempty"`
	CountryInfo string          `json:"country_info,omitempty"`
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Метод (вспомогательный), проверяющий, что пользователь не вводил численные значение IATA (т.к. наше API такого не позволяет)
func (a *API) NotNumb(numb string) (Message, error) {
	_, err := strconv.Atoi(numb)
	if err == nil {
		a.logger.Info("Пользователь попытался выполнить запрос с численным значением IATA")
		return Message{Message: "Убедитесь в правильности введенного запроса. IATA не может быть числом"}, nil
	}
	return Message{}, err
}

// Функция для получения информации о стране через наш сторонний API (проверка что IATA введена верно)
func (a *API) GetCountryByIATA(iata string) ([]*suggest.CountrySuggestion, error) {
	suggestions, err := a.countryAPI.CountryByID(context.Background(), iata)
	if err != nil {
		a.logger.Info("Произошла ошибка при попытке получения информации о стране")
		return nil, err
	}

	return suggestions, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Получение всех гайдов (тут особо нет смысла считывать с кэша, т.к. нужны все элементы)
func (a *API) GetAllGuides(c *gin.Context) {
	// Каждый обработчик работает с атомарным счетчиком, по которому мы отследим что все они завершились
	a.Add(1)
	defer a.Done()
	a.logger.Info("Пользователь выполняет 'GET: GetAllGuides /guides'")

	// Получаем все гайды
	guides, err := a.storage.Guide().GetAllGuides()
	if err != nil {
		a.logger.Info("Проблемы c подключением к БД (таблица guides)")
		Message := Message{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}

	// Если все прошло успешно
	Message := Message{
		Message: "Гайды успешно найдены",
		Guide:   guides,
	}
	c.JSON(http.StatusOK, Message)

	a.logger.Info("Запрос 'GET: GetAllGuides /guides' успешно выполнен")
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Получение гайда по IATA
func (a *API) GetGuideByIATA(c *gin.Context) {
	// Каждый обработчик работает с атомарным счетчиком, по которому мы отследим что все они завершились
	a.Add(1)
	defer a.Done()
	a.logger.Info("Пользователь выполняет 'GET: GetGuideByIATA /guide/:iata'")

	// Считываем значение IATA, проверка что это не число и приведение к прописному виду в случае успеха
	iata := c.Params.ByName("iata")
	message, err := a.NotNumb(iata)
	if err == nil {
		c.JSON(http.StatusBadRequest, message)
		return
	}
	iataUP := strings.ToUpper(iata)

	// Проверяем гайд в кэше
	a.logger.Info("Поиск гайда в кэше")
	g, ok := a.cache.Get(iataUP)
	if ok {
		guideSlice := make([]*models.Guide, 0, 1)
		Message := Message{
			Message: "Гайд успешно найден",
			Guide:   append(guideSlice, g),
		}
		c.JSON(http.StatusOK, Message)
		a.logger.Info("Гайд был найден в кэше")
		a.logger.Info("Запрос 'GET: GetGuideByIATA /guide/:iata' успешно выполнен")
		return
	}
	a.logger.Info("Гайд в кэше не найден")

	// Получаем гайд
	guide, ok, err := a.storage.Guide().GetGuideByIATA(iataUP)
	if err != nil {
		a.logger.Info("Проблемы c подключением к БД (таблица guides)")
		Message := Message{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}
	// Если гайд не найден
	if !ok {
		a.logger.Info("Гайд с таким IATA в базе данных не найден")
		Message := Message{
			Message: "Гайда с таким IATA не существует",
		}
		c.JSON(http.StatusNotFound, Message)
		return
	}

	// Если все прошло успешно
	guideSlice := make([]*models.Guide, 0, 1)
	Message := Message{
		Message: "Гайд успешно найден",
		Guide:   append(guideSlice, guide),
	}
	c.JSON(http.StatusOK, Message)

	// Добавляем гайд в кэш
	a.logger.Info("При получении гайда по IATA он был добавлен в кэш")
	a.cache.Set(iataUP, guide)

	a.logger.Info("Запрос 'GET: GetGuideByIATA /guide/:iata' успешно выполнен")
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Добавление гайда по IATA
func (a *API) PostGuideByIATA(c *gin.Context) {
	// Каждый обработчик работает с атомарным счетчиком, по которому мы отследим что все они завершились
	a.Add(1)
	defer a.Done()
	a.logger.Info("Пользователь выполняет 'POST: PostGuideByIATA /guide/:iata'")

	// Считываем значение IATA, проверка что это не число и приведение к прописному виду в случае успеха
	iata := c.Params.ByName("iata")
	message, err := a.NotNumb(iata)
	if err == nil {
		c.JSON(http.StatusBadRequest, message)
		return
	}
	iataUP := strings.ToUpper(iata)

	// Подготавливаем переменную для гайда
	var guide models.Guide

	// Парсим тело запроса
	err = c.ShouldBindJSON(&guide)
	if err != nil {
		a.logger.Info("Пользователь предоставил неккоректный json")
		Message := Message{
			Message: "Предоставленные данные имеют неверный формат",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Проверяем что IATA в запросе и IATA в теле запроса не отличаются
	if iataUP != strings.ToUpper(guide.IATA) {
		a.logger.Info("Данные в запросе не совпадают с данными в json")
		Message := Message{
			Message: "IATA в запросе не соответствует IATA в теле запроса",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Проверяем что страна с таким IATA существует
	suggestions, err := a.GetCountryByIATA(iataUP)
	if err != nil {
		a.logger.Info("При создании гайда произошла ошибка с проверкой страны по ее IATA")
		Message := Message{
			Message: "Извините, в данный момент эта функция недоступна. Попробуйте позже",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}

	// Если все прошло хорошо, но никакой информации о стране не было получено
	if len(suggestions) == 0 {
		a.logger.Info("При создании гайда произошла ошибка, пользователь указал несуществующий IATA")
		Message := Message{
			Message: "Вы указали несуществующий IATA",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Если все прошло хорошо переходим к поиску гайда в БД
	_, ok, err := a.storage.Guide().GetGuideByIATA(iataUP)
	if err != nil {
		a.logger.Info("Проблемы c подключением к БД (таблица guides)")
		Message := Message{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}

	// Если гайд был найден
	if ok {
		a.logger.Info("Пользователь пытается создать уже существующий гайд")
		Message := Message{
			Message: "Вы пытаетесь создать гайд с уже существующим IATA",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Только после всего этого переходим непосредственно к созданию гайда
	guideFinal, err := a.storage.Guide().CreateGuide(&guide)
	if err != nil {
		a.logger.Info("Проблемы c подключением к БД (таблица guides)")
		Message := Message{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}

	// Если все прошло хорошо
	guideSlice := make([]*models.Guide, 0, 1)
	Message := Message{
		Message: fmt.Sprintf("Гайд с идентификатором {IATA: %v} успешно добавлен", guideFinal.IATA),
		Guide:   append(guideSlice, guideFinal),
	}
	c.JSON(http.StatusCreated, Message)

	a.logger.Info("Инвалидация кэша при добавлении гайда")
	a.cache = a.cache.Delete()

	a.logger.Info("Запрос 'POST: PostGuideByIATA /guide/:iata' успешно выполнен")
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Обновление гайда по IATA
func (a *API) PutGuideByIATA(c *gin.Context) {
	// Каждый обработчик работает с атомарным счетчиком, по которому мы отследим что все они завершились
	a.Add(1)
	defer a.Done()
	a.logger.Info("Пользователь выполняет 'PUT: PutGuideByIATA /guide/:iata'")

	// Считываем значение IATA, проверка что это не число и приведение к прописному виду в случае успеха
	iata := c.Params.ByName("iata")
	message, err := a.NotNumb(iata)
	if err == nil {
		c.JSON(http.StatusBadRequest, message)
		return
	}
	iataUP := strings.ToUpper(iata)

	// Подготавливаем переменную для гайда
	var guide models.Guide

	// Парсим тело запроса
	err = c.ShouldBindJSON(&guide)
	if err != nil {
		a.logger.Info("Пользователь предоставил неккоректный json")
		Message := Message{
			Message: "Предоставленные данные имеют неверный формат",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Проверяем что IATA в запросе и IATA в теле запроса не отличаются
	if iataUP != strings.ToUpper(guide.IATA) {
		a.logger.Info("Данные в запросе не совпадают с данными в json")
		Message := Message{
			Message: "IATA в запросе не соответствует IATA в теле запроса",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Если все прошло хорошо переходим к поиску гайда в БД
	_, ok, err := a.storage.Guide().GetGuideByIATA(iataUP)
	if err != nil {
		a.logger.Info("Проблемы c подключением к БД (таблица guides)")
		Message := Message{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}
	// Если гайд не найден
	if !ok {
		a.logger.Info("Пользователь пытается обновить несуществующий гайд")
		Message := Message{
			Message: "Вы пытаетесь обновить несуществующий гайд",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Только после всего этого переходим непосредственно к обновлению гайда
	guideFinal, err := a.storage.Guide().UpdateGuideByIATA(&guide)
	if err != nil {
		a.logger.Info("Проблемы c подключением к БД (таблица guides)")
		Message := Message{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}

	// Если все прошло хорошо
	guideSlice := make([]*models.Guide, 0, 1)
	Message := Message{
		Message: fmt.Sprintf("Гайд с идентификатором {IATA: %v} успешно обновлен", guideFinal.IATA),
		Guide:   append(guideSlice, guideFinal),
	}
	c.JSON(http.StatusOK, Message)

	a.logger.Info("Инвалидация кэша при обновлении гайда")
	a.cache = a.cache.Delete()

	a.logger.Info("Запрос 'PUT: PutGuideByIATA /guide/:iata' успешно выполнен")
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Удаление гайда по IATA
func (a *API) DeleteGuideByIATA(c *gin.Context) {
	// Каждый обработчик работает с атомарным счетчиком, по которому мы отследим что все они завершились
	a.Add(1)
	defer a.Done()
	a.logger.Info("Пользователь выполняет 'DELETE: DeleteGuideByIATA /guide/:iata'")

	// Считываем значение IATA, проверка что это не число и приведение к прописному виду в случае успеха
	iata := c.Params.ByName("iata")
	message, err := a.NotNumb(iata)
	if err == nil {
		c.JSON(http.StatusBadRequest, message)
		return
	}
	iataUP := strings.ToUpper(iata)

	// Ищем гайд в БД
	_, ok, err := a.storage.Guide().GetGuideByIATA(iataUP)
	if err != nil {
		a.logger.Info("Проблемы c подключением к БД (таблица guides)")
		Message := Message{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}
	// Если гайд не найден
	if !ok {
		a.logger.Info("Пользователь пытается удалить несуществующий гайд")
		Message := Message{
			Message: "Вы пытаетесь удалить несуществующий гайд",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Удаляем гайд
	guideFinal, err := a.storage.Guide().DeleteGuideByIATA(iataUP)
	if err != nil {
		a.logger.Info("Проблемы c подключением к БД (таблица guides)")
		Message := Message{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}

	// Если все прошло хорошо
	guideSlice := make([]*models.Guide, 0, 1)
	Message := Message{
		Message: fmt.Sprintf("Гайд с идентификатором {IATA: %v} успешно удален", guideFinal.IATA),
		Guide:   append(guideSlice, guideFinal),
	}
	c.JSON(http.StatusOK, Message)

	a.logger.Info("Инвалидация кэша при удалении гайда")
	a.cache = a.cache.Delete()

	a.logger.Info("Запрос 'DELETE: DeleteGuideByIATA /guide/:iata' успешно выполнен")
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
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
