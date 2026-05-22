package app

import (
	"github.com/lihaiya/aiops/pkg/admin"
)

func Admin() {
	var a App
	admin.New(a, "app", "应用")
}
