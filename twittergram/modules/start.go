package modules

import (
	"fmt"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func StartCmd(bot *telego.Bot, message telego.Message) {
	bot.SendMessage(telegoutil.Message(
		telegoutil.ID(message.Chat.ID),
		fmt.Sprintf("Hello %s!", message.From.FirstName),
	))
}
