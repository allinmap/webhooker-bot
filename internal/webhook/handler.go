package webhook

import (
	"net/http"

	"telegram-webhooker/internal/bot"
	"telegram-webhooker/internal/config"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	bot    *bot.Bot
	config *config.Config
}

type WebhookPayload struct {
	MessageType string         `json:"message_type"`
	Data        map[string]any `json:"data"`
	Message     string         `json:"message,omitempty"`
}

func NewHandler(bot *bot.Bot, config *config.Config) *Handler {
	return &Handler{
		bot:    bot,
		config: config,
	}
}

func (h *Handler) HandleWebhook(c *gin.Context) {
	hostName := c.Param("host")

	// Check if host is configured
	hostConfig := h.config.GetHostConfig(hostName)
	if hostConfig == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unknown or disabled host",
			"host":  hostName,
		})
		return
	}

	var payload WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON payload",
		})
		return
	}

	// If no specific data provided, use the message field
	if payload.Data == nil {
		payload.Data = make(map[string]any)
	}

	if payload.Message != "" {
		payload.Data["message"] = payload.Message
	}

	payload.Data["host"] = hostName

	err := h.bot.SendFormattedMessage(hostName, payload.MessageType, payload.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send message",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "Message sent successfully",
		"host":   hostName,
		"type":   payload.MessageType,
	})
}

func (h *Handler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"hosts":  h.getEnabledHosts(),
	})
}

func (h *Handler) getEnabledHosts() []string {
	var hosts []string
	for _, host := range h.config.Hosts {
		if host.Enabled {
			hosts = append(hosts, host.Name)
		}
	}
	return hosts
}
