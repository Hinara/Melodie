package main

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
)

const (

	/* Location of the sound */
	soundPath = "./sound/"

	/* Bot prefix */
	prefix = "!mb"

	/* The help text */
	help = "The available commands are :\n" +
		"-`" + prefix + " state` to display the current bot state on this channel\n" +
		"-`" + prefix + " add 'soundname'` to add a music to the paylist\n" +
		"-`" + prefix + " remove 'playlist_music_number'` to remove a sound from the playlist\n" +
		//"-`" + prefix + " reset` reset the bot (playlist and connection)\n" +
		"-`" + prefix + " music-list` display the list of the available sounds\n" +
		"-`" + prefix + " join` make me join your channel\n" +
		"-`" + prefix + " play [playlist_music_number]` to start/resume the music\n" +
		"-`" + prefix + " pause` to pause the music\n" +
		"-`" + prefix + " stop` to stop the music\n" +
		"-`" + prefix + " next` to play the next music\n" +
		"-`" + prefix + " previous` to play the previous music\n" +
		"-`" + prefix + " repeat on|off` to set repeat on or off\n" +
		"-`" + prefix + " random on|off` to set random on or off\n" +
		"-`" + prefix + " help` to display this help\n" +
		"Bot owner only :\n" +
		//"-`" + prefix + " full-reset` to reset the bot on all Guilds and the music list\n" +
		"-`" + prefix + " shutdown` to shutdown the bot\n"
)

func command(message []string, server *Server, authorID string, s *discordgo.Session) string {
	if message[1] == "help" {
		return help
	}
	if message[1] == "add" {
		return commandAdd(message, server)
	}
	if message[1] == "remove" {
		return commandRemove(message, server)
	}
	if message[1] == "music-list" {
		return commandMusicList(message)
	}
	if message[1] == "join" {
		return commandJoin(message, server, authorID, s)
	}
	if message[1] == "play" {
		return commandPlay(message, server)
	}
	if message[1] == "stop" {
		return commandStop(message, server)
	}
	if message[1] == "pause" {
		return commandPause(message, server)
	}
	if message[1] == "next" {
		return commandNext(message, server)
	}
	if message[1] == "previous" {
		return commandPrevious(message, server)
	}
	if message[1] == "repeat" {
		return commandRepeat(message, server)
	}
	if message[1] == "random" {
		return commandRepeat(message, server)
	}
	if message[1] == "state" {
		return commandState(server)
	}
	if message[1] == "reset" {
		//return commandPlay(message, guildID)
	}
	if message[1] == "full-reset" {
		//return commandFullReset(message, authorID)
	}
	if message[1] == "shutdown" {
		return commandShutdown(message, authorID)
	}
	return "There is no `" + message[1] + "` command.\n Try \"!musicbot help\" to display the available commands"
}

func commandState(s *Server) string {
	state := ""
	switch s.State() {
	case Playing:
		state = "▶"
		break
	case Paused:
		state = "⏸"
		break
	case Stopped:
		state = "⏹"
		break
	}
	repeat := s.Repeat()
	if repeat == Repeat {
		state += "🔁"
	}
	if repeat == RepeatOne {
		state += "🔂"
	}
	if s.Random() {
		state += "🔀"
	}
	vcStatus := "❎"
	if s.Vc() != nil {
		vcStatus = "✅"
	}
	text := "**Status:** " + state + "\n" +
		"**Voice Connection:** " + vcStatus + "\n" +
		"**Playlist:**"
	playlist := s.Playlist()
	playing := s.Playing()
	for id, sound := range playlist {
		text += "\n" + strconv.Itoa(id) + ". "
		if id == playing {
			text += "**" + sound + "**"
		} else {
			text += sound
		}
	}
	if len(playlist) == 0 {
		text += " Empty"
	}
	return text

}

func commandAdd(message []string, s *Server) string {
	if len(message) != 3 {
		return "The correct use of `add` command is `" + prefix + " add 'musicname'`"
	}
	err := s.PlaylistAdd(message[2])
	if err != nil {
		return err.Error()
	}
	return "The music `" + message[2] + "` has been added to the playlist"
}

func commandRemove(message []string, s *Server) string {
	if len(message) != 3 {
		return "The correct use of `remove` command is `" + prefix + " remove 'number_in_the_playlist'`"
	}
	n, err := strconv.Atoi(message[2])
	if err != nil {
		return "\"" + message[2] + "\" is not a number"
	}
	name, err := s.PlaylistRemove(n)
	if err != nil {
		return err.Error()
	}
	return "The sound `" + name + "` has been removed"
}

