package bot

import (
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/thecodeah/taco-bot/src/commands"
	"github.com/thecodeah/taco-bot/src/storage"
)

// Configuration contains settings loaded in from environment
// variable (.env) files.
type Configuration struct {
	Token    string `required:"true"`
	Prefix   string `default:"!"`
	Cooldown int    `default:"3"`

	MongoHost string `split_words:"true" default:"127.0.0.1" required:"true"`
	MongoPort string `split_words:"true" default:"27017" required:"true"`
	MongoName string `split_words:"true" default:"tacobot" required:"true"`
	MongoUser string `split_words:"true" required:"false"`
	MongoPass string `split_words:"true" required:"false"`
}

// Bot contains information that's necessary for the bot.
type Bot struct {
	config         Configuration
	session        *discordgo.Session
	commandHandler *commands.CommandHandler
	database       *storage.Database
}

// New initializes the bot as well as all commands.
func New(config Configuration) (bot *Bot, err error) {
	bot = &Bot{
		config: config,
	}

	bot.database, err = storage.Connect(storage.Config{
		Host: config.MongoHost,
		Port: config.MongoPort,
		Name: config.MongoName,
		User: config.MongoUser,
		Pass: config.MongoPass,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to MongoDB datatabase.")
	}

	bot.session, err = discordgo.New("Bot " + config.Token)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create new Discord session")
	}

	err = bot.session.Open()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create a websocket connection to Discord")
	}

	bot.commandHandler = commands.New(bot.database, commands.Config{
		Prefix:   config.Prefix,
		Cooldown: time.Duration(config.Cooldown) * time.Second,
	})
	bot.registerCommands()

	bot.session.AddHandler(bot.onMessageCreate)

	return
}

// Close closes the bot
func (bot Bot) Close() {
	bot.session.Close()
}

func (bot Bot) registerCommands() {
	bot.commandHandler.Register("balance", commands.CommandInfo{
		Description: "Shows you how many tacos you have.",
		Function:    commands.BalanceCommand,
	})

	bot.commandHandler.Register("pay", commands.CommandInfo{
		Description: "Allows you to transfer funds to another user.",
		Function:    commands.PayCommand,
	})

	bot.commandHandler.Register("lord", commands.CommandInfo{
		Description: "Shows you the name of the richest of em' all!",
		Function:    commands.LordCommand,
	})

	bot.commandHandler.Register("ping", commands.CommandInfo{
		Description: "Send you back a 'Pong!' message!",
		Function:    commands.PingCommand,
		Hidden:      true,
	})

	bot.commandHandler.Register("help", commands.CommandInfo{
		Description: "Shows you all available commands.",
		Function:    commands.HelpCommand,
	})
}

func (bot Bot) onMessageCreate(session *discordgo.Session, info *discordgo.MessageCreate) {
	bot.commandHandler.Process(session, info)

	// Give out taco (1/100 chance)
	rand.Seed(time.Now().UnixNano())
	if random := rand.Intn(100); random == 0 {
		// Get guild
		channel, err := session.Channel(info.ChannelID)
		if err != nil {
			return
		}
		guild, err := session.Guild(channel.GuildID)
		if err != nil {
			return
		}

		user, err := bot.database.GetUser(info.Author.ID, guild.ID)
		user.Balance++
		bot.database.UpdateUser(user)
	}
}
