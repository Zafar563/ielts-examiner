package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"ielts_bot/internal/models"
)

var mainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("✍️ Insho tekshirish"),
		tgbotapi.NewKeyboardButton("📊 Natijalarim"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("ℹ️ Yordam & Mezonlar"),
		tgbotapi.NewKeyboardButton("❌ Bekor qilish"),
	),
)

var skipTopicKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("⏩ Mavzuni o'tkazib yuborish"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("❌ Bekor qilish"),
	),
)

func (b *Bot) handleStart(message *tgbotapi.Message) {
	text := fmt.Sprintf(
		"👋 Assalomu alaykum, *%s*!\n\n"+
			"Men — qat'iy va xolis **IELTS Writing Examiner** botiman.\n\n"+
			"Sizning insholaringizni 4 ta xalqaro mezon bo'yicha baholayman:\n"+
			"• **T/R** — Task Response (0–75)\n"+
			"• **C/C** — Coherence & Cohesion (0–75)\n"+
			"• **G/A** — Grammar & Accuracy (0–75)\n"+
			"• **L/R** — Lexical Resource (0–75)\n\n"+
			"📊 **Baholash darajalari:**\n"+
			"• 41–50 = B1\n"+
			"• 51–64 = B2\n"+
			"• 65–75 = C1\n\n"+
			"Boshlash uchun pastdagi **\"✍️ Insho tekshirish\"** tugmasini bosing yoki to'g'ridan-to'g'ri insho matnini yuboring.",
		escapeMarkdown(message.From.FirstName),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = mainKeyboard
	b.api.Send(msg)
}

func (b *Bot) handleHelp(message *tgbotapi.Message) {
	text := "📖 *IELTS Writing Baholash Tizimi haqida:*\n\n" +
		"Har bir mezon 0 dan 75 ballgacha baholanadi va o'rtacha ball chiqariladi:\n" +
		"• **65–75 (C1):** Kuchli nazorat, chuqur rivojlangan fikrlar, boy va aniq lug'at, xilma-xil grammatika.\n" +
		"• **51–64 (B2):** Samarali muloqot, lekin fikr rivojlantirish yoki lug'atda sezilarli chegaralar bor.\n" +
		"• **41–50 (B1):** Tushunarli, ammo cheklangan grammatika, oddiy lug'at va zaif izchillik.\n\n" +
		"📌 *Har bir tekshiruvda siz olasiz:*\n" +
		"1. Mezonlar bo'yicha aniq ball va batafsil izoh.\n" +
		"2. Xatolar va tabiiy bo'lmagan jumlalar jadvali (Grammar, Vocabulary, Collocation, Style...).\n" +
		"3. Nima uchun bundan yuqori ball ololmaganingiz sabablari.\n" +
		"4. 70+ ballga chiqish uchun 3–5 ta aniq maslahat.\n" +
		"5. Inshongizning asl g'oyasini saqlagan holda 70–75 ballik professional qayta yozilgan varianti (Rewrite)."

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = mainKeyboard
	b.api.Send(msg)
}

func (b *Bot) handleCheckPrompt(message *tgbotapi.Message) {
	session := b.state.GetSession(message.From.ID)
	session.Step = StepWaitingTopic
	session.Topic = ""

	text := "📌 **1-QADAM:** Iltimos, insho mavzusini (Task Prompt / Savol matnini) yuboring.\n\n" +
		"Agar faqat insho matnini tekshirmoqchi bo'lsangiz, **\"⏩ Mavzuni o'tkazib yuborish\"** tugmasini bosing."

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = skipTopicKeyboard
	b.api.Send(msg)
}

func (b *Bot) handleCancel(message *tgbotapi.Message) {
	b.state.ResetSession(message.From.ID)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Amaliyot bekor qilindi. Bosh menyudasiz.")
	msg.ReplyMarkup = mainKeyboard
	b.api.Send(msg)
}

func (b *Bot) handleHistory(message *tgbotapi.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, avgScore, err := b.repo.GetUserStats(ctx, message.From.ID)
	if err != nil {
		log.Printf("Error getting user stats: %v", err)
	}

	submissions, err := b.repo.GetUserSubmissions(ctx, message.From.ID, 5)
	if err != nil {
		log.Printf("Error getting user submissions: %v", err)
		b.sendPlainMessage(message.Chat.ID, "Tarixni yuklashda xatolik yuz berdi.")
		return
	}

	if len(submissions) == 0 {
		b.sendPlainMessage(message.Chat.ID, "Siz hali insho tekshirmagansiz. Insho yuborish uchun '✍️ Insho tekshirish' tugmasini bosing.")
		return
	}

	text := fmt.Sprintf("📊 *Sizning statistikangiz:*\n• Jami insholar: *%d ta*\n• O'rtacha ball: *%.1f / 75*\n\n📋 *So'nggi tekshiruvlar:*\n", count, avgScore)
	for i, s := range submissions {
		topicDisplay := s.Topic
		if topicDisplay == "" {
			topicDisplay = "Mavzusiz insho"
		}
		if len(topicDisplay) > 35 {
			topicDisplay = topicDisplay[:32] + "..."
		}
		text += fmt.Sprintf("%d. *%s* — *%d/75* (%s)\n   _TR:%d | CC:%d | GA:%d | LR:%d_ — %s\n\n",
			i+1,
			escapeMarkdown(topicDisplay),
			s.OverallScore,
			s.CEFRLevel,
			s.TRScore, s.CCScore, s.GAScore, s.LRScore,
			s.CreatedAt.Format("02.01.2006 15:04"),
		)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = mainKeyboard
	b.api.Send(msg)
}

func (b *Bot) handleIncomingText(message *tgbotapi.Message) {
	session := b.state.GetSession(message.From.ID)

	switch session.Step {
	case StepWaitingTopic:
		if message.Text == "⏩ Mavzuni o'tkazib yuborish" {
			session.Topic = ""
		} else {
			session.Topic = message.Text
		}
		session.Step = StepWaitingEssay

		prompt := "📝 **2-QADAM:** Endi insho matnini to'liq yuboring:"
		msg := tgbotapi.NewMessage(message.Chat.ID, prompt)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("❌ Bekor qilish")),
		)
		b.api.Send(msg)

	case StepWaitingEssay:
		essay := message.Text
		topic := session.Topic
		b.state.ResetSession(message.From.ID)
		go b.evaluateAndReply(message.Chat.ID, message.From, topic, essay)

	default:
		// If user pastes an essay directly (word count >= 30 or length >= 150)
		if len(strings.Fields(message.Text)) >= 25 {
			go b.evaluateAndReply(message.Chat.ID, message.From, "", message.Text)
		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Insho tekshirish uchun '✍️ Insho tekshirish' tugmasini bosing yoki insho matnini to'liq yuboring.")
			msg.ReplyMarkup = mainKeyboard
			b.api.Send(msg)
		}
	}
}

