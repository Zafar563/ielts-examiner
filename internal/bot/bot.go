package bot

import (
	"context"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"ielts_bot/internal/config"
	"ielts_bot/internal/database"
	"ielts_bot/internal/llm"
	"ielts_bot/internal/models"
)

type Bot struct {
	api   *tgbotapi.BotAPI
	cfg   *config.Config
	repo  *database.Repository
	llm   *llm.LLMClient
	state *StateManager
}

func New(cfg *config.Config, repo *database.Repository, llmClient *llm.LLMClient) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}

	return &Bot{
		api:   api,
		cfg:   cfg,
		repo:  repo,
		llm:   llmClient,
		state: NewStateManager(),
	}, nil
}

// Start begins the long-polling loop for Telegram updates
func (b *Bot) Start() error {
	log.Printf("Authorized on account @%s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Ensure user is recorded in database
		go b.syncUser(update.Message.From)

		// Handle commands or text
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				b.handleStart(update.Message)
			case "help":
				b.handleHelp(update.Message)
			case "check":
				b.handleCheckPrompt(update.Message)
			case "history":
				b.handleHistory(update.Message)
			case "cancel":
				b.handleCancel(update.Message)
			default:
				b.sendPlainMessage(update.Message.Chat.ID, "Noma'lum buyruq. Mavjud buyruqlar: /start, /check, /history, /help, /cancel")
			}
			continue
		}

		// Handle button clicks or regular messages
		switch update.Message.Text {
		case "✍️ Insho tekshirish":
			b.handleCheckPrompt(update.Message)
		case "📊 Natijalarim":
			b.handleHistory(update.Message)
		case "ℹ️ Yordam & Mezonlar":
			b.handleHelp(update.Message)
		case "❌ Bekor qilish":
			b.handleCancel(update.Message)
		default:
			b.handleIncomingText(update.Message)
		}
	}

	return nil
}

func (b *Bot) syncUser(tgUser *tgbotapi.User) {
	if tgUser == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := &models.User{
		TelegramID: tgUser.ID,
		Username:   tgUser.UserName,
		FirstName:  tgUser.FirstName,
		LastName:   tgUser.LastName,
	}

	if err := b.repo.UpsertUser(ctx, user); err != nil {
		log.Printf("Failed to upsert user %d: %v", tgUser.ID, err)
	}
}
