/*
 * Copyright (c) 2023-2023 curoky(cccuroky@gmail.com).
 *
 * This file is part of telegram-bot-vercel-go-example.
 * See https://github.com/curoky/telegram-bot-vercel-go-example for further info.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
)

var onceSetup sync.Once
var gBot *tele.Bot

func createBot(token string) *tele.Bot {
	bot, err := tele.NewBot(tele.Settings{
		Token:       token,
		Poller:      &tele.LongPoller{Timeout: 10 * time.Second},
		Synchronous: true,
	})
	if err != nil {
		log.Panic(err)
	}
	bot.Handle("/hello", func(c tele.Context) error {
		return c.Send("Hello!")
	})

	return bot
}

func Handler(w http.ResponseWriter, r *http.Request) {
	onceSetup.Do(func() {
		log.Println("Start to setup bot")
		viper.AutomaticEnv()
		gBot = createBot(viper.GetString("TELEGRAM_TOKEN"))
	})
	log.Println("Receive request")

	if r.Method == "GET" {
		webhookUrl := "telegram-bot-vercel-go-example.vercel.app/api/index"
		err := gBot.SetWebhook(&tele.Webhook{Endpoint: &tele.WebhookEndpoint{PublicURL: webhookUrl}})
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		wh, err := gBot.Webhook()
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		_, _ = w.Write([]byte(fmt.Sprintf("Set webhook on %v\n", wh.Listen)))
	} else if r.Method == "POST" {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("Received body: %v\n", len(body))
		u := &tele.Update{}
		err = json.Unmarshal(body, u)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("Process update: %v(%v)\n", u.Message.Sender.Username, u.Message.Text)

		if u.Message != nil && u.Message.Text != "" && !strings.HasPrefix(u.Message.Text, "/") {
			c := gBot.NewContext(*u)
			_ = c.Reply(u.Message.Text)
		} else {
			gBot.ProcessUpdate(*u)
		}

		log.Println("Process finished!")
	}
}
