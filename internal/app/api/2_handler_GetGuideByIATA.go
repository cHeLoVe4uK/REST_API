package api

import (
	"net/http"
	"restapi/internal/app/models"
	"strings"

	"github.com/gin-gonic/gin"
)

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
