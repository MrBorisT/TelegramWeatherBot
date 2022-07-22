package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("TOKEN")
	weatherApiToken := os.Getenv("OPENWEATHERAPIKEY")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = os.Getenv("ENV") == "DEBUG"

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите город\nabout - инфо о создателе")
				bot.Send(msg)
			} else if update.Message.Text == "about" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "made by MrBorisT")
				bot.Send(msg)
			} else {
				currentWeather := WeatherAPI(getAPIAddress(update.Message.Text, weatherApiToken, update.Message.From.LanguageCode))

				weatherStr := formatWeather(currentWeather)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, weatherStr)
				msg.ReplyToMessageID = update.Message.MessageID

				bot.Send(msg)
			}
		}
	}
}

func getAPIAddress(city, apiKey, lang string) string {
	address := "https://api.openweathermap.org/data/2.5/weather?q=" + city
	address += ("&lang=" + lang)
	address += ("&units=metric")
	address += ("&appid=" + apiKey)
	return address
}

func WeatherAPI(request string) CallResults {
	var sr CallResults
	//Sending request
	if response, err := http.Get(request); err != nil {
		log.Println(err)
	} else {
		defer response.Body.Close()

		//Reading answer
		if response.StatusCode != http.StatusOK {
			return nil
		}
		contents, _ := ioutil.ReadAll(response.Body)

		//Unmarshal answer and set it to SearchResults struct
		sr = CallResults{}
		if err = json.Unmarshal([]byte(contents), &sr); err != nil {
			log.Println(err)
			return nil
		}
	}
	return sr
}

func formatWeather(cr CallResults) string {
	if cr == nil {
		return "Неверный запрос!"
	}
	main := cr["main"].(map[string]interface{})
	weather := cr["weather"].([]interface{})[0]
	wind := cr["wind"].(map[string]interface{})
	clouds := cr["clouds"].(map[string]interface{})

	ans := "Температура: " + fmt.Sprintf("%.1f", main["temp"].(float64)) + "°C\n"
	ans += "Описание: " + weather.(map[string]interface{})["description"].(string) + "\n"
	ans += "Атм. давление: " + fmt.Sprintf("%.0f", main["pressure"].(float64)*0.75006157584566) + " мм рт.ст.\n" // convert hPa to mm Hg
	ans += "Влажность воздуха: " + fmt.Sprintf("%.0f", main["humidity"].(float64)) + "%\n"
	ans += "Скорость ветра: " + fmt.Sprintf("%.0f", wind["speed"].(float64)) + "м/с\n"
	ans += "Облачность: " + fmt.Sprintf("%.0f", clouds["all"].(float64)) + "%\n"

	return ans
}

type CallResults map[string]interface{}
