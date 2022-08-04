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

type CallResults map[string]interface{}

var weatherApiToken string
var responses map[string]string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("TOKEN")
	weatherApiToken = os.Getenv("OPENWEATHERAPIKEY")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = os.Getenv("ENV") == "DEBUG"

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	lastCities := make(map[string]string)

	responsesBytes, err := ioutil.ReadFile("./responses.json")
	if err != nil {
		log.Fatal("Error opening file: ", err)
	}
	responses = make(map[string]string)
	if err = json.Unmarshal(responsesBytes, &responses); err != nil {
		log.Fatal("Error during unmarshal: ", err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		language := update.Message.From.LanguageCode
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg.Text = GetText("start", language)
			case "about":
				msg.Text = "made by MrBorisT"
			case "hello":
				switch h := update.Message.Time().Hour(); {
				case h >= 0 && h < 6:
					msg.Text = GetText("good_night", language) + ", " + update.Message.Chat.LastName + "!"
				case h >= 6 && h < 12:
					msg.Text = GetText("good_morning", language) + ", " + update.Message.Chat.LastName + "!"
				case h >= 12 && h < 18:
					msg.Text = GetText("good_afternoon", language) + ", " + update.Message.Chat.LastName + "! "
				default:
					msg.Text = GetText("good_evening", language) + ", " + update.Message.Chat.LastName + "!"
				}
			case "last_city":
				if city, ok := lastCities[update.Message.Chat.UserName]; ok {
					SetMsgCityWeather(city, update, &msg)
				} else {
					msg.Text = GetText("last_city_fail", language)
				}
			}
		} else {
			cityName := update.Message.Text
			lastCities[update.Message.Chat.UserName] = cityName
			SetMsgCityWeather(cityName, update, &msg)
		}
		bot.Send(msg)
	}
}

func getAPIAddress(city, apiKey, lang string) string {
	address := "https://api.openweathermap.org/data/2.5/weather?q=" + city
	address += ("&lang=" + lang)
	address += ("&units=metric")
	address += ("&appid=" + weatherApiToken)
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

func FormatWeather(cr CallResults, language string) string {
	if cr == nil {
		return "Invalid request"
	}
	main := cr["main"].(map[string]interface{})
	weather := cr["weather"].([]interface{})[0]
	wind := cr["wind"].(map[string]interface{})
	clouds := cr["clouds"].(map[string]interface{})

	ans := GetText("temperature", language) + ": " + fmt.Sprintf("%.1f", main["temp"].(float64)) + "°C\n"
	ans += GetText("description", language) + ": " + weather.(map[string]interface{})["description"].(string) + "\n"
	ans += GetText("pressure", language) + ": " + fmt.Sprintf("%.0f", main["pressure"].(float64)*0.75006157584566) + " " + GetText("mmhg", language) + "\n" // convert hPa to mm Hg
	ans += GetText("humidity", language) + ": " + fmt.Sprintf("%.0f", main["humidity"].(float64)) + "%\n"
	ans += GetText("wind_speed", language) + ": " + fmt.Sprintf("%.0f", wind["speed"].(float64)) + GetText("mps", language) + "\n"
	ans += GetText("cloudiness", language) + ": " + fmt.Sprintf("%.0f", clouds["all"].(float64)) + "%\n"

	return ans
}

func SetMsgCityWeather(city string, update tgbotapi.Update, msg *tgbotapi.MessageConfig) {
	apiAddress := getAPIAddress(city, weatherApiToken, update.Message.From.LanguageCode)
	currentWeather := WeatherAPI(apiAddress)
	msg.Text = FormatWeather(currentWeather, update.Message.From.LanguageCode)
	msg.ReplyToMessageID = update.Message.MessageID
}

func GetText(id, lang string) string {
	if lang == "ru" {
		id += "_ru"
	}
	return responses[id]
}
