package modules

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/valyala/fasthttp"
)

func checkAdmin(bot *telego.Bot, update telego.Update) bool {
	message := update.Message
	if message == nil {
		message = update.CallbackQuery.Message.(*telego.Message)
	}

	if message.Chat.Type == telego.ChatTypePrivate {
		return true
	}

	if !strings.Contains(message.Chat.Type, "group") {
		return false
	}

	userID := message.From.ID
	if update.CallbackQuery != nil {
		userID = update.CallbackQuery.From.ID
	}

	chatMember, err := bot.GetChatMember(&telego.GetChatMemberParams{
		ChatID: telegoutil.ID(message.Chat.ID),
		UserID: userID,
	})
	if err != nil {
		log.Println(err)
		return false
	}

	if chatMember.MemberStatus() == "creator" || chatMember.MemberStatus() == "administrator" {
		return true
	}

	return false
}

func TwitterAPI(link string, query map[string]string) *fasthttp.Response {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	client := &fasthttp.Client{ReadBufferSize: 8192}

	req.Header.SetMethod("GET")
	csrfToken := strings.ReplaceAll((uuid.New()).String(), "-", "")
	headers := map[string]string{
		"Authorization":             "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
		"Cookie":                    fmt.Sprintf("auth_token=ee4ebd1070835b90a9b8016d1e6c6130ccc89637; ct0=%v; ", csrfToken),
		"x-twitter-active-user":     "yes",
		"x-twitter-auth-type":       "OAuth2Session",
		"x-twitter-client-language": "en",
		"x-csrf-token":              csrfToken,
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	req.SetRequestURI(link)
	for key, value := range query {
		req.URI().QueryArgs().Add(key, value)
	}

	err := client.Do(req, res)
	if err != nil {
		log.Fatal(err)
	}

	return res
}
