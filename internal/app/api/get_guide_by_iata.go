package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Получение гайда по IATA
func (a *API) GetGuideByIATA(c *gin.Context) {
	// Каждый обработчик работает с атомарным счетчиком, по которому мы отследим что все они завершились
	a.Add(1)
	defer a.Done()
	a.logger.Info("User do 'GET: GetGuideByIATA /guide/:iata'")

	// Считываем значение IATA, проверка что это не число и приведение к прописному виду в случае успеха
	iata := c.Params.ByName("iata")
	message, err := a.NotNumb(iata)
	if err == nil {
		c.JSON(http.StatusBadRequest, message)
		return
	}
	iataUP := strings.ToUpper(iata)

	// Проверяем гайд в кэше
	a.logger.Info("Searching guide in Cache")
	g, ok := a.cache.Get(iataUP)
	if ok {
		Message := RequestMessage{
			Message: "Гайд успешно найден",
			Guide:   g,
		}
		c.JSON(http.StatusOK, Message)
		a.logger.Info("Guide was found in Cache")
		a.logger.Info("Request 'GET: GetGuideByIATA /guide/:iata' successfully done")
		return
	}
	a.logger.Info("Guide was not found in Cache")

	// Получаем гайд
	guide, ok, err := a.storage.Guide().GetGuideByIATA(iataUP)
	if err != nil {
		a.logger.Error("Trouble with connecting to DB (table guides):", " ", err)
		Message := RequestMessage{
			Message: "Извините, возникли проблемы с доступом к базе данных",
		}
		c.JSON(http.StatusInternalServerError, Message)
		return
	}
	// Если гайд не найден
	if !ok {
		a.logger.Info("Guide with this IATA not found in DB:", " ", iataUP)
		Message := RequestMessage{
			Message: "Гайда с таким IATA не существует",
		}
		c.JSON(http.StatusNotFound, Message)
		return
	}

	// Если все прошло успешно
	Message := RequestMessage{
		Message: "Гайд успешно найден",
		Guide:   guide,
	}
	c.JSON(http.StatusOK, Message)

	// Добавляем гайд в кэш
	a.logger.Info("Guide was add to Cache while recieve by IATA")
	a.cache.Set(iataUP, guide)

	a.logger.Info("Request 'GET: GetGuideByIATA /guide/:iata' successfully done")
}
