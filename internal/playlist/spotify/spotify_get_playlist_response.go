package spotify

type SpotifyPlaylistOwnerMetadata struct {
	Id string `json:"id"`
}

type SpotifySearchPlaylist struct {
	Description   string                       `json:"description"`
	Name          string                       `json:"name"`
	Public        bool                         `json:"public"`
	OwnerMetadata SpotifyPlaylistOwnerMetadata `json:"owner"`
}

type SpotifySearchPlaylists struct {
	Items []SpotifySearchPlaylist `json:"items"`
}

type SpotifySearchPlaylistResponse struct {
	Playlists SpotifySearchPlaylists `json:"playlists"`
}
