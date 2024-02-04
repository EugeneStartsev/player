package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type SongsForPlaylist struct {
	cmdChannel chan int
	sync.Mutex
	songs []Song
}

type Song struct {
	SongId       int
	SongName     string
	SongDuration time.Duration
}

const (
	playFromCmdChannel  = 1
	pauseFromCmdChannel = 2
	exitFromCmdChannel  = 3
)

func main() {
	var playlist SongsForPlaylist
	cmdChannel := make(chan int)

	s := grpcServer{
		playlist:   &playlist,
		cmdChannel: &cmdChannel,
	}

	go playlist.playPlaylist(&cmdChannel)

	err := s.run()

	if err != nil {
		log.Fatal(err)
	}
}

func (p *SongsForPlaylist) playPlaylist(cmdChannel *chan int) {
	var valueFromCmdChannel int
	var notifyChannel <-chan time.Time
	var endTimeOfSong time.Duration
	var song *Song
	var timeStartSong time.Time

	pauseFlag := true

	for {
		select {
		case valueFromCmdChannel = <-*cmdChannel:
			if valueFromCmdChannel == playFromCmdChannel {
				if song == nil {
					song = p.getNextSong()
					endTimeOfSong = song.SongDuration
					fmt.Printf("\nСейчас играет песня: %s. Длительность: %s\n", song.SongName, endTimeOfSong)
					timeStartSong = time.Now()
					notifyChannel = time.After(endTimeOfSong)
				} else {
					fmt.Printf("\nПесня продолжила играть: %s. Осталось: %s\n", song.SongName, endTimeOfSong)
					pauseFlag = true
					timeStartSong = time.Now()
					notifyChannel = time.After(endTimeOfSong)
				}
			} else if valueFromCmdChannel == pauseFromCmdChannel {
				if pauseFlag {
					fmt.Printf("\nПесня %s поставлена на паузу. Прошло: %s\n", song.SongName, time.Since(timeStartSong))
					notifyChannel = nil
					endTimeOfSong = song.SongDuration - time.Since(timeStartSong)
					pauseFlag = false
				} else {
					fmt.Println("Плейлист уже на паузе")
				}
			} else {
				fmt.Println("\nСпасибо за прослушивание!")
				return
			}
		case <-notifyChannel:
			song = p.getNextSong()
			endTimeOfSong = song.SongDuration
			fmt.Printf("\nСейчас играет песня: %s. Длительность: %s\n", song.SongName, endTimeOfSong)
			timeStartSong = time.Now()
			notifyChannel = time.After(endTimeOfSong)
		}
	}
}

func (p *SongsForPlaylist) getNextSong() *Song {
	if len(p.songs) > 0 {
		p.Lock()

		song := p.songs[0]

		p.songs = p.songs[1:]
		p.songs = append(p.songs, song)

		p.Unlock()
		return &song
	} else {
		return nil
	}
}

func (p *SongsForPlaylist) playSong(cmdChannel *chan int) string {
	if len(p.songs) == 0 {
		return "Сейчас в плейлисте ничего нет"
	} else {
		*cmdChannel <- playFromCmdChannel
		return "Плейлист начал играть"
	}
}

func (p *SongsForPlaylist) pauseSong(cmdChannel *chan int) string {
	if len(p.songs) == 0 {
		return "Сейчас в плейлисте ничего нет, поставить на паузу нельзя"
	} else {
		*cmdChannel <- pauseFromCmdChannel
		return "Плейлист поставлен на паузу"
	}
}

func (p *SongsForPlaylist) stopPlay(cmdChannel *chan int) string {
	if len(p.songs) == 0 {
		return "Сейчас в плейлисте ничего нет"
	} else {
		*cmdChannel <- exitFromCmdChannel
		return "Спасибо за прослушивание!"
	}
}

func (p *SongsForPlaylist) addSong(songName string, songDuration string) (int, error) {
	var song Song

	var err error

	song.SongId = len(p.songs) + 1
	song.SongName = songName

	song.SongDuration, err = time.ParseDuration(songDuration)
	if err != nil {
		return 0, err
	}

	p.Lock()

	p.songs = append(p.songs, song)

	p.Unlock()

	fmt.Println(song)

	return song.SongId, nil
}

func (p *SongsForPlaylist) deleteSong(songId int) (string, error) {
	var err error

	p.Lock()

	songName := p.songs[songId-1].SongName

	p.songs, err = removeSongFromPlaylist(p.songs, songId-1)
	if err != nil {
		return "", err
	}

	p.Unlock()

	return songName, nil
}

func (p *SongsForPlaylist) showSongs() []Song {
	var songs []Song

	p.Lock()

	for _, song := range p.songs {
		songs = append(songs, song)
	}

	p.Unlock()

	return songs
}

func removeSongFromPlaylist(playlist []Song, i int) ([]Song, error) {
	if i >= len(playlist) || i < 0 {
		return nil, fmt.Errorf("индекс вышел за границы плейлиста. Значение индекса %d, а длина плейлиста %d",
			i, len(playlist))
	}

	return append(playlist[:i], playlist[i+1:]...), nil
}
