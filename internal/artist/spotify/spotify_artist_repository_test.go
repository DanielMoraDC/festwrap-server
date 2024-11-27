package spotify

import (
	"context"
	"errors"
	"path/filepath"

	types "festwrap/internal"
	"festwrap/internal/artist"
	httpsender "festwrap/internal/http/sender"
	"festwrap/internal/serialization"
	"festwrap/internal/testtools"
	"testing"
)

func defaultSearchName() string {
	return "Movements"
}

func defaultLimit() int {
	return 2
}

func defaultTokenKey() types.ContextKey {
	return "myKey"
}

func defaultContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, defaultTokenKey(), "some_token")
	return ctx
}

func expectedHttpOptions() httpsender.HTTPRequestOptions {
	url := "https://api.spotify.com/v1/search?limit=2&q=artist%3AMovements&type=artist"
	options := httpsender.NewHTTPRequestOptions(url, httpsender.GET, 200)
	options.SetHeaders(
		map[string]string{"Authorization": "Bearer some_token"},
	)
	return options
}

func defaultResponse() []byte {
	return []byte(`
		{
			"artists":
			{
				"items":[
					{"name":"Movements"},
					{"name":"The Movement","images":[{"url":"https://some.url1"}, {"url":https://some.url2"}]}
				]
			}
		}
	`)
}

func integrationResponse(t *testing.T) []byte {
	path := filepath.Join(testtools.GetParentDir(t), "testdata", "spotify_artist_search_response.json")
	return testtools.LoadTestDataOrError(t, path)
}

func defaultDeserializedResponse() spotifyResponse {
	return spotifyResponse{
		Artists: spotifyArtists{
			ArtistItems: []spotifyArtist{
				{Name: "Movements"},
				{
					Name: "The Movement",
					Images: []spotifyImage{
						{Url: "https://some.url1"},
						{Url: "https://some.url2"},
					},
				},
			},
		},
	}
}

func defaultArtists() []artist.Artist {
	return []artist.Artist{
		artist.NewArtist("Movements"),
		artist.NewArtistWithImageUri("The Movement", "https://some.url2"),
	}
}

func integrationArtists() []artist.Artist {
	return []artist.Artist{
		artist.NewArtistWithImageUri(
			"The Beatles",
			"https://i.scdn.co/image/ab6761610000f178e9348cc01ff5d55971b22433",
		),
		artist.NewArtistWithImageUri(
			"The Beatles Tribute Band",
			"https://i.scdn.co/image/ab67616d00004851a53d58fac4e46d5264adc122",
		),
		artist.NewArtist("The Beatles Recovered Band"),
		artist.NewArtistWithImageUri(
			"The Beatles Greatest Hits Performed By The Frank Berman Band",
			"https://i.scdn.co/image/ab67616d00004851f903d75acdce7727b3c4aa2c",
		),
		artist.NewArtist("The Beatles Revival Band"),
	}
}

func defaultSender() *httpsender.FakeHTTPSender {
	sender := &httpsender.FakeHTTPSender{}
	response := defaultResponse()
	sender.SetResponse(&response)
	return sender
}

func defaultDeserializer() *serialization.FakeDeserializer[spotifyResponse] {
	deserializer := &serialization.FakeDeserializer[spotifyResponse]{}
	deserializer.SetResponse(defaultDeserializedResponse())
	return deserializer
}

func spotifySongRepository(sender httpsender.HTTPRequestSender) SpotifyArtistRepository {
	repository := NewSpotifyArtistRepository(sender)
	repository.SetDeserializer(defaultDeserializer())
	repository.SetTokenKey(defaultTokenKey())
	return *repository
}

func TestSearchArtistSendsRequestWithProperOptions(t *testing.T) {
	sender := defaultSender()
	repository := spotifySongRepository(sender)

	_, err := repository.SearchArtist(defaultContext(), defaultSearchName(), defaultLimit())

	testtools.AssertErrorIsNil(t, err)
	testtools.AssertEqual(t, sender.GetSendArgs(), expectedHttpOptions())
}

func TestSearchArtistReturnsErrorOnWrongKeyType(t *testing.T) {
	ctx := defaultContext()
	ctx = context.WithValue(ctx, defaultTokenKey(), 42)
	repository := spotifySongRepository(defaultSender())

	_, err := repository.SearchArtist(ctx, defaultSearchName(), defaultLimit())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestSearchArtistReturnsErrorOnSendError(t *testing.T) {
	sender := &httpsender.FakeHTTPSender{}
	sender.SetError(errors.New("test error"))
	repository := spotifySongRepository(sender)

	_, err := repository.SearchArtist(defaultContext(), defaultSearchName(), defaultLimit())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestSearchArtistCallsDeserializeWithSendResponseBody(t *testing.T) {
	repository := spotifySongRepository(defaultSender())
	deserializer := defaultDeserializer()
	repository.SetDeserializer(deserializer)

	_, err := repository.SearchArtist(defaultContext(), defaultSearchName(), defaultLimit())

	testtools.AssertErrorIsNil(t, err)
	testtools.AssertEqual(t, deserializer.GetArgs(), defaultResponse())
}

func TestSearchArtistsReturnsErrorOnResponseBodyDeserializationError(t *testing.T) {
	repository := spotifySongRepository(defaultSender())
	deserializer := defaultDeserializer()
	deserializer.SetError(errors.New("test error"))
	repository.SetDeserializer(deserializer)

	_, err := repository.SearchArtist(defaultContext(), defaultSearchName(), defaultLimit())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestSearchArtistReturnsDeserializedArtists(t *testing.T) {
	repository := spotifySongRepository(defaultSender())

	artists, _ := repository.SearchArtist(defaultContext(), defaultSearchName(), defaultLimit())

	testtools.AssertEqual(t, artists, defaultArtists())
}

func TestSearchArtistReturnsEmptyIfNoneFound(t *testing.T) {
	repository := spotifySongRepository(defaultSender())
	deserializer := defaultDeserializer()
	emptyResponse := spotifyResponse{Artists: spotifyArtists{ArtistItems: []spotifyArtist{}}}
	deserializer.SetResponse(emptyResponse)
	repository.SetDeserializer(deserializer)

	artists, _ := repository.SearchArtist(defaultContext(), defaultSearchName(), defaultLimit())

	testtools.AssertEqual(t, artists, []artist.Artist{})
}

func TestSearchArtistReturnsDeserializedArtistsIntegration(t *testing.T) {
	testtools.SkipOnShortRun(t)

	sender := defaultSender()
	response := integrationResponse(t)
	sender.SetResponse(&response)
	repository := NewSpotifyArtistRepository(sender)
	repository.SetTokenKey(defaultTokenKey())

	artists, _ := repository.SearchArtist(defaultContext(), defaultSearchName(), defaultLimit())

	testtools.AssertEqual(t, artists, integrationArtists())
}
