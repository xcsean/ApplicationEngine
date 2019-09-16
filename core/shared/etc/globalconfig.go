package etc

import (
	"strings"
	"sync"

	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

type categoryEntry struct {
	category string
	kv       map[string]string
}

type globalConfig struct {
	lock    sync.RWMutex
	entries map[string]*categoryEntry
}

func newGlobalConfig() *globalConfig {
	return &globalConfig{
		entries: make(map[string]*categoryEntry),
	}
}

func (gc *globalConfig) getEntries() map[string]*categoryEntry {
	gc.lock.RLock()
	defer gc.lock.RUnlock()
	return gc.entries
}

func (gc *globalConfig) getValue(category, key string) (string, bool) {
	if category == "" || key == "" {
		return "", false
	}

	entries := gc.getEntries()
	entry, ok := entries[category]
	if !ok {
		return "", false
	}
	value, ok := entry.kv[key]
	return value, ok
}

func (gc *globalConfig) contains(category, key, pattern string) bool {
	if pattern == "" {
		return false
	}

	value, ok := gc.getValue(category, key)
	if !ok {
		return false
	}
	return strings.Contains(value, pattern)
}

func (gc *globalConfig) replace(newer map[string]*categoryEntry) {
	gc.lock.Lock()
	defer gc.lock.Unlock()
	gc.entries = newer
}

func (gc *globalConfig) dump() {
	log.Debug("global config:")
	entries := gc.getEntries()
	for key, entry := range entries {
		log.Debug("key=%s", key)
		for k, v := range entry.kv {
			log.Debug("    [%s]=%v", k, v)
		}
	}
}