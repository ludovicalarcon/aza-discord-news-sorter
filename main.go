package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ludovicalarcon/aza-discord-news-sorter/cmd/bot"
	"github.com/ludovicalarcon/aza-discord-news-sorter/cmd/todoist"
)

const DEFAULT_TIMEOUT = 2

func main() {
	httpTimeout := flag.Int("timeout", DEFAULT_TIMEOUT, "http client timeout")
	bot := bot.Bot{Todo: todoist.Todoist{Client: &http.Client{Timeout: time.Duration(*httpTimeout) * time.Second}}}
	err := bot.Start()
	if err != nil {
		log.Fatalln("bot could not start", err)
	}

	log.Println("bot is now running.\nPress CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
