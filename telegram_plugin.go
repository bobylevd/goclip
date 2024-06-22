package main

import (
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramPlugin struct {
	fileQueue *FileWatcherPlugin
	chatID    int64
	botToken  string
	bot       *tgbotapi.BotAPI
}

func NewTelegramPlugin(fileQueue *FileWatcherPlugin, config Config) *TelegramPlugin {
	return &TelegramPlugin{
		fileQueue: fileQueue,
		chatID:    config.ChatID,
		botToken:  config.BotToken,
	}
}

func (tp *TelegramPlugin) startClient() error {
	var err error
	tp.bot, err = tgbotapi.NewBotAPI(tp.botToken)
	if err != nil {
		return err
	}
	log.Debugf("Authorized on account %s", tp.bot.Self.UserName)
	return nil
}

func (tp *TelegramPlugin) postVideo(filename string) error {
	log.Debugf("posting video `%s` to chat: `%d`", filename, tp.chatID)
	video := tgbotapi.NewInputMediaVideo(tgbotapi.FilePath(filename))
	var listMediaVideoInput []interface{}
	mediaGroup := tgbotapi.NewMediaGroup(tp.chatID, append(listMediaVideoInput, video))
	_, err := tp.bot.Send(mediaGroup)
	return err
}

func (tp *TelegramPlugin) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	if err := tp.startClient(); err != nil {
		log.Fatalf("failed to start Telegram client: %v", err)
	}

	for {
		select {
		case file := <-tp.fileQueue.processedFiles:
			tp.postVideo(file)
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
