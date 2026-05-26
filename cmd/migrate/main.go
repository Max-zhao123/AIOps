package main

import (
	"github.com/lihaiya/aiops/config"
	"github.com/lihaiya/aiops/initialize"
	aimigrate "github.com/lihaiya/aiops/pkg/aiops/migrate"
)

func main() {
	config.GVA_VP = initialize.Viper()
	config.GVA_DB = initialize.Gorm()
	if config.GVA_DB != nil {
		aimigrate.RegisterTables(config.GVA_DB)
		db, _ := config.GVA_DB.DB()
		defer db.Close()
	}
}
