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
		message := update.CallbackQuery.Message.(*telego.Message)

		bot.EditMessageText(&telego.EditMessageTextParams{
			ChatID:    telegoutil.ID(update.CallbackQuery.Message.GetChat().ID),
			MessageID: update.CallbackQuery.Message.GetMessageID(),
			Text:      fmt.Sprintf(localization.Get("start_message_private", message.Chat), message.Chat.FirstName, botUser.FirstName),
			ParseMode: "HTML",
			ReplyMarkup: telegoutil.InlineKeyboard(
				telegoutil.InlineKeyboardRow(
					telego.InlineKeyboardButton{
						Text:         localization.Get("language_button", message.Chat),
						CallbackData: "LanguageMenu",
					},
					telego.InlineKeyboardButton{
						Text:         localization.Get("about_button", message.Chat),
						CallbackData: "about",
					},
				)),
		})
	} else {
		if strings.Contains(update.Message.Chat.Type, "group") {
			bot.SendMessage(&telego.SendMessageParams{
				ChatID:    telegoutil.ID(update.Message.Chat.ID),
				Text:      fmt.Sprintf(localization.Get("start_message_group", update.Message.Chat), botUser.FirstName),
				ParseMode: "HTML",
				ReplyMarkup: telegoutil.InlineKeyboard(telegoutil.InlineKeyboardRow(
					telego.InlineKeyboardButton{
						Text: localization.Get("start_button", update.Message.Chat),
						URL:  fmt.Sprintf("https://t.me/%s?start=start", botUser.Username),
					})),
			})
			return
		}
		bot.SendMessage(&telego.SendMessageParams{
			ChatID:    telegoutil.ID(update.Message.Chat.ID),
			Text:      fmt.Sprintf(localization.Get("start_message_private", update.Message.Chat), update.Message.From.FirstName, botUser.FirstName),
			ParseMode: "HTML",
			ReplyMarkup: telegoutil.InlineKeyboard(
				telegoutil.InlineKeyboardRow(
					telego.InlineKeyboardButton{
						Text:         localization.Get("language_button", update.Message.Chat),
						CallbackData: "LanguageMenu",
					},
					telego.InlineKeyboardButton{
						Text:         localization.Get("about_button", update.Message.Chat),
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
		ChatID:    telegoutil.ID(query.Message.GetChat().ID),
		MessageID: query.Message.GetMessageID(),
		Text:      fmt.Sprintf(localization.Get("info_message", query.Message.GetChat())+localization.Get("donate_mesage", query.Message.GetChat()), botUser.FirstName),
		ParseMode: "HTML",
		ReplyMarkup: telegoutil.InlineKeyboard(
			telegoutil.InlineKeyboardRow(
				telego.InlineKeyboardButton{
					Text:         localization.Get("back_button", query.Message.GetChat()),
					CallbackData: "start",
				}),
		),
	})

}

func LanguageMenu(bot *telego.Bot, update telego.Update) {
	if !checkAdmin(bot, update) {
		if update.Message == nil {
			bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            "Você não é admin",
				ShowAlert:       true,
			})
		}
		return
	}

	message := update.Message
	if message == nil {
		message = update.CallbackQuery.Message.(*telego.Message)
	}

	chat := message.GetChat()

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

	// Query the database to retrieve the language info based on the chat type.
	row := database.DB.QueryRow("SELECT language FROM users WHERE id = ?;", chat.ID)
	if strings.Contains(chat.Type, "group") {
		row = database.DB.QueryRow("SELECT language FROM groups WHERE id = ?;", chat.ID)
	}
	var language string        // Variable to store the language information retrieved from the database.
	err := row.Scan(&language) // Scan method to retrieve the value of the "language" column from the query result.
	if err != nil {
		log.Print(err)
	}

	if update.Message == nil {
		bot.EditMessageText(&telego.EditMessageTextParams{
			ChatID:      telegoutil.ID(chat.ID),
			MessageID:   update.CallbackQuery.Message.GetMessageID(),
			Text:        fmt.Sprintf(localization.Get("language_menu_mesage", chat), localization.Get("lang_flag", chat), localization.Get("lang_name", chat)),
			ParseMode:   "HTML",
			ReplyMarkup: telegoutil.InlineKeyboard(buttons...),
		})
	} else {
		bot.SendMessage(&telego.SendMessageParams{
			ChatID:      telegoutil.ID(update.Message.Chat.ID),
			Text:        fmt.Sprintf(localization.Get("language_menu_mesage", chat), localization.Get("lang_flag", chat), localization.Get("lang_name", chat)),
			ParseMode:   "HTML",
			ReplyMarkup: telegoutil.InlineKeyboard(buttons...),
		})
	}
}

// LanguageSet updates the language preference for a user or a group based on the provided CallbackQuery.
// It retrieves the language information from the CallbackQuery data, determines the appropriate database table (users or groups),
// and updates the language for the corresponding user or group in the database.
func LanguageSet(bot *telego.Bot, update telego.Update) {
	if !checkAdmin(bot, update) {
		if update.Message == nil {
			bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            localization.Get("no_admin", update.CallbackQuery.Message.GetChat()),
				ShowAlert:       true,
			})
		}
		return
	}

	lang := strings.ReplaceAll(update.CallbackQuery.Data, "setLang ", "")

	// Determine the appropriate database table based on the chat type.
	dbQuery := "UPDATE users SET language = ? WHERE id = ?;"
	if strings.Contains(update.CallbackQuery.Message.GetChat().Type, "group") {
		dbQuery = "UPDATE groups SET language = ? WHERE id = ?;"
	}
	_, err := database.DB.Exec(dbQuery, lang, update.CallbackQuery.Message.GetChat().ID)
	if err != nil {
		log.Print("Error inserting user:", err)
	}

	buttons := make([][]telego.InlineKeyboardButton, 0, len(database.AvailableLocales))

	if update.CallbackQuery.Message.GetChat().Type == telego.ChatTypePrivate {
		buttons = append(buttons, []telego.InlineKeyboardButton{{
			Text:         localization.Get("back_button", update.CallbackQuery.Message.GetChat()),
			CallbackData: "start",
		}})
	}

	bot.EditMessageText(&telego.EditMessageTextParams{
		ChatID:      telegoutil.ID(update.CallbackQuery.Message.GetChat().ID),
		MessageID:   update.CallbackQuery.Message.GetMessageID(),
		Text:        localization.Get("language_changed_successfully", update.CallbackQuery.Message.GetChat()),
		ParseMode:   "HTML",
		ReplyMarkup: telegoutil.InlineKeyboard(buttons...),
	})

}
