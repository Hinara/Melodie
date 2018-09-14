package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

func (s *Server) routineNext() {
	s.Lock()
	defer s.Unlock()
	if s.repeat == RepeatOne {
		return
	}
	if s.random {
		if s.repeat == Repeat {
			s.playing = rand.Intn(len(s.playlist))
		} else {
			s.playing = rand.Intn(len(s.playlist) + 1)
		}
	} else {
		s.playing++
	}
	if s.playing >= len(s.playlist) {
		if s.repeat != Repeat {
			s.state = Stopped
		}
		s.playing = 0
	}
}

func (s *Server) playingRoutine(soundDataChan chan []byte) {
	// 960 packet size 24000 Frequency forced by discord 2 is for the number of channel
	ticker := time.NewTicker(time.Millisecond * time.Duration(960/(2*24000/1000)))
	vc := s.Vc()
	defer ticker.Stop()
	skip := s.State() == Paused
	for {
		select {
		case reload := <-s.reloadChan:
			fmt.Println(reload)
			if reload {
				return
			}
			skip = s.State() == Paused
		case <-ticker.C:
			if !skip {
				data, ok := <-soundDataChan
				if !ok {
					s.routineNext()
					return
				}
				vc.OpusSend <- data
			}
		}
	}
}

func (s *Server) player() {
	defer func() {
		s.Lock()
		close(s.reloadChan)
		s.reloadChan = nil
		s.Unlock()
	}()
	for s.State() != Stopped {
		soundDataChan := make(chan []byte)
		stopChan := make(chan bool, 1)
		go reader(s.Playlist()[s.Playing()], soundDataChan, stopChan)
		s.playingRoutine(soundDataChan)
		close(stopChan)
	}
}

func reader(audioName string, soundDataChan chan []byte, stopChan chan bool) {
	defer func() {
		close(soundDataChan)
	}()
	file, err := os.Open(soundPath + audioName + ".dca")
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return
	}

	var opuslen int16

	for {
		select {
		case <-stopChan:
			file.Close()
			return
		default:
			// Read opus frame length from dca file.
			err = binary.Read(file, binary.LittleEndian, &opuslen)

			// If this is the end of the file, just return.
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				file.Close()
				return
			}

			if err != nil {
				file.Close()
				fmt.Println("Error reading from dca file :", err)
				return
			}

			// Read encoded pcm from dca file.
			InBuf := make([]byte, opuslen)
			err = binary.Read(file, binary.LittleEndian, &InBuf)

			// Should not be any end of file errors
			if err != nil {
				file.Close()
				fmt.Println("Error reading from dca file :", err)
				return
			}

			// Append encoded pcm data to the buffer.
			soundDataChan <- InBuf
		}
	}
}
