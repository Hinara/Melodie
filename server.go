package main

import (
	"errors"
	"sync"

	"github.com/bwmarrin/discordgo"
)

const (
	//Stopped is the state value of the bot when stopped
	Stopped = iota
	//Playing is the state value of the bot when playing
	Playing
	//Paused is the state value of the bot when paused
	Paused
)

const (
	NoRepeat = iota
	Repeat
	RepeatOne
)

//Server is a structure containing all information about a connection to a guild
type Server struct {
	sync.Mutex
	vc         *discordgo.VoiceConnection
	reloadChan chan bool
	state      int8
	playing    int
	repeat     int8
	random     bool
	playlist   []string
	guildID    string
}

//NewServer create a server structure
func NewServer(guildID string) (s *Server) {
	s = &Server{}
	s.guildID = guildID
	return
}

func (s *Server) GuildID() string {
	return s.guildID
}

func (s *Server) Vc() *discordgo.VoiceConnection {
	s.Lock()
	defer s.Unlock()
	return s.vc
}

func (s *Server) Playlist() []string {
	s.Lock()
	defer s.Unlock()
	return s.playlist
}

func (s *Server) Playing() int {
	s.Lock()
	defer s.Unlock()
	return s.playing
}

func (s *Server) Repeat() int8 {
	s.Lock()
	defer s.Unlock()
	return s.repeat
}

func (s *Server) Random() bool {
	s.Lock()
	defer s.Unlock()
	return s.random
}

func (s *Server) State() int8 {
	s.Lock()
	defer s.Unlock()
	return s.state
}

//PlaylistAdd add toAdd to the playlist if it exist in the music list
func (s *Server) PlaylistAdd(toAdd string) error {
	for _, key := range soundList {
		if key == toAdd {
			s.Lock()
			defer s.Unlock()
			s.playlist = append(s.playlist, toAdd)
			return nil
		}
	}
	return errors.New("Music not found")
}

func (s *Server) PlaylistRemove(pos int) (string, error) {
	if pos < 0 || pos > len(serverList) {
		return "", errors.New(string(pos) + " is out of the playlist")
	}
	s.Lock()
	defer s.Unlock()
	name := s.playlist[pos]
	s.playlist = append(s.playlist[:pos], s.playlist[pos+1:]...)
	if s.playing == pos {
		if s.state != Stopped {
			s.state = Stopped
			s.reloadChan <- true
		}
		s.playing = 0
	} else if s.playing > pos {
		s.playing--
	}
	return name, nil
}

func (s *Server) Previous() error {
	s.Lock()
	defer s.Unlock()
	s.playing--
	if s.playing < 0 {
		s.playing = len(s.playlist) - 1
	}
	if s.state != Stopped {
		s.state = Playing
		s.reloadChan <- true
		return nil
	}
	return nil
}

func (s *Server) Next() error {
	s.Lock()
	defer s.Unlock()
	s.playing++
	if s.playing >= len(s.playlist) {
		s.playing = 0
	}
	if s.state != Stopped {
		s.state = Playing
		s.reloadChan <- true
		s.Unlock()
		return nil
	}
	return nil
}

func (s *Server) Select(i int) error {
	s.Lock()
	defer s.Unlock()
	if i < 0 && i >= len(s.playlist) {
		return errors.New("Invalid selection")
	}
	s.playing = i
	if s.state != Stopped {
		s.state = Playing
		s.reloadChan <- true
	}
	return nil
}

func (s *Server) SetRepeat(repeat int8) {
	s.Lock()
	defer s.Unlock()
	s.repeat = repeat
}

func (s *Server) SetRandom(random bool) {
	s.Lock()
	defer s.Unlock()
	s.random = random
}

func (s *Server) Play() error {
	s.Lock()
	defer s.Unlock()
	if s.state == Playing {
		return nil
	}
	if len(s.playlist) <= 0 {
		return errors.New("Empty playlist")
	}
	if s.vc == nil {
		return errors.New("Not in a channel")
	}
	if s.state == Paused {
		s.state = Playing
		s.reloadChan <- false
	} else if s.state == Stopped && s.reloadChan == nil {
		s.state = Playing
		s.reloadChan = make(chan bool, 0)
		go s.player()
	}
	return nil
}

func (s *Server) Pause() error {
	s.Lock()
	defer s.Unlock()
	if s.state == Stopped {
		return errors.New("Cannot pause when stopped")
	}
	if s.state == Paused {
		return errors.New("Already paused")
	}
	s.state = Paused
	s.reloadChan <- false
	return nil
}

func (s *Server) Stop() error {
	s.Lock()
	defer s.Unlock()
	if s.state == Stopped {
		return errors.New("Already stopped")
	}
	s.state = Stopped
	s.reloadChan <- true
	return nil
}

func (s *Server) SetVoiceConnection(vc *discordgo.VoiceConnection) {
	s.Lock()
	defer s.Unlock()
	s.vc = vc
	if s.state != Stopped && s.reloadChan != nil {
		s.reloadChan <- false
	}
}

func (s *Server) Disconnect() error {
	s.Lock()
	defer s.Unlock()
	if s.vc != nil {
		err := s.vc.Disconnect()
		s.vc = nil
		if err != nil {
			return err
		}
	}
	return nil
}
