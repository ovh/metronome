package pg

import (
	"sync"

	"github.com/gobuffalo/packr"
)

var (
	boxOnce sync.Once
	box     packr.Box
)

// Assets return a packr box which contains all assets
func Assets() *packr.Box {
	boxOnce.Do(func() {
		box = packr.NewBox("./schema")
	})

	return &box
}
