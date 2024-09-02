package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Получение информации о стране и гайда по IATA
func (a *API) GetCountryInfoByIATA(c *gin.Context) {
	// Каждый обработчик работает с атомарным счетчиком, по которому мы отследим что все они завершились
	a.Add(1)
	defer a.Done()
	a.logger.Info("User do 'GET: GetCountryInfoByIATA /info/:iata'")

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
	a.logger.Info("Searching info about country in Redis")
	val, err := a.redisClient.GetValue(iataUP)
	if err != nil {
		a.logger.Error("An error occurred when recieving info about country from Redis:", " ", err)
	} else {
		a.logger.Info("Country info was found in Redis")
		a.logger.Info("Searching guide in Cache")
		// Проверяем гайд в кэше
		g, ok := a.cache.Get(iataUP)
		if ok {
			Message := RequestMessage{
				Message:     "Запрос на получение информации о стране и получение гайда выполнен успешно",
				Guide:       g,
				CountryInfo: val,
			}
			c.JSON(http.StatusOK, Message)
			a.logger.Info("Guide was found in Cache when calling GetCountryInfoByIATA")
			a.logger.Info("Request 'GET: GetCountryInfoByIATA /info/:iata' successfully done")
			return
		}
		a.logger.Info("Guide was not found in Cache")

		// Ищем гайд в БД
		guide, ok, err := a.storage.Guide().GetGuideByIATA(iataUP)
		if err != nil {
			a.logger.Error("Trouble with connecting to DB (table guides):", " ", err)
			Message := RequestMessage{
				Message: "Извините, возникли проблемы с доступом к базе данных",
			}
			c.JSON(http.StatusInternalServerError, Message)
			return
		}
		// Если гайд не был найден
		if !ok {
			a.logger.Info("Guide with this IATA not found in DB:", " ", iataUP)
			Message := RequestMessage{
				Message:     "Гайда с таким IATA не существует",
				CountryInfo: val,
			}
			c.JSON(http.StatusOK, Message)
			return
		}

		// Если информация о стране была успешно получена и был найден гайд
		Message := RequestMessage{
			Message:     "Запрос на получение информации о стране и получение гайда выполнен успешно",
			Guide:       guide,
			CountryInfo: val,
		}
		c.JSON(http.StatusOK, Message)
		a.logger.Info("Request 'GET: GetCountryInfoByIATA /info/:iata' successfully done")
		return
	}
	a.logger.Info("Country info not found in Redis")

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// Если информация о стране не была найдена в Кэше редиса
	// Проверяем что страна с таким IATA существует
	suggestions, err := a.GetCountryByIATA(iataUP)
	if err != nil {
		a.logger.Error("An error occurred when calling GetCountryByIATA:", " ", err)
		Message := RequestMessage{
			Message: "Извините, в данный момент эта функция недоступна. Попробуйте позже",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}
	// Если все прошло хорошо, но никакой информации о стране не было получено
	if len(suggestions) == 0 {
		a.logger.Info("User provided non existed IATA in PostGuideByIATA")
		Message := RequestMessage{
			Message: "Вы указали несуществующий IATA",
		}
		c.JSON(http.StatusBadRequest, Message)
		return
	}

	// Если информация о стране была успешно получена
	a.logger.Info("Country info successfully recieved from DaData")
	var countryName string
	for _, s := range suggestions {
		countryName = s.Value
	}

	// Дбавляем элемент в Кэш редиса по его IATA
	err = a.redisClient.SetValue(iataUP, countryName)
	if err != nil {
		a.logger.Error("An error occurred when adding Country info in Redis", " ", err)
	}
	a.logger.Info("Country info was added to Redis")

	// Проверяем гайд в кэше
	a.logger.Info("Searching guide in Cache")
	g, ok := a.cache.Get(iataUP)
	if ok {
		Message := RequestMessage{
			Message:     "Запрос на получение информации о стране и получение гайда выполнен успешно",
			Guide:       g,
			CountryInfo: countryName,
		}
		c.JSON(http.StatusOK, Message)
		a.logger.Info("Guide was found in Cache when calling GetCountryInfoByIATA")
		a.logger.Info("Request 'GET: GetCountryInfoByIATA /info/:iata' successfully done")
		return
	}
	a.logger.Info("Guide was not found in Cache")

	// Ищем гайд в БД
	guide, ok, err := a.storage.Guide().GetGuideByIATA(iataUP)
	if err != nil {
		a.logger.Error("Trouble with connecting to DB (table guides):", " ", err)
		Message := RequestMessage{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}
	// Если гайд не был найден
	if !ok {
		a.logger.Info("Guide with this IATA not found in DB:", " ", iataUP)
		Message := RequestMessage{
			Message:     "Гайда с таким IATA не существует",
			CountryInfo: countryName,
		}
		c.JSON(http.StatusOK, Message)
		return
	}

	// Если информация о стране была успешно получена и был найден гайд
	Message := RequestMessage{
		Message:     "Запрос на получение информации о стране и получение гайда выполнен успешно",
		Guide:       guide,
		CountryInfo: countryName,
	}
	c.JSON(http.StatusOK, Message)

	a.logger.Info("Request 'GET: GetCountryInfoByIATA /info/:iata' successfully done")
}
