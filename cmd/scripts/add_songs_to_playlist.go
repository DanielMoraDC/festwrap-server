package scripts

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	types "festwrap/internal"
	httpclient "festwrap/internal/http/client"
	httpsender "festwrap/internal/http/sender"
	"festwrap/internal/playlist"
	spotifyPlaylist "festwrap/internal/playlist/spotify"
	"festwrap/internal/setlist/setlistfm"
	spotifySong "festwrap/internal/song/spotify"
)

func main() {
	spotifyAccessToken := flag.String("spotify-token", "", "Spotify access token")
	setlistfmApiKey := flag.String("setlistfm-key", "", "Setlistfm API Key")
	artist := flag.String("artist", "", "Artist to add to the playlist")
	playlistId := flag.String("playlist-id", "", "Spotify playlist identifier where to add songs")
	minSongsPerSetlist := flag.Int("min-setlist-songs", 5, "Minimum number of songs to include")
	flag.Parse()

	httpClient := &http.Client{}
	baseHttpClient := httpclient.NewBaseHTTPClient(httpClient)
	httpSender := httpsender.NewBaseHTTPRequestSender(&baseHttpClient)

	fmt.Printf("Adding latest setlist songs for %s into Spotify playlist with id %s \n", *artist, *playlistId)

	setlistRepository := setlistfm.NewSetlistFMSetlistRepository(
		*setlistfmApiKey,
		&httpSender,
	)
	songRepository := spotifySong.NewSpotifySongRepository(
		*spotifyAccessToken,
		&httpSender,
	)

	var tokenKey types.ContextKey = "token"
	playlistRepository := spotifyPlaylist.NewSpotifyPlaylistRepository(&httpSender)
	playlistRepository.SetTokenKey(tokenKey)
	playlistService := playlist.NewConcurrentPlaylistService(
		&playlistRepository,
		setlistRepository,
		songRepository,
	)
	playlistService.SetMinSongs(*minSongsPerSetlist)

	ctx := context.Background()
	ctx = context.WithValue(ctx, tokenKey, *spotifyAccessToken)
	err := playlistService.AddSetlist(ctx, *playlistId, *artist)
	if err != nil {
		message := fmt.Sprintf("Could not add songs to setlist: %v", err)
		fmt.Println(message)
		os.Exit(1)
	}
}
