package storage

import (
	"fmt"
	"log"
	"restapi/internal/app/models"
	"strings"
)

// Сущность модельного репозитория
type GuideRepository struct {
	storage *Storage // Хранит в себе БД, т.к. необходимо общаться с БД посредством репозитория (небольшое замыкание)
}

// Метод для получения гайда по стране с помощью ее IATA по стандарту ISO3166
func (guide *GuideRepository) GetGuideByIATA(iata string) (*models.Guide, bool, error) {
	var founded bool

	guides, err := guide.GetAllGuides()
	if err != nil {
		log.Println("При вызове метода получения гайда по IATA произошла ошибка:", err)
		return nil, founded, err
	}

	var guideFounded *models.Guide
	for _, g := range guides {
		if g.IATA == iata {
			guideFounded = g
			founded = true
			break
		}
	}

	return guideFounded, founded, nil
}

// Метод для добавления гайда по стране с помощью ее IATA по стандарту ISO3166
func (guide *GuideRepository) CreateGuide(g *models.Guide) (*models.Guide, error) {
	query := fmt.Sprintf("INSERT INTO %s (iata, title, author, content) VALUES ($1, $2, $3, $4) RETURNING iata", guide.storage.config.TableName)

	err := guide.storage.db.QueryRow(query, strings.ToUpper(g.IATA), g.Title, g.Author, g.Content).Scan(&g.IATA)
	if err != nil {
		log.Println("При вызове метода создания гайда произошла ошибка:", err)
		return nil, err
	}

	return g, nil
}

// Метод для обновления гайда по стране с помощью ее IATA по стандарту ISO3166
func (guide *GuideRepository) UpdateGuideByIATA(g *models.Guide) (*models.Guide, error) {
	query := fmt.Sprintf("UPDATE %s SET title=$1, author=$2, content=$3 WHERE iata=$4 RETURNING iata", guide.storage.config.TableName)

	err := guide.storage.db.QueryRow(query, g.Title, g.Author, g.Content, strings.ToUpper(g.IATA)).Scan(&g.IATA)
	if err != nil {
		log.Println("При вызове метода обновления гайда произошла ошибка:", err)
		return nil, err
	}

	return g, nil
}

// Метод для удаления гайда по стране с помощью ее IATA по стандарту ISO3166
func (guide *GuideRepository) DeleteGuideByIATA(iata string) (*models.Guide, error) {
	g, ok, err := guide.GetGuideByIATA(iata)
	if err != nil {
		return nil, err
	}
	if ok {
		query := fmt.Sprintf("DELETE FROM %s WHERE iata=$1", guide.storage.config.TableName)
		_, err := guide.storage.db.Exec(query, iata)
		if err != nil {
			log.Println("При вызове метода удаления гайда произошла ошибка:", err)
			return nil, err
		}
	}

	return g, nil
}

// Метод для получения всех гайдов в приложении
func (guide *GuideRepository) GetAllGuides() ([]*models.Guide, error) {
	query := fmt.Sprintf("SELECT * FROM %s", guide.storage.config.TableName)

	rows, err := guide.storage.db.Query(query)
	if err != nil {
		log.Println("При вызове метода получения всех гайдов произошла ошибка:", err)
		return nil, err
	}
	defer rows.Close()

	guides := make([]*models.Guide, 0)

	for rows.Next() {
		g := models.Guide{}
		err := rows.Scan(&g.IATA, &g.Title, &g.Author, &g.Content)
		if err != nil {
			log.Println(err)
			continue
		}
		guides = append(guides, &g)
	}

	return guides, nil
}
