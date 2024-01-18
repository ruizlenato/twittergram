package localization

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"twittergram/twittergram/database"

	"github.com/mymmrac/telego"
)

var langCache sync.Map

func GetAllLocalesFiles() error {
	database.AvailableLocales = nil

	err := filepath.Walk("twittergram/localization/localizations", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".json" {
			// Get the file name without extension
			fileName := filepath.Base(path[:len(path)-len(filepath.Ext(path))])
			// Append the file name to the global variable availableLocales
			database.AvailableLocales = append(database.AvailableLocales, fileName)
		}

		return nil
	})

	return err
}

func Load(lang string) (map[string]string, error) {
	if cached, ok := langCache.Load(lang); ok {
		if langMap, ok := cached.(map[string]string); ok {
			return langMap, nil
		}
	}

	data, err := os.ReadFile(fmt.Sprintf("twittergram/localization/localizations/%s.json", lang))
	if err != nil {
		return nil, err
	}

	langMap := make(map[string]string)
	err = json.Unmarshal(data, &langMap)
	if err != nil {
		return nil, err
	}

	langCache.Store(lang, langMap)

	return langMap, nil
}

func Get(key string, message telego.Message) string {
	row := database.DB.QueryRow("SELECT language FROM users WHERE id = ?;", message.Chat.ID)
	var language string
	err := row.Scan(&language)
	if err != nil {
		log.Print(err)
	}
	loaded, err := Load(language)
	if err != nil {
		log.Fatal("Error loading language file:", err)
	}
	value, ok := loaded[key]
	if !ok {
		return "KEY_NOT_FOUND"
	}
	return value
}
