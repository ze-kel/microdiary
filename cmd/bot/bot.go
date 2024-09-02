package mdbot

import (
	"context"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"regexp"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ze-kel/microdiary/cmd/db"
	"github.com/ze-kel/microdiary/cmd/exporter"
	storage "github.com/ze-kel/microdiary/cmd/filestorage"
	"gorm.io/gorm"
)

type MicroDiaryBot struct {
	db      *gorm.DB
	storage storage.GenericStorage

	token string
}

func New(db *gorm.DB, token string, str storage.GenericStorage) *MicroDiaryBot {
	return &MicroDiaryBot{
		db:      db,
		token:   token,
		storage: str,
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

func logAndReportError(ctx context.Context, bb *bot.Bot, update *models.Update, err error) {
	log.Printf(err.Error())
	bb.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Error: %s", err.Error()),
	})
}

func (mdb *MicroDiaryBot) saveFile(ctx context.Context, bb *bot.Bot, update *models.Update, fileId string, mimeType string) {
	file, err := bb.GetFile(ctx, &bot.GetFileParams{
		FileID: fileId,
	})

	if err != nil {
		logAndReportError(ctx, bb, update, err)
	}

	var format string

	if mimeType == "image/jpeg" {
		format = ".jpg"
	} else if mimeType == "audio/ogg" {
		format = ".ogg"

	} else {
		extensions, _ := mime.ExtensionsByType(mimeType)
		format = extensions[0]
	}

	fileName := fmt.Sprintf("%s%s", file.FileUniqueID, format)
	url := bb.FileDownloadLink(file)

	resp, err := http.Get(url)
	if err != nil {
		logAndReportError(ctx, bb, update, err)
		return
	}
	defer resp.Body.Close()

	err = mdb.storage.SaveFile(ctx, resp.Body, fileName, mimeType)

	if err != nil {
		logAndReportError(ctx, bb, update, err)
	}

	mdb.db.Create(&db.FilesForMessages{MessageId: int64(update.Message.ID), ChatId: update.Message.Chat.ID, Filename: fileName, Format: format, Date: int64(update.Message.Date)})
}

func (mdb *MicroDiaryBot) handler(ctx context.Context, bb *bot.Bot, update *models.Update) {
	log.Printf("Generic handler %d", update.ID)

	if update.EditedMessage != nil {
		mdb.db.Model(&db.Message{}).Where("message_id = ?", update.EditedMessage.ID).Update("message", update.EditedMessage.Text)
	} else if update.Message != nil {
		mdb.db.Create(&db.Message{MessageId: int64(update.Message.ID), ChatId: update.Message.Chat.ID, Date: int64(update.Message.Date), Message: update.Message.Text})

		// Voice message
		if update.Message.Voice != nil {
			audio := update.Message.Voice.FileID
			if len(audio) > 0 {
				mdb.saveFile(ctx, bb, update, audio, update.Message.Voice.MimeType)
			}
		}

		if len(update.Message.Photo) > 0 {
			largestSize := update.Message.Photo[len(update.Message.Photo)-1]
			mdb.saveFile(ctx, bb, update, largestSize.FileID, "image/jpeg")
		}

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

	var files []db.FilesForMessages

	if err := mdb.db.Find(&files, "chat_id = ?", update.Message.Chat.ID).Error; err != nil {
		log.Fatalln(err)
	}

	log.Printf("Exporting chat %d files: %d", update.Message.Chat.ID, len(files))

	exportContent, exportFormat, err := exporter.CreateExportFile(ctx, messages, files, mdb.storage)

	if err != nil {
		logAndReportError(ctx, bb, update, err)
	}

	bb.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID:   update.Message.Chat.ID,
		Document: &models.InputFileUpload{Filename: fmt.Sprintf("%d_%d%s", update.Message.ID, update.Message.Chat.ID, exportFormat), Data: exportContent},
		Caption:  "Export",
	})
}

func (mdb *MicroDiaryBot) handleDelete(ctx context.Context, bb *bot.Bot, update *models.Update) {
	mdb.handleExport(ctx, bb, update)
	mdb.db.Where("chat_id = ?", update.Message.Chat.ID).Unscoped().Delete(&db.Message{})
	mdb.db.Where("chat_id = ?", update.Message.Chat.ID).Unscoped().Delete(&db.FilesForMessages{})
	bb.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Cleared all messages"),
	})
}
