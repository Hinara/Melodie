package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	/* Bot owner */
	owner string

	/* The list of the sounds found in the sound path */
	soundList = make([]string, 0)

	/* The list of connected server */
	serverList = make(map[string]*Server, 0)

	/* A Chan to close properly the bot via ""!musicbot shutdwon" */
	shutdownChan = make(chan bool)
)

/* List all sounds in the sound path and put them in the soundlist */
func populateSoundCollection() {
	//Create a temporary soundlist (in order that other routines always access to a full populated sound list)
	var tempSoundList []string
	//tempSoundList := make([]string, 0)
	//Open the sound path
	folder, err := os.Open(soundPath)
	//Read the name of all files in the directory
	files, err := folder.Readdirnames(0)
	if err != nil {
		fmt.Println("Error Reading directory :", err)
		return
	}
	// For each file verify it has the dca suffix (only dca files are accepted by the bot) and add them to the list without the ".dca" suffix (user don't need to type it)
	for _, file := range files {
		if strings.HasSuffix(file, ".dca") {
			name := strings.Replace(file, ".dca", "", -1)
			tempSoundList = append(tempSoundList, name)
		}
	}
	//Put the full populated sound list to soundList var
	soundList = tempSoundList
}

/* Main function of the program */
func main() {
	token := os.Getenv("TOKEN")
	if token == "" {
		fmt.Println("No token provided")
	}

	// Create a new Discord session using the provided token.
	dg, err := discordgo.New(token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	//Get the bot owner
	app, err := dg.Application("@me")
	if err != nil {
		fmt.Println(err)
		return
	}
	owner = app.Owner.ID

	// Fill the sound list
	populateSoundCollection()

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)

	//Prevent forced stop
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	select {
	case <-c:
	case <-shutdownChan:
		break
	}

	// Simple way to keep program running until CTRL-C is pressed or shutdown command entered
	for _, server := range serverList {
		server.Disconnect()
	}
	dg.Logout()
	dg.Close()
	fmt.Println("Goodbye ^^")
	return
}

func guildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	serverList[g.ID] = NewServer(g.ID)
}

func guildDelete(s *discordgo.Session, g *discordgo.GuildDelete) {
	delete(serverList, g.ID)
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, prefix) {
		message := strings.Split(m.Content, " ")

		//Get the channel object from the channelID
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			// Could not find channel.
			return
		}
		server, exist := serverList[c.GuildID]
		if exist == false {
			// Could not find guild
			return
		}
		//Verify there is a command after the suffix
		if len(message) > 1 {
			msg := command(message, server, m.Author.ID, s)
			if msg != "" {
				_, _ = s.ChannelMessageSend(m.ChannelID, msg)
			}
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, help)
		}
	}
}

func getAuthorChannel(s *discordgo.Session, authorID, guildID string) (string, error) {
	g, err := s.State.Guild(guildID)
	if err != nil {
		return "", errors.New("Error getting guild information")
	}
	for _, vs := range g.VoiceStates {
		if vs.UserID == authorID {
			return vs.ChannelID, nil
		}
	}
	return "", errors.New("User not found")
}
