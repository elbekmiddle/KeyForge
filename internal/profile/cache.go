package profile

import (
	"sync"
)

type Cache struct {
	mu       sync.RWMutex
	profiles map[string]Profile
	active   *Profile
}

func NewCache() *Cache {
	return &Cache{
		profiles: make(map[string]Profile),
	}
}

func (c *Cache) Set(profile Profile) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.profiles[profile.ID] = profile
}

func (c *Cache) SetActive(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	profile, ok := c.profiles[id]
	if !ok {
		return false
	}

	c.active = &profile

	return true
}

func (c *Cache) Active() (Profile, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.active == nil {
		return Profile{}, false
	}

	return *c.active, true
}

// All returns every cached profile, in no particular order — used by
// Match to pick the best fit for the currently focused application.
func (c *Cache) All() []Profile {
	c.mu.RLock()
	defer c.mu.RUnlock()

	all := make([]Profile, 0, len(c.profiles))
	for _, p := range c.profiles {
		all = append(all, p)
	}

	return all
}

func (c *Cache) ClearActive() {
	c.mu.Lock()
	c.active = nil
	c.mu.Unlock()
}
