package playlist

import (
	"context"
	"festwrap/internal/song"
)

type PlaylistRepository interface {
	CreatePlaylist(ctx context.Context, userId string, playlist Playlist) error
	AddSongs(ctx context.Context, playlistId string, songs []song.Song) error
}
