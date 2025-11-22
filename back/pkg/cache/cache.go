package cache

// Cache определяет интерфейс для кеширования данных
type Cache[T any] interface {
	Get(key string) (*T, bool)
	Set(key string, entry *T)
	Delete(key string)
}
