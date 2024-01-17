package modules

import (
	"fmt"
	"log"
	"twittergram/twittergram/localization"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func Start(bot *telego.Bot, update telego.Update) {
	botUser, err := bot.GetMe()
	if err != nil {
		log.Fatal(err)
	}
	if update.Message == nil {
		bot.EditMessageText(&telego.EditMessageTextParams{
			ChatID:      telegoutil.ID(update.CallbackQuery.Message.Chat.ID),
			MessageID:   update.CallbackQuery.Message.MessageID,
			Text:        fmt.Sprintf(localization.Get("start_message", *update.CallbackQuery.Message), update.CallbackQuery.Message.From.FirstName, botUser.FirstName),
			ParseMode:   "HTML",
			ReplyMarkup: telegoutil.InlineKeyboard(telegoutil.InlineKeyboardRow(telegoutil.InlineKeyboardButton(localization.Get("about_button", *update.CallbackQuery.Message)).WithCallbackData("about"))),
		})
	} else {
		bot.SendMessage(&telego.SendMessageParams{
			ChatID:      telegoutil.ID(update.Message.Chat.ID),
			Text:        fmt.Sprintf(localization.Get("start_message", *update.Message), update.Message.From.FirstName, botUser.FirstName),
			ParseMode:   "HTML",
			ReplyMarkup: telegoutil.InlineKeyboard(telegoutil.InlineKeyboardRow(telegoutil.InlineKeyboardButton(localization.Get("about_button", *update.Message)).WithCallbackData("about"))),
		})
	}

}

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
