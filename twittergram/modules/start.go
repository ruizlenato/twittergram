package modules

import (
	"fmt"
	"log"
	"twittergram/twittergram/localization"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func Start(bot *telego.Bot, message telego.Message) {
	botUser, err := bot.GetMe()
	if err != nil {
		log.Fatal(err)
	}

	bot.SendMessage(telegoutil.Message(
		telegoutil.ID(message.Chat.ID),
		fmt.Sprintf(localization.Get("start_message", message), message.From.FirstName, botUser.FirstName),
	).WithParseMode(telego.ModeHTML))
}
