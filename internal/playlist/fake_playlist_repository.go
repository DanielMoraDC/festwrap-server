package playlist

import (
	"context"
	"festwrap/internal/song"
)

type AddSongsArgs struct {
	Context    context.Context
	PlaylistId string
	Songs      []song.Song
}

type CreatePlaylistArgs struct {
	Context  context.Context
	UserId   string
	Playlist Playlist
}

type SearchPlaylistArgs struct {
	Context      context.Context
	PlaylistName string
	Limit        int
}

type FakePlaylistRepository struct {
	addSongArgs        AddSongsArgs
	createPlaylistArgs CreatePlaylistArgs
	searchPlaylistArgs SearchPlaylistArgs
	searchedPlaylists  []Playlist
	err                error
}

func NewFakePlaylistRepository() FakePlaylistRepository {
	return FakePlaylistRepository{searchedPlaylists: []Playlist{}}
}

func (s *FakePlaylistRepository) CreatePlaylist(ctx context.Context, userId string, playlist Playlist) error {
	s.createPlaylistArgs = CreatePlaylistArgs{Context: ctx, UserId: userId, Playlist: playlist}
	return s.err
}

func (s *FakePlaylistRepository) SearchPlaylist(
	ctx context.Context, playlistName string, limit int,
) ([]Playlist, error) {
	s.searchPlaylistArgs = SearchPlaylistArgs{Context: ctx, PlaylistName: playlistName, Limit: limit}
	return s.searchedPlaylists, s.err
}

func (s *FakePlaylistRepository) AddSongs(ctx context.Context, playlistId string, songs []song.Song) error {
	s.addSongArgs = AddSongsArgs{Context: ctx, PlaylistId: playlistId, Songs: songs}
	return s.err
}

func (s *FakePlaylistRepository) SetError(err error) {
	s.err = err
}

func (s *FakePlaylistRepository) GetAddSongArgs() AddSongsArgs {
	return s.addSongArgs
}

func (s *FakePlaylistRepository) GetCreatePlaylistSongArgs() CreatePlaylistArgs {
	return s.createPlaylistArgs
}

func (s *FakePlaylistRepository) GetSearchPlaylistArgs() SearchPlaylistArgs {
	return s.searchPlaylistArgs
}

func (s *FakePlaylistRepository) SetSearchedPlaylists(playlists []Playlist) {
	s.searchedPlaylists = playlists
}
