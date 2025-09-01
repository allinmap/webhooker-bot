package bot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"telegram-webhooker/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	config *config.Config
}

func NewBot(cfg *config.Config) (*Bot, error) {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("failed to create bot: token missing")
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	chatIDsString := os.Getenv("CHAT_IDS")
	if chatIDsString == "" {
		return nil, fmt.Errorf("failed to create bot: chat IDs missing")
	}

	chatIDsStrings := strings.Split(chatIDsString, ",")

	// Convert string chat IDs to int64
	for _, chatIDStr := range chatIDsStrings {
		chatIDStr = strings.TrimSpace(chatIDStr)
		if chatIDStr == "" {
			continue
		}
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chat ID '%s': %w", chatIDStr, err)
		}
		cfg.Server.ChatIDs = append(cfg.Server.ChatIDs, chatID)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &Bot{
		api:    bot,
		config: cfg,
	}, nil
}

func (b *Bot) SendMessage(chatId int64, text string) error {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "Markdown"

	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (b *Bot) SendFormattedMessage(hostName, messageType string, data map[string]any) error {
	hostConfig := b.config.GetHostConfig(hostName)
	if hostConfig == nil {
		// Send error message to all configured chat IDs
		errorMsg := fmt.Sprintf("Unknown host: %s", hostName)
		for _, chatID := range b.config.Server.ChatIDs {
			if err := b.SendMessage(chatID, errorMsg); err != nil {
				log.Printf("Failed to send error message to chat %d: %v", chatID, err)
			}
		}
		return fmt.Errorf("unknown host: %s", hostName)
	}

	// Get template for message type, fallback to default
	template, exists := hostConfig.Templates[messageType]
	if !exists {
		template = hostConfig.Templates["default"]
		if template == "" {
			template = "📡 {{.host}}: {{.message}}"
		}
	}

	// Simple template replacement (in production, consider using text/template)
	message := b.replaceTemplateVars(template, data)

	// Send message to all configured chat IDs
	var lastError error
	for _, chatID := range b.config.Server.ChatIDs {
		if err := b.SendMessage(chatID, message); err != nil {
			log.Printf("Failed to send message to chat %d: %v", chatID, err)
			lastError = err
		}
	}

	return lastError
}

func (b *Bot) replaceTemplateVars(template string, data map[string]any) string {
	result := template

	// Simple replacement - in production use text/template package
	for key, value := range data {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		result = replaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result
}

func replaceAll(s, old, new string) string {
	// Simple string replacement
	for i := 0; i < len(s); i++ {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			s = s[:i] + new + s[i+len(old):]
			i += len(new) - 1
		}
	}
	return s
}
