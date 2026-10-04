package main

import (
	"context"
	"log"
	"time"

	backupapp "github.com/shawns-yao/shawn-blog/server/internal/app/backup"
	"github.com/shawns-yao/shawn-blog/server/internal/config"
)

func main() {
	if err := config.LoadRootEnv(); err != nil {
		log.Println("Root environment configuration could not be loaded; using process variables")
	}
	cfg := config.Load()
	timeout := cfg.Backup.CommandTimeout
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	log.Println("[restore] pending full-site restore detected")
	if err := backupapp.ExecutePendingRestore(ctx, cfg.Backup, cfg.Database.DSN, cfg.Redis); err != nil {
		log.Fatalf("[restore] failed: %v", err)
	}
	log.Println("[restore] full-site restore completed")
}