func (b *Bot) evaluateAndReply(chatID int64, from *tgbotapi.User, topic, essay string) {
	// Send "thinking / analyzing" status
	statusMsg := tgbotapi.NewMessage(chatID, "⏳ *Insho qabul qilindi!*\n\nQat'iy mezonlar bo'yicha tahlil qilinmoqda, iltimos kuting...")
	statusMsg.ParseMode = "Markdown"
	sentStatus, err := b.api.Send(statusMsg)
	if err != nil {
		log.Printf("Error sending status message: %v", err)
	}

	// Send chat action (typing)
	chatAction := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	b.api.Send(chatAction)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()

	// Call Examiner LLM
	result, err := b.llm.AssessEssay(ctx, topic, essay)
	if err != nil {
		log.Printf("Error evaluating essay: %v", err)
		b.deleteMessage(chatID, sentStatus.MessageID)
		b.sendPlainMessage(chatID, "Kechirasiz, inshoni tekshirishda xatolik yuz berdi. Iltimos, qaytadan urinib ko'ring yoki keyinroq sinab ko'ring.")
		return
	}

	// Save to database
	go func() {
		saveCtx, saveCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer saveCancel()

		sub := &models.Submission{
			UserID:       from.ID,
			Topic:        topic,
			EssayText:    essay,
			OverallScore: result.OverallScore,
			TRScore:      result.TRScore,
			CCScore:      result.CCScore,
			GAScore:      result.GAScore,
			LRScore:      result.LRScore,
			CEFRLevel:    result.CEFRLevel,
			Feedback:     result.Feedback,
		}
		if err := b.repo.SaveSubmission(saveCtx, sub); err != nil {
			log.Printf("Error saving submission to db: %v", err)
		}
	}()

	// Delete status message
	b.deleteMessage(chatID, sentStatus.MessageID)

	// Send response split into chunks if necessary
	b.sendSplittedResponse(chatID, result.Feedback)
}

func (b *Bot) sendSplittedResponse(chatID int64, text string) {
	// Telegram message character limit is 4096. Keep chunks under 3800 for safety.
	chunks := splitText(text, 3800)

	for _, chunk := range chunks {
		msg := tgbotapi.NewMessage(chatID, chunk)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = mainKeyboard

		_, err := b.api.Send(msg)
		if err != nil {
			// If markdown parsing fails due to special characters, send plain text
			log.Printf("Markdown send failed (%v), retrying plain text", err)
			plainMsg := tgbotapi.NewMessage(chatID, chunk)
			plainMsg.ReplyMarkup = mainKeyboard
			b.api.Send(plainMsg)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (b *Bot) sendPlainMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = mainKeyboard
	b.api.Send(msg)
}

func (b *Bot) deleteMessage(chatID int64, messageID int) {
	if messageID == 0 {
		return
	}
	del := tgbotapi.NewDeleteMessage(chatID, messageID)
	b.api.Send(del)
}

// splitText divides long text into readable chunks without splitting inside words
func splitText(text string, limit int) []string {
	var chunks []string
	runes := []rune(text)

	for len(runes) > limit {
		// Find last newline before limit
		splitIdx := -1
		for i := limit; i >= limit-500 && i > 0; i-- {
			if runes[i] == '\n' {
				splitIdx = i
				break
			}
		}

		if splitIdx == -1 {
			splitIdx = limit
		}

		chunks = append(chunks, string(runes[:splitIdx]))
		runes = runes[splitIdx:]
	}

	if len(runes) > 0 {
		chunks = append(chunks, string(runes))
	}

	return chunks
}

func escapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"`", "\\`",
	)
	return replacer.Replace(text)
}
