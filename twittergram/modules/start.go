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
		i18n := localization.Get(message.Chat)
		bot.EditMessageText(&telego.EditMessageTextParams{
			ChatID:    telegoutil.ID(update.CallbackQuery.Message.GetChat().ID),
			MessageID: update.CallbackQuery.Message.GetMessageID(),
			Text:      fmt.Sprintf(i18n("start_message_private"), message.Chat.FirstName, botUser.FirstName),
			ParseMode: "HTML",
			ReplyMarkup: telegoutil.InlineKeyboard(
				telegoutil.InlineKeyboardRow(
					telego.InlineKeyboardButton{
						Text:         i18n("language_button"),
						CallbackData: "LanguageMenu",
					},
					telego.InlineKeyboardButton{
						Text:         i18n("about_button"),
					},
				),
				telegoutil.InlineKeyboardRow(
					telego.InlineKeyboardButton{
						Text:         i18n("help_button"),
						CallbackData: "HelpMenu",
					},
				),
			),
		})
	} else {
		i18n := localization.Get(update.Message.Chat)
		if strings.Contains(update.Message.Chat.Type, "group") {
			bot.SendMessage(&telego.SendMessageParams{
				ChatID:    telegoutil.ID(update.Message.Chat.ID),
				Text:      fmt.Sprintf(i18n("start_message_group"), botUser.FirstName),
				ParseMode: "HTML",
				ReplyMarkup: telegoutil.InlineKeyboard(telegoutil.InlineKeyboardRow(
					telego.InlineKeyboardButton{
						Text: i18n("start_button"),
						URL:  fmt.Sprintf("https://t.me/%s?start=start", botUser.Username),
					})),
			})
			return
		}
		bot.SendMessage(&telego.SendMessageParams{
			ChatID:    telegoutil.ID(update.Message.Chat.ID),
			Text:      fmt.Sprintf(i18n("start_message_private"), update.Message.From.FirstName, botUser.FirstName),
			ParseMode: "HTML",
			ReplyMarkup: telegoutil.InlineKeyboard(
				telegoutil.InlineKeyboardRow(
					telego.InlineKeyboardButton{
						Text:         i18n("language_button"),
						CallbackData: "LanguageMenu",
					},
					telego.InlineKeyboardButton{
						Text:         i18n("about_button"),
					},
				),
				telegoutil.InlineKeyboardRow(
					telego.InlineKeyboardButton{
						Text:         i18n("help_button"),
						CallbackData: "HelpMenu",
					},
				),
			),
		})
	}
}

func About(bot *telego.Bot, query telego.CallbackQuery) {
	botUser, err := bot.GetMe()
	if err != nil {
		log.Fatal(err)
	}
	i18n := localization.Get(query.Message.GetChat())

	bot.EditMessageText(&telego.EditMessageTextParams{
		ChatID:    telegoutil.ID(query.Message.GetChat().ID),
		MessageID: query.Message.GetMessageID(),
		Text:      fmt.Sprintf(i18n("info_message")+i18n("donate_mesage"), botUser.FirstName),
		ParseMode: "HTML",
		ReplyMarkup: telegoutil.InlineKeyboard(
			telegoutil.InlineKeyboardRow(
				telego.InlineKeyboardButton{
					Text:         i18n("back_button"),
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
		loaded, ok := localization.LangCache[lang]
		if !ok {
			log.Fatalf("Language '%s' not found in the cache.", lang)
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

	i18n := localization.Get(chat)

	if update.Message == nil {
		bot.EditMessageText(&telego.EditMessageTextParams{
			ChatID:      telegoutil.ID(chat.ID),
			MessageID:   update.CallbackQuery.Message.GetMessageID(),
			Text:        fmt.Sprintf(i18n("language_menu_mesage"), i18n("lang_flag"), i18n("lang_name")),
			ParseMode:   "HTML",
			ReplyMarkup: telegoutil.InlineKeyboard(buttons...),
		})
	} else {
		bot.SendMessage(&telego.SendMessageParams{
			ChatID:      telegoutil.ID(update.Message.Chat.ID),
			Text:        fmt.Sprintf(i18n("language_menu_mesage"), i18n("lang_flag"), i18n("lang_name")),
			ParseMode:   "HTML",
			ReplyMarkup: telegoutil.InlineKeyboard(buttons...),
		})
	}
}

// LanguageSet updates the language preference for a user or a group based on the provided CallbackQuery.
// It retrieves the language information from the CallbackQuery data, determines the appropriate database table (users or groups),
// and updates the language for the corresponding user or group in the database.
func LanguageSet(bot *telego.Bot, update telego.Update) {
	i18n := localization.Get(update.CallbackQuery.Message.GetChat())
	if !checkAdmin(bot, update) {
		if update.Message == nil {
			bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            i18n("no_admin"),
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
			Text:         i18n("back_button"),
			CallbackData: "start",
		}})
	}

	bot.EditMessageText(&telego.EditMessageTextParams{
		ChatID:      telegoutil.ID(update.CallbackQuery.Message.GetChat().ID),
		MessageID:   update.CallbackQuery.Message.GetMessageID(),
		Text:        i18n("language_changed_successfully"),
		ParseMode:   "HTML",
		ReplyMarkup: telegoutil.InlineKeyboard(buttons...),
	})

}
