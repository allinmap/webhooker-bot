package main

import (
	"log"

	"telegram-webhooker/internal/bot"
	"telegram-webhooker/internal/config"
	"telegram-webhooker/internal/webhook"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	telegramBot, err := bot.NewBot(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	webhookHandler := webhook.NewHandler(telegramBot, cfg)

	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.POST("/webhook/:host", webhookHandler.HandleWebhook)
	router.GET("/health", webhookHandler.HandleHealth)

	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting server on %s", addr)
	log.Printf("Webhook endpoint: http://%s/webhook/{host}", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
