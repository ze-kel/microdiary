package mdbot

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ze-kel/microdiary/cmd/db"
	"github.com/ze-kel/microdiary/cmd/exporter"
	"gorm.io/gorm"
)

type MicroDiaryBot struct {
	db    *gorm.DB
	token string
}

func New(db *gorm.DB, token string) *MicroDiaryBot {
	return &MicroDiaryBot{
		db:    db,
		token: token,
	}
}

func (mdb *MicroDiaryBot) Start() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(mdb.handler),
	}

	b, err := bot.New(mdb.token, opts...)
	if err != nil {
		panic(err)
	}

	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, regexp.MustCompile("/export"), mdb.handleExport)
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, regexp.MustCompile("/clear"), mdb.handleDelete)

	log.Print("Starting bot")
	b.Start(ctx)
}

func (mdb *MicroDiaryBot) handler(ctx context.Context, bb *bot.Bot, update *models.Update) {
	log.Printf("Generic handler %d", update.ID)

	if update.EditedMessage != nil {
		mdb.db.Model(&db.Message{}).Where("message_id = ?", update.EditedMessage.ID).Update("message", update.EditedMessage.Text)
	} else if update.Message != nil {
		mdb.db.Create(&db.Message{MessageId: int64(update.Message.ID), ChatId: update.Message.Chat.ID, Date: int64(update.Message.Date), Message: update.Message.Text})
	}
}

func (mdb *MicroDiaryBot) handleExport(ctx context.Context, bb *bot.Bot, update *models.Update) {
	log.Printf("Export request %d", update.Message.Chat.ID)

	var messages []db.Message

	if err := mdb.db.Order("date asc").Find(&messages, "chat_id = ?", update.Message.Chat.ID).Error; err != nil {
		log.Fatalln(err)
	}
	log.Printf("Exporting chat %d messages: %d", update.Message.Chat.ID, len(messages))

	if len(messages) == 0 {
		bb.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Nothing to export yet",
		})
	}

	fullText := exporter.ComposeTextFromMessages(messages)
	filename := fmt.Sprintf("%d_%d.md", update.Message.ID, update.Message.Chat.ID)

	params := &bot.SendDocumentParams{
		ChatID:   update.Message.Chat.ID,
		Document: &models.InputFileUpload{Filename: filename, Data: bytes.NewReader([]byte(fullText))},
		Caption:  "Export",
	}

	bb.SendDocument(ctx, params)
}

func (mdb *MicroDiaryBot) handleDelete(ctx context.Context, bb *bot.Bot, update *models.Update) {
	mdb.handleExport(ctx, bb, update)
	mdb.db.Where("chat_id = ?", update.Message.Chat.ID).Unscoped().Delete(&db.Message{})
	bb.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Cleared all messages"),
	})
}
