package commands

import (
	"fmt"
	"strconv"
)

// PayCommand allows you to transfer funds to another user.
func PayCommand(commandInfo CommandInfo) {
	if len(commandInfo.Args) != 2 {
		commandInfo.Session.ChannelMessageSend(commandInfo.Message.ChannelID,
			fmt.Sprintf("%s Usage : %spay [mention] [amount]", commandInfo.Message.Author.Mention(), commandInfo.CommandHandler.config.Prefix),
		)
	} else {
		database := commandInfo.CommandHandler.Database
		amount, err := strconv.Atoi(commandInfo.Args[1])
		if err != nil {
			return
		}

		// Get user data
		sender, err := database.GetUser(commandInfo.Message.Author.ID, commandInfo.Guild.ID)
		if err != nil {
			return
		}
		receiver, err := database.GetUser(commandInfo.Message.Mentions[0].ID, commandInfo.Guild.ID)
		if err != nil {
			return
		}

		// Perform checks
		if amount <= 0 || sender.Balance < amount || commandInfo.Message.Mentions[0].ID == commandInfo.Message.Author.ID {
			return
		}

		// Change balance and update in database
		sender.Balance -= amount
		receiver.Balance += amount
		err = database.UpdateUser(sender)
		if err != nil {
			return
		}
		err = database.UpdateUser(receiver)
		if err != nil {
			return
		}

		commandInfo.Session.ChannelMessageSend(commandInfo.Message.ChannelID,
			fmt.Sprintf("%s Successfully transferred tacos!", commandInfo.Message.Author.Mention()),
		)
	}
}