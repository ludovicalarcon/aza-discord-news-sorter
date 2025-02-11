package bot

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/ludovicalarcon/aza-discord-news-sorter/cmd/todoist"
	"github.com/ludovicalarcon/aza-discord-news-sorter/internal/helpers"
)

const (
	DISCORD_TOKEN = "DISCORD_TOKEN"
	PROJECT_NAME  = "News"
)

var (
	ErrTokenNotProvided      = errors.New("DISCORD_TOKEN must be provided by env var")
	ErrApiKeyNotProvided     = errors.New("API_KEY must be provided by env var")
	ErrCouldNotRetrieveTitle = errors.New("could not retrieve title")
)

type Bot struct {
	Todo    todoist.Todoist
	session *discordgo.Session
}

func (b *Bot) sendErrorMessageToChannel(channelId, errMessage string) {
	log.Println("error:", errMessage)
	b.session.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
		Content: errMessage,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

func (b *Bot) processMessage(message, emoji string) (err error) {
	if emoji == "😍" || emoji == "👌" || emoji == "👍" || emoji == "✅" {
		title := ""

		// TODO: use go routine
		title, err = helpers.GetTitleFromUrl(message)
		if err != nil {
			return fmt.Errorf("%s for %s", ErrCouldNotRetrieveTitle.Error(), message)
		}
		err = b.Todo.CreateTodo(title, message)
	}
	return
}

func (b *Bot) messageReactionAdd(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	emoji := reaction.Emoji.Name
	channelId := reaction.ChannelID
	b.session = session
	message, err := session.ChannelMessage(channelId, reaction.MessageID)
	if err != nil {
		log.Println("error on retrieve message:", err)
		return
	}

	err = b.processMessage(message.Content, emoji)
	if err != nil {
		if err != todoist.ErrAlreadyExist {
			b.sendErrorMessageToChannel(channelId, err.Error())
			return
		}
	}
}

func (b *Bot) messageReactionRemove(session *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	emoji := reaction.Emoji.Name

	log.Println(emoji)
}

func (b *Bot) Start() (err error) {
	token := os.Getenv(DISCORD_TOKEN)
	apiKey := os.Getenv(todoist.API_KEY)

	if token == "" {
		return ErrTokenNotProvided
	}
	if apiKey == "" {
		return ErrApiKeyNotProvided
	}

	err = b.Todo.Init(PROJECT_NAME)
	if err != nil {
		return
	}

	dg, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		return
	}

	dg.AddHandler(b.messageReactionAdd)
	dg.AddHandler(b.messageReactionRemove)

	dg.Identify.Intents = discordgo.IntentGuildMessageReactions

	err = dg.Open()
	if err != nil {
		return
	}
	defer dg.Close()

	return
}
