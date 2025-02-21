package tool

import (
	"strings"
	"sync"

	"github.com/showntop/flatmap"
)

type config struct {
	config *flatmap.Map
	once   sync.Once
}

var DefaultConfig config

func WithConfig(c map[string]any) error {
	var err error
	DefaultConfig.once.Do(func() {
		var mmm *flatmap.Map
		mmm, err = flatmap.Flatten(c, flatmap.DefaultTokenizer)
		DefaultConfig.config = mmm
	})
	return err
}

func (c *config) GetString(fields ...string) string {
	x, _ := c.config.Get(strings.Join(fields, flatmap.DefaultTokenizer.Separator())).(string)
	return x
}