func commandShutdown(message []string, authorID string) string {
	if authorID == owner {
		if len(message) == 2 {
			shutdownChan <- true
			return "See you soon :wink:"
		}
		return "The correct use of shutdown command is `" + prefix + " shutdown`"
	}
	return "Only bot owner can access to this command"
}

func commandMusicList(message []string) string {
	if len(message) == 2 {
		if len(soundList) == 0 {
			return "The current sound list is empty.\nAsk the bot owner to add some sounds for more fun !"
		}
		soundListText := "The current music list is :"
		for _, sound := range soundList {
			soundListText = soundListText + "\n" + "-" + sound
		}
		return soundListText
	}
	return "The correct use of music-list command is `" + prefix + " music-list`"
}

func commandJoin(message []string, server *Server, authorID string, s *discordgo.Session) string {
	if len(message) == 2 {
		channelID, err := getAuthorChannel(s, authorID, server.GuildID())
		if err != nil {
			return err.Error()
		}
		vc, err := s.ChannelVoiceJoin(server.GuildID(), channelID, false, true)
		server.SetVoiceConnection(vc)
		if err != nil {
			return "Error joining voice channel: `" + err.Error() + "`\nVerify my permissions"
		}
		return "Channel joined"
	}
	return "The correct use of `join` command is `" + prefix + " join`"
}

func commandPlay(message []string, s *Server) string {
	if len(message) != 2 && len(message) != 3 {
		return "The correct use of `play` command is `" + prefix + " play`"
	}
	if len(message) == 3 {
		n, err := strconv.Atoi(message[2])
		if err != nil {
			return "The parameter is not a valid number"
		}
		err = s.Select(n)
		if err != nil {
			return err.Error()
		}
		s.Play()
	} else {
		if s.State() == Playing {
			return "Already playing"
		}
		err := s.Play()
		if err != nil {
			return err.Error()
		}
	}
	return "Play the music"
}

func commandStop(message []string, s *Server) string {
	if len(message) != 2 {
		return "The correct use of `stop` command is `" + prefix + " stop`"
	}
	err := s.Stop()
	if err != nil {
		return err.Error()
	}
	return "Stop the music"
}

func commandPause(message []string, s *Server) string {
	if len(message) != 2 {
		return "The correct use of `pause` command is `" + prefix + " pause`"
	}
	err := s.Pause()
	if err != nil {
		return err.Error()
	}
	return "Pause the music"
}

func commandRepeat(message []string, s *Server) string {
	if len(message) == 3 {
		if message[2] == "on" {
			s.SetRepeat(Repeat)
			return "The playlist will be repeat"
		}
		if message[2] == "off" {
			s.SetRepeat(NoRepeat)
			return "The playlist will not be repeat"
		}
		if message[2] == "one" {
			s.SetRepeat(RepeatOne)
			return "The music will be repeat"
		}
	}
	return "The correct use of `repeat` command is `" + prefix + " repeat on|off|one`"
}

func commandRandom(message []string, s *Server) string {
	if len(message) == 3 {
		if message[2] == "on" {
			s.SetRandom(true)
			return "The playlist will be played randomly"
		}
		if message[2] == "off" {
			s.SetRandom(false)
			return "The playlist will be played normally"
		}
	}
	return "The correct use of `random` command is `" + prefix + " random on|off`"
}

func commandNext(message []string, s *Server) string {
	if len(message) != 2 {
		return "The correct use of `next` command is `" + prefix + " next`"
	}
	err := s.Next()
	if err != nil {
		return "I can't change of music when I'm not playing music"
	}
	return "Next music"
}

func commandPrevious(message []string, s *Server) string {
	if len(message) != 2 {
		return "The correct use of `previous` command is `" + prefix + " next`"
	}
	err := s.Previous()
	if err != nil {
		return "I can't change of music when I'm not playing music"
	}
	return "Previous music"
}

/*
func commandFullReset(message []string, authorID string) string {
	if authorID == owner {
		if len(message) == 2 {
			for serverID, server := range serverList {
				if server.state != "stop" {
					server.controlChan <- "stop"
				}
				server.vc.Disconnect()
				delete(serverList, serverID)
			}
			populateSoundCollection()
			return "The bot has been fully reseted"
		}
		return "The correct use of full-reset command is `" + prefix + " full-reset`"
	}
	return "Only bot owner can access to this command"
}
*/

/*
func commandReset(message []string, guildID string) string {
	if len(message) == 2 {
		if serverList[guildID] != nil {
			if serverList[guildID].state != "stop" {
				serverList[guildID].controlChan <- "stop"
			}
			serverList[guildID].vc.Disconnect()
			serverList[guildID] = nil
		}
		return "I'm reseted for this guild"
	}
	return "The correct use of reset command is `" + prefix + " reset`"
}
*/
