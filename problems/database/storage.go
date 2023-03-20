package database

import (
	"github.com/rs/zerolog/log"
	"sync"
)

const (
	VersionKey = "version"
)

var (
	storage = map[string]string{}
	lock    = &sync.RWMutex{}
)

func init() {
	storage[VersionKey] = "Silly UDP 1.0"
}

func Query(key string) string {
	lock.RLock()
	defer lock.RUnlock()
	return storage[key]
}

func Set(key, value string) {
	if key == VersionKey {
		log.Warn().Msg("version attempted to be set")
		return
	}

	lock.Lock()
	defer lock.Unlock()
	storage[key] = value
	log.Info().Str("key", key).Str("value", value).Msg("updated map")
}
