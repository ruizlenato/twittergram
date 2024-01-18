package modules

import (
	"fmt"
	"log"
	"strings"
	"twittergram/twittergram/database"
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
			ChatID:    telegoutil.ID(update.CallbackQuery.Message.Chat.ID),
			MessageID: update.CallbackQuery.Message.MessageID,
			Text:      fmt.Sprintf(localization.Get("start_message", *update.CallbackQuery.Message), update.CallbackQuery.Message.Chat.FirstName, botUser.FirstName),
			ParseMode: "HTML",
			ReplyMarkup: telegoutil.InlineKeyboard(
				telegoutil.InlineKeyboardRow(
					telego.InlineKeyboardButton{
						Text:         localization.Get("language_button", *update.CallbackQuery.Message),
						CallbackData: "LanguageMenu",
					},
					telego.InlineKeyboardButton{
						Text:         localization.Get("about_button", *update.CallbackQuery.Message),
						CallbackData: "about",
					},
				)),
		})
	} else {
		bot.SendMessage(&telego.SendMessageParams{
			ChatID:    telegoutil.ID(update.Message.Chat.ID),
			Text:      fmt.Sprintf(localization.Get("start_message", *update.Message), update.Message.From.FirstName, botUser.FirstName),
			ParseMode: "HTML",
			ReplyMarkup: telegoutil.InlineKeyboard(
				telegoutil.InlineKeyboardRow(
					telego.InlineKeyboardButton{
						Text:         localization.Get("language_button", *update.Message),
						CallbackData: "LanguageMenu",
					},
					telego.InlineKeyboardButton{
						Text:         localization.Get("about_button", *update.Message),
						CallbackData: "about",
					},
				)),
		})
	}

}

func About(bot *telego.Bot, query telego.CallbackQuery) {
	botUser, err := bot.GetMe()
	if err != nil {
		log.Fatal(err)
	}

	bot.EditMessageText(&telego.EditMessageTextParams{
		ChatID:    telegoutil.ID(query.Message.Chat.ID),
		MessageID: query.Message.MessageID,
		Text:      fmt.Sprintf(localization.Get("info_message", *query.Message)+localization.Get("donate_mesage", *query.Message), botUser.FirstName),
		ParseMode: "HTML",
		ReplyMarkup: telegoutil.InlineKeyboard(
			telegoutil.InlineKeyboardRow(
				telego.InlineKeyboardButton{
					Text:         localization.Get("back_button", *query.Message),
					CallbackData: "start",
				}),
		),
	})

}

func LanguageMenu(bot *telego.Bot, update telego.Update) {
	buttons := make([][]telego.InlineKeyboardButton, 0, len(database.AvailableLocales))
	for _, lang := range database.AvailableLocales {
		loaded, err := localization.Load(lang)
		if err != nil {
			log.Print(err)
		}

		buttons = append(buttons, []telego.InlineKeyboardButton{{
			Text:         loaded["lang_flag"] + loaded["lang_name"],
			CallbackData: fmt.Sprintf("setLang %s", lang),
		}})
	}

	// Get User Language
	row := database.DB.QueryRow("SELECT language FROM users WHERE id = ?;", update.CallbackQuery.Message.Chat.ID)
	var language string
	err := row.Scan(&language)
	if err != nil {
		log.Print(err)
	}

	bot.EditMessageText(&telego.EditMessageTextParams{
		ChatID:      telegoutil.ID(update.CallbackQuery.Message.Chat.ID),
		MessageID:   update.CallbackQuery.Message.MessageID,
		Text:        fmt.Sprintf(localization.Get("language_menu_mesage", *update.CallbackQuery.Message), localization.Get("lang_flag", *update.CallbackQuery.Message), localization.Get("lang_name", *update.CallbackQuery.Message)),
		ParseMode:   "HTML",
		ReplyMarkup: telegoutil.InlineKeyboard(buttons...),
	})
}

func LanguageSet(bot *telego.Bot, query telego.CallbackQuery) {
	lang := strings.ReplaceAll(query.Data, "setLang ", "")

	dbQuery := "UPDATE users SET language = ? WHERE id = ?;"
	_, err := database.DB.Exec(dbQuery, lang, query.Message.Chat.ID)
	if err != nil {
		log.Print("Error inserting user:", err)
	}

	bot.EditMessageText(&telego.EditMessageTextParams{
		ChatID:    telegoutil.ID(query.Message.Chat.ID),
		MessageID: query.Message.MessageID,
		Text:      localization.Get("language_changed_successfully", *query.Message),
		ParseMode: "HTML",
		ReplyMarkup: telegoutil.InlineKeyboard(
			telegoutil.InlineKeyboardRow(
				telego.InlineKeyboardButton{
					Text:         localization.Get("back_button", *query.Message),
					CallbackData: "start",
				}),
		),
	})

}
