package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
