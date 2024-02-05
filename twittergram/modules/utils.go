package modules

import (
	"log"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
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
