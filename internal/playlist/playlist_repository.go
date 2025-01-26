package playlist

import (
	"context"
	"festwrap/internal/song"
)

type PlaylistRepository interface {
	CreatePlaylist(ctx context.Context, playlist Playlist) (string, error)
	SearchPlaylist(ctx context.Context, name string, limit int) ([]Playlist, error)
	AddSongs(ctx context.Context, playlistId string, songs []song.Song) error
}
