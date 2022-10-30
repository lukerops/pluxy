package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lukerops/pluxy/pkg/svc"
)

func GetChannelPlaylist(c *fiber.Ctx) error {
	channelID := c.Params("channel")

	playlist, err := svc.Channel.GetChannelPlaylist(channelID)
	if err != nil {
		return err
	}

	c.Set("Content-Type", "application/x-mpegurl")
	return c.SendString(playlist)
}
