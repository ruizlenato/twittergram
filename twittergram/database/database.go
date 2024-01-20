package database

import (
	"database/sql"
	"fmt"
	"log"
	"slices"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

var (
	DB               *sql.DB
	AvailableLocales []string
)

func Open(databaseFile string) error {
	db, err := sql.Open("sqlite3", databaseFile+"?_journal_mode=WAL")
	if err != nil {
		return err
	}

	// Check if journal_mode is set to WAL
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		db.Close()
		return err
	}
	DB = db

	return nil
}

func CreateTables() error {
	query := `
        CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
            language TEXT DEFAULT 'en-us',
            twitter_username TEXT
        );
		CREATE TABLE IF NOT EXISTS groups (
            id INTEGER PRIMARY KEY,
            language TEXT DEFAULT 'en-us'
        );
    `
	_, err := DB.Exec(query)
	return err
}

func Close() {
	fmt.Println("Database closed")
	if DB != nil {
		DB.Close()
	}
}

func SaveUsers(bot *telego.Bot, update telego.Update, next telegohandler.Handler) {
	Message := update.Message
	if Message == nil {
		Message = update.CallbackQuery.Message
	}

	if Message.SenderChat != nil {
		return
	}

	if Message.From.ID != Message.Chat.ID {
		query := "INSERT OR IGNORE INTO groups (id) VALUES (?);"
		_, err := DB.Exec(query, Message.Chat.ID)
		if err != nil {
			log.Print("Error inserting group:", err)
		}
	}

	query := "INSERT OR IGNORE INTO users (id, language) VALUES (?, ?);"
	lang := Message.From.LanguageCode
	if !slices.Contains(AvailableLocales, lang) {
		lang = "en-us"
	}
	_, err := DB.Exec(query, Message.From.ID, lang)
	if err != nil {
		log.Print("Error inserting user:", err)
	}
	next(bot, update)
}
