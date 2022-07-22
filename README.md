# Weather telegram bot

## About
A little telegram bot that responds with weather info to entered city. The description of info is in russian, but city could be written in any language. 
Uses [this api](https://github.com/go-telegram-bot-api/telegram-bot-api) and [OpenWeather](https://openweathermap.org/current)

## Init Setup

1. Make *.env* file with keys **TOKEN**, **ENV**, **OPENWEATHERAPIKEY**
2. Get API key for telegram bot from [BotFather](https://telegram.me/BotFather)
3. Get API key for current weather data from [OpenWeather](https://openweathermap.org/)
4. If **ENV** is set to **DEBUG**, you'll get log data
5. Set **TOKEN** to API key from step 2 and set **OPENWEATHERAPIKEY** to API key from step 3
