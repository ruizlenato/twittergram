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
			Text:      fmt.Sprintf(localization.Get("start_message_private", *update.CallbackQuery.Message), update.CallbackQuery.Message.Chat.FirstName, botUser.FirstName),
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
		if strings.Contains(update.Message.Chat.Type, "group") {
			bot.SendMessage(&telego.SendMessageParams{
				ChatID:    telegoutil.ID(update.Message.Chat.ID),
				Text:      fmt.Sprintf(localization.Get("start_message_group", *update.Message), botUser.FirstName),
				ParseMode: "HTML",
				ReplyMarkup: telegoutil.InlineKeyboard(telegoutil.InlineKeyboardRow(
					telego.InlineKeyboardButton{
						Text: localization.Get("start_button", *update.Message),
						URL:  fmt.Sprintf("https://t.me/%s?start=start", botUser.Username),
					})),
			})
			return
		}
		bot.SendMessage(&telego.SendMessageParams{
			ChatID:    telegoutil.ID(update.Message.Chat.ID),
			Text:      fmt.Sprintf(localization.Get("start_message_private", *update.Message), update.Message.From.FirstName, botUser.FirstName),
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

	// Query the database to retrieve the language info based on the chat type.
	row := database.DB.QueryRow("SELECT language FROM users WHERE id = ?;", update.CallbackQuery.Message.Chat.ID)
	if strings.Contains(update.CallbackQuery.Message.Chat.Type, "group") {
		row = database.DB.QueryRow("SELECT language FROM groups WHERE id = ?;", update.CallbackQuery.Message.Chat.Type)
	}
	var language string        // Variable to store the language information retrieved from the database.
	err := row.Scan(&language) // Scan method to retrieve the value of the "language" column from the query result.
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

// LanguageSet updates the language preference for a user or a group based on the provided CallbackQuery.
// It retrieves the language information from the CallbackQuery data, determines the appropriate database table (users or groups),
// and updates the language for the corresponding user or group in the database.
func LanguageSet(bot *telego.Bot, query telego.CallbackQuery) {
	lang := strings.ReplaceAll(query.Data, "setLang ", "")

	// Determine the appropriate database table based on the chat type.
	dbQuery := "UPDATE users SET language = ? WHERE id = ?;"
	if strings.Contains(query.Message.Chat.Type, "group") {
		dbQuery = "UPDATE groups SET language = ? WHERE id = ?;"
	}
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
