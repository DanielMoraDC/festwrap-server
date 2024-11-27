package playlist

import "context"

type PlaylistService interface {
	AddSetlist(ctx context.Context, playlistId string, artist string) error
}
