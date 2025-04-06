package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"goout/config"
	"goout/internal/service"
)

func main() {
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	group, _ := errgroup.WithContext(ctx)

	telegramService, err := service.NewTelegramService(cfg, group)
	if err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		log.Println("Received shutdown signal")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := telegramService.Stop(shutdownCtx); err != nil {
			log.Println("Error while stopping telegram service:", err)
		}

		cancel()
	}()

	if err := group.Wait(); err != nil {
		log.Println("Service exited with error:", err)
	} else {
		log.Println("Shutdown complete")
	}
}
