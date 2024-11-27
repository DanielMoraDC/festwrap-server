package playlist

import (
	"context"
	"festwrap/internal/playlist/errors"
	"festwrap/internal/setlist"
	"festwrap/internal/song"
	"fmt"
)

type FetchSongResult struct {
	Song *song.Song
	Err  error
}

type ConcurrentPlaylistService struct {
	playlistRepository PlaylistRepository
	setlistRepository  setlist.SetlistRepository
	songRepository     song.SongRepository
	minSongs           int
}

func NewConcurrentPlaylistService(
	playlistRepository PlaylistRepository,
	setlistRepository setlist.SetlistRepository,
	songRepository song.SongRepository,
) ConcurrentPlaylistService {
	return ConcurrentPlaylistService{
		playlistRepository: playlistRepository,
		setlistRepository:  setlistRepository,
		songRepository:     songRepository,
		minSongs:           4,
	}
}

func (s *ConcurrentPlaylistService) AddSetlist(ctx context.Context, playlistId string, artist string) error {
	setlist, err := s.setlistRepository.GetSetlist(artist, s.minSongs)
	if err != nil {
		return err
	}

	ch := make(chan FetchSongResult)
	for _, song := range setlist.GetSongs() {
		go s.fetchSong(artist, song, ch)
	}

	songs := []song.Song{}
	for i := 0; i < len(setlist.GetSongs()); i++ {
		result := <-ch
		if result.Err == nil {
			songs = append(songs, *result.Song)
		}
	}

	if len(songs) == 0 {
		message := fmt.Sprintf("No songs to add to playlist %s", playlistId)
		return errors.NewCannotAddSongsToPlaylistError(message)
	}

	err = s.playlistRepository.AddSongs(ctx, playlistId, songs)
	if err != nil {
		return err
	}

	return nil
}

func (s *ConcurrentPlaylistService) SetMinSongs(minSongs int) {
	s.minSongs = minSongs
}

func (s *ConcurrentPlaylistService) fetchSong(artist string, song setlist.Song, ch chan<- FetchSongResult) {
	songDetails, err := s.songRepository.GetSong(artist, song.GetTitle())
	ch <- FetchSongResult{Song: songDetails, Err: err}
}
