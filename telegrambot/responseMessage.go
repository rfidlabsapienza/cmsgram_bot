package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

type InlineKeyboard struct {
	Text          string `json:"text"`
	Callback_data string `json:"callback_data"`
}

type ReplyMarkup struct {
	Inline_keyboard [][]InlineKeyboard `json:"inline_keyboard"`
}

type ResponseMessage struct {
	Chat_id                  int64        `json:"chat_id"`
	Text                     string       `json:"text"`
	Disable_web_page_preview bool         `json:"disable_web_page_preview"`
	Parse_mode               string       `json:"parse_mode"`
	Reply_markup             *ReplyMarkup `json:"reply_markup,omitempty"`
}

func (rm *ResponseMessage) AddCallbackButton(text, data string) {
	var (
		buttons [][]InlineKeyboard
		row1    []InlineKeyboard = []InlineKeyboard{{text, data}}
	)
	buttons = append(buttons, row1)

	rm.Reply_markup = &ReplyMarkup{buttons}
}

// Sends message 'text' to the the specified chat (an ID)
func (rm *ResponseMessage) Send(text string, to int) (err error) {
	var (
		response = struct {
			Ok     bool
			Result Message
		}{}
		url = config.BotAPIBaseURL + config.BotAPIToken + "/sendMessage"
	)

	conn := pool.Get()
	defer conn.Close()

	chat, err := redis.Int64(conn.Do("GET", "tgbot:user:chat:"+strconv.Itoa(to)))
	if err != nil {
		return
	}

	// Initialize message
	rm.Chat_id = chat
	rm.Text = text
	rm.Disable_web_page_preview = true
	rm.Parse_mode = "HTML"

	// Encode data into JSON
	payload, err := json.Marshal(rm)
	if err != nil {
		return
	}

	// Send the payload to the BotAPI
	res, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	res.Body.Close()

	// Decode the JSON payload
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println("Send(): json.Unmarshal():", err)
		return
	}

	if !response.Ok {
		log.Println("Send(): Invalid request", response)
		return
	}

	return
}
