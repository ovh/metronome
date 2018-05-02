package core

import (
	"fmt"
	"sync"

	"github.com/gobuffalo/packr"
)

var (
	boxOnce sync.Once
	boxes   map[string]packr.Box
)

// Assets return a packr box which contains all assets
func Assets(namespace string) (*packr.Box, error) {
	boxOnce.Do(func() {
		boxes = map[string]packr.Box{
			"auth": packr.NewBox("../controllers/auth/schema"),
			"task": packr.NewBox("../controllers/task/schema"),
			"user": packr.NewBox("../controllers/user/schema"),
		}
	})

	box, ok := boxes[namespace]
	if !ok {
		return nil, fmt.Errorf("No box found for namespace '%s'", namespace)
	}

	return &box, nil
}
