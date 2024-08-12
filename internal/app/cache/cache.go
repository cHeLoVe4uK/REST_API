package cache

import (
	"restapi/internal/app/models"
	"sync"
	"time"
)

// Структура, представляющая встроенный кэш для нашего приложения
type Cache struct {
	m               sync.RWMutex         // Поле для избежания race condition
	item            map[string]ItemCache // Элементы кэша (в формате ключ-значение)
	cleanupInterval time.Duration        // Интервал времени очистки (пусть всегда будет пол часа)
}

// Конструктор возвращающий экземпляр Кэша
func NewCache() *Cache {
	return &Cache{
		item:            make(map[string]ItemCache),
		cleanupInterval: 30 * time.Minute,
	}
}

// Структура, представляющая элементы кэша
type ItemCache struct {
	Value      *models.Guide // Значение элемента
	Created    time.Time     // Поле, показывающее время создания элемента
	Expiration int64         // Поле, отвечающее за длительность жизни элемента (пусть всегда будет час)
}

// Метод для добавления или изменения элемента в кэше
func (c *Cache) Set(key string, value *models.Guide) {
	c.m.Lock()
	defer c.m.Unlock()

	// Добавляем элемент в кэш по ключу
	c.item[key] = ItemCache{
		Value:      value,
		Created:    time.Now(),
		Expiration: time.Now().Add(1 * time.Hour).UnixNano(),
	}
}

// Метод для получения элемента из Кэша
func (c *Cache) Get(key string) (*models.Guide, bool) {
	c.m.RLock()
	defer c.m.RUnlock()
	// Проверяем есть ли элемент по ключу
	item, ok := c.item[key]
	if !ok {
		return nil, false
	}
	// Если время жизни элемента истекло
	if time.Now().UnixNano() > item.Expiration {
		return nil, false
	}

	return item.Value, true
}

// Метод для инвалидации кэша (удаления его элементов)
func (c *Cache) Delete() *Cache {
	c.m.Lock()
	defer c.m.Unlock()

	c = NewCache()

	return c
}

// Метод запускающий отдельную горутину для чистки кэша от устаревших элементов
func (c *Cache) StartGC() {
	go c.GC()
}

// ///////////////////////////////////////////////////////////////////////////////////////
// Метод, который будет удалять устаревшие элементы Кэша
func (c *Cache) GC() {
	for {
		// Ожидаем время установленное в Кэше как интервал времени очистки
		<-time.After(c.cleanupInterval)
		c.cleanupInterval = 30 * time.Minute

		if c.item == nil {
			return
		}

		if keys := c.expiredKeys(); len(keys) != 0 {
			c.clearExpiredItems(keys)
		}

	}
}

// Метод для поиска просроченных ключей
func (c *Cache) expiredKeys() []string {
	c.m.RLock()
	defer c.m.RUnlock()

	var keys []string
	for k, i := range c.item {
		if time.Now().UnixNano() > i.Expiration {
			keys = append(keys, k)
		}
	}
	return keys
}

// Метод для удаления просроченных элементов по ключам
func (c *Cache) clearExpiredItems(keys []string) {
	c.m.Lock()
	defer c.m.Unlock()

	for _, k := range keys {
		delete(c.item, k)
	}
}
