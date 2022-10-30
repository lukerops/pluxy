package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/gofiber/fiber/v2"
	"github.com/lukerops/pluxy/cmd/handlers"
	"github.com/lukerops/pluxy/pkg/downloader"
	"github.com/lukerops/pluxy/pkg/segmentmanager"
)

func init() {
	plutoTx := make(chan string)
	plutoRx := make(chan string)
	downloader.NewDownloader().Run(plutoRx, plutoTx)
}

func main() {
	app := fiber.New()

	app.Static("/segments", segmentmanager.SegmentManager.Dir)

	app.Get("/", handlers.Index)
	app.Get("/channels/:channel/master.m3u8", handlers.GetChannelPlaylist)

	serverShutdown := make(chan os.Signal, 1)
	signal.Notify(serverShutdown, os.Interrupt)

	go func() {
		<-serverShutdown
		fmt.Println("Gracefully shutting down...")
		app.Shutdown()
	}()

	app.Listen(":8080")

	// Finalizando o servidor
	if err := segmentmanager.SegmentManager.Stop(); err != nil {
		fmt.Println("Failed to stop SegmentManager; err:", err.Error())
	}
}
