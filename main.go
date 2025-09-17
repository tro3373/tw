package main

import (
	"bufio"
	"log/slog"
	"os"

	twscraper "github.com/imperatrona/twitter-scraper"
)

func main() {
	client, err := NewXClient()
	if err != nil {
		panic(err)
	}

	fm, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	if (fm.Mode() & os.ModeCharDevice) != 0 {
		panic("no input")
	}
	// pipe (if stdin is not charactor device(like keyboard, etc))
	scanner := bufio.NewScanner(os.Stdin)
	message := ""
	for scanner.Scan() {
		message += scanner.Text() + "\n"
	}
	nt := twscraper.NewTweet{
		Text: message,
	}
	tweet, err := client.Scraper.CreateTweet(nt)
	if err != nil {
		panic(err)
	}
	slog.Info("tweeted!", "url", tweet.PermanentURL)
}
