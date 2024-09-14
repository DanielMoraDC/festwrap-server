package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	types "festwrap/internal"
	spotifyArtist "festwrap/internal/artist/spotify"
	httpclient "festwrap/internal/http/client"
	httpsender "festwrap/internal/http/sender"
)

func main() {
	spotifyAccessToken := flag.String("spotify-token", "", "Spotify access token")
	artist := flag.String("artist", "", "Artist to search for")
	limit := flag.Int("limit", 5, "Number of results to retrieve")
	flag.Parse()

	httpClient := &http.Client{}
	baseHttpClient := httpclient.NewBaseHTTPClient(httpClient)
	httpSender := httpsender.NewBaseHTTPRequestSender(&baseHttpClient)

	fmt.Printf("Searching for artist %s into Spotify API, retrieving %d results at most\n", *artist, *limit)

	tokenKey := types.ContextKey("myToken")
	artistRepository := spotifyArtist.NewSpotifyArtistRepository(&httpSender)
	artistRepository.SetTokenKey(tokenKey)

	ctx := context.Background()
	ctx = context.WithValue(ctx, tokenKey, *spotifyAccessToken)
	artists, err := artistRepository.SearchArtist(ctx, *artist, *limit)
	if err != nil {
		message := fmt.Sprintf("Error searching artist: %v", err)
		fmt.Println(message)
		os.Exit(1)
	}

	fmt.Printf("Found %v\n", artists)
}
