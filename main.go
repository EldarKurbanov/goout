package main

import (
	"context"
	"errors"
	"fmt"
	"goout/web"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/davecgh/go-spew/spew"
	"github.com/glebarez/sqlite"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Event struct {
	Date        string
	Time        string
	Description string
}

type Message struct {
	ChannelID int64 `gorm:"primaryKey"`
	ID        int   `gorm:"primaryKey"`
	Text      string
}

func extractEvents(messages []string) []Event {
	var events []Event

	dateRegex := regexp.MustCompile(`(?i)(\d{1,2} (?:января|февраля|марта|апреля|мая|июня|июля|августа|сентября|октября|ноября|декабря))`)
	timeRegex := regexp.MustCompile(`(\d{1,2}:\d{2})`)

	for _, msg := range messages {
		dateMatch := dateRegex.FindString(msg)
		timeMatch := timeRegex.FindString(msg)

		if dateMatch != "" {
			description := strings.Split(msg, "\n")[0] // Берем первую строку как описание
			events = append(events, Event{
				Date:        dateMatch,
				Time:        timeMatch,
				Description: description,
			})
		}
	}
	return events
}

// http://localhost:9997/?set=code&code=XXXXX

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	wa := web.GetWebAuth()
	// start web api
	go web.Start(wa)

	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		log.Fatal("failed to parse APP_ID:", err)
	}

	dialector := sqlite.Open("goout.db")

	client, err := gotgproto.NewClient(
		// Get AppID from https://my.telegram.org/apps
		appID,
		// Get ApiHash from https://my.telegram.org/apps
		os.Getenv("API_HASH"),
		// ClientType, as we defined above
		gotgproto.ClientTypePhone(os.Getenv("PHONE_NUMBER")),
		// Optional parameters of client
		&gotgproto.ClientOpts{

			// custom authenticator using web api
			AuthConversator: wa,
			Session:         sessionMaker.SqlSession(dialector),
			Device: &telegram.DeviceConfig{
				DeviceModel:    "web",
				SystemVersion:  "web",
				AppVersion:     "0.0.1",
				SystemLangCode: "en",
				LangCode:       "en",
			},
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	defer client.Stop()

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal("failed to open database:", err)
	}

	err = db.AutoMigrate(&Message{})
	if err != nil {
		log.Fatal("failed to automigrate message:", err)
	}

	// Создаём контекст
	tgCtx := client.CreateContext()

	// 1. Разрешаем username "rogozmedia" → получаем информацию о канале
	res, err := tgCtx.Raw.ContactsResolveUsername(context.TODO(), &tg.ContactsResolveUsernameRequest{
		Username: "rogozmedia",
	})
	if err != nil {
		log.Fatal("failed to resolve username:", err)
	}

	// 2. Проверяем, что это канал
	if len(res.Chats) == 0 {
		log.Fatal("No channel found with this username")
	}

	channel, ok := res.Chats[0].(*tg.Channel)
	if !ok {
		log.Fatal("Failed to convert chat to channel")
	}

	// 3. Получаем InputPeerChannel
	inputPeer := &tg.InputPeerChannel{
		ChannelID:  channel.ID,
		AccessHash: channel.AccessHash,
	}

	// Получаем последнее сохранённое сообщение
	var lastMsg Message

	err = db.Select("id").Where("channel_id = ?", channel.ID).Order("id desc").First(&lastMsg).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatal("failed to get last message:", err)
	}

	// 4. Запрашиваем последние 10 сообщений
	msgs, err := tgCtx.Raw.MessagesGetHistory(context.TODO(), &tg.MessagesGetHistoryRequest{
		Peer:      inputPeer,
		Limit:     100,
		AddOffset: -lastMsg.ID,
	})
	if err != nil {
		log.Fatal("failed to get message history:", err)
	}

	// 5. Сохраняем новые сообщения
	for _, msg := range msgs.(*tg.MessagesChannelMessages).Messages {
		if message, ok := msg.(*tg.Message); ok {
			fmt.Printf("Message ID: %d | Text: %s\n", message.ID, message.Message)

			err = db.Create(&Message{
				ChannelID: channel.ID,
				ID:        message.ID,
				Text:      message.Message,
			}).Error
			if err != nil {
				log.Fatal("failed to create message in db:", err)
			}
		}
	}

	// 6. Получаем все сообщения из базы
	var messages []string

	err = db.Model(&Message{}).Where("channel_id = ?", channel.ID).Pluck("text", &messages).Error
	if err != nil {
		log.Fatal("failed to get messages from db:", err)
	}

	events := extractEvents(messages)
	spew.Dump(events)
}
