package main

import (
	"context"
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
)

type Event struct {
	Date        string
	Time        string
	Description string
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
			Session:         sessionMaker.SqlSession(sqlite.Open("goout.db")),
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

	// 4. Запрашиваем последние 10 сообщений
	msgs, err := tgCtx.Raw.MessagesGetHistory(context.TODO(), &tg.MessagesGetHistoryRequest{
		Peer:  inputPeer,
		Limit: 10,
	})
	if err != nil {
		log.Fatal("failed to get message history:", err)
	}

	var messages []string

	// 5. Выводим сообщения
	for _, msg := range msgs.(*tg.MessagesChannelMessages).Messages {
		if message, ok := msg.(*tg.Message); ok {
			fmt.Printf("Message ID: %d | Text: %s\n", message.ID, message.Message)

			messages = append(messages, message.Message)
		}
	}

	events := extractEvents(messages)
	spew.Dump(events)
}
