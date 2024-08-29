package api

import (
	"context"
	"restapi/internal/app/models"
	"strconv"

	"github.com/ekomobile/dadata/v2/api/suggest"
)

// Структура для предоставления информации пользователю
type Message struct {
	Message     string          `json:"message"`
	Guide       []*models.Guide `json:"guide,omitempty"`
	CountryInfo string          `json:"country_info,omitempty"`
}

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
