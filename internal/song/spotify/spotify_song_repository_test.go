package spotify

import (
	"errors"
	httpsender "festwrap/internal/http/sender"
	"festwrap/internal/serialization"
	"festwrap/internal/song"
	"festwrap/internal/testtools"
	"path/filepath"
	"testing"
)

func defaultArtist() string {
	return "Movements"
}

func defaultTitle() string {
	return "Daylily"
}

func expectedHttpOptions() httpsender.HTTPRequestOptions {
	url := "https://api.spotify.com/v1/search?q=%2Bartist%3AMovements%2Btrack%3ADaylily&type=track"
	options := httpsender.NewHTTPRequestOptions(url, httpsender.GET, 200)
	options.SetHeaders(
		map[string]string{"Authorization": "Bearer some_token"},
	)
	return options
}

func defaultResponse() []byte {
	return []byte(`{"tracks":{"items":[{"uri":"some uri"},{}"uri":"another uri"]}}`)
}

func integrationResponse(t *testing.T) []byte {
	return testtools.LoadTestDataOrError(
		t,
		filepath.Join(testtools.GetParentDir(t), "testdata", "search_song_response.json"),
	)
}

func defaultDeserializedResponse() spotifyResponse {
	return spotifyResponse{
		Tracks: spotifyTracks{
			Songs: []spotifySong{
				{"some uri"},
				{"another uri"},
			},
		},
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

func spotifySongRepository(sender httpsender.HTTPRequestSender) SpotifySongRepository {
	repository := NewSpotifySongRepository("some_token", sender)
	repository.SetDeserializer(defaultDeserializer())
	return *repository
}

func TestGetSongSendsRequestWithProperOptions(t *testing.T) {
	sender := defaultSender()
	repository := spotifySongRepository(sender)

	_, err := repository.GetSong(defaultArtist(), defaultTitle())

	testtools.AssertErrorIsNil(t, err)
	testtools.AssertEqual(t, sender.GetSendArgs(), expectedHttpOptions())
}

func TestGetSongReturnsErrorOnSendError(t *testing.T) {
	sender := &httpsender.FakeHTTPSender{}
	sender.SetError(errors.New("test error"))
	repository := spotifySongRepository(sender)

	_, err := repository.GetSong(defaultArtist(), defaultTitle())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestGetSongCallsDeserializeWithSendResponseBody(t *testing.T) {
	repository := spotifySongRepository(defaultSender())
	deserializer := defaultDeserializer()
	repository.SetDeserializer(deserializer)

	_, err := repository.GetSong(defaultArtist(), defaultTitle())

	testtools.AssertErrorIsNil(t, err)
	testtools.AssertEqual(t, deserializer.GetArgs(), defaultResponse())
}

func TestGetSongReturnsErrorOnResponseBodyDeserializationError(t *testing.T) {
	repository := spotifySongRepository(defaultSender())
	deserializer := defaultDeserializer()
	deserializer.SetError(errors.New("test error"))
	repository.SetDeserializer(deserializer)

	_, err := repository.GetSong(defaultArtist(), defaultTitle())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestGetSongReturnsErrorIfNoSongsFound(t *testing.T) {
	repository := spotifySongRepository(defaultSender())
	deserializer := defaultDeserializer()
	emptyResponse := spotifyResponse{Tracks: spotifyTracks{Songs: []spotifySong{}}}
	deserializer.SetResponse(emptyResponse)
	repository.SetDeserializer(deserializer)

	_, err := repository.GetSong(defaultArtist(), defaultTitle())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestGetSongReturnsFirstSongFound(t *testing.T) {
	repository := spotifySongRepository(defaultSender())

	actual, err := repository.GetSong(defaultArtist(), defaultTitle())

	expected := song.NewSong(defaultDeserializedResponse().Tracks.Songs[0].Uri)
	testtools.AssertErrorIsNil(t, err)
	testtools.AssertEqual(t, *actual, expected)
}

func TestGetSongReturnsFirstSongFoundIntegration(t *testing.T) {
	testtools.SkipOnShortRun(t)

	sender := defaultSender()
	response := integrationResponse(t)
	sender.SetResponse(&response)
	repository := NewSpotifySongRepository("some_token", sender)

	actual, err := repository.GetSong(defaultArtist(), defaultTitle())

	expected := song.NewSong("spotify:track:4rH1kFLYW0b28UNRyn7dK3")
	testtools.AssertErrorIsNil(t, err)
	testtools.AssertEqual(t, *actual, expected)
}
