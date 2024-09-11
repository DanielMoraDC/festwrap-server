package spotify

import (
	"errors"
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

func expectedHttpOptions() httpsender.HTTPRequestOptions {
	url := "https://api.spotify.com/v1/search?limit=2&q=artist%3AMovements&type=artist"
	options := httpsender.NewHTTPRequestOptions(url, httpsender.GET, 200)
	options.SetHeaders(
		map[string]string{"Authorization": "Bearer some_token"},
	)
	return options
}

func defaultSenderResponse() *[]byte {
	response := []byte("some body")
	return &response
}

func defaultArtists() []artist.Artist {
	return []artist.Artist{
		artist.NewArtist("Movements"),
		artist.NewArtistWithImageUri("The Movement", "https://some.url"),
	}
}

func defaultSender() *httpsender.FakeHTTPSender {
	sender := &httpsender.FakeHTTPSender{}
	sender.SetResponse(defaultSenderResponse())
	return sender
}

func defaultDeserializer() *serialization.FakeDeserializer[[]artist.Artist] {
	deserializer := &serialization.FakeDeserializer[[]artist.Artist]{}
	response := defaultArtists()
	deserializer.SetResponse(&response)
	return deserializer
}

func spotifySongRepository(sender httpsender.HTTPRequestSender) SpotifyArtistRepository {
	repository := NewSpotifyArtistRepository("some_token", sender)
	repository.SetDeserializer(defaultDeserializer())
	return *repository
}

func TestSearchArtistSendsRequestWithProperOptions(t *testing.T) {
	sender := defaultSender()
	repository := spotifySongRepository(sender)

	_, err := repository.SearchArtist(defaultSearchName(), defaultLimit())

	testtools.AssertErrorIsNil(t, err)
	testtools.AssertEqual(t, sender.GetSendArgs(), expectedHttpOptions())
}

func TestSearchArtistReturnsErrorOnSendError(t *testing.T) {
	sender := &httpsender.FakeHTTPSender{}
	sender.SetError(errors.New("test error"))
	repository := spotifySongRepository(sender)

	_, err := repository.SearchArtist(defaultSearchName(), defaultLimit())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestSearchArtistCallsDeserializeWithSendResponseBody(t *testing.T) {
	repository := spotifySongRepository(defaultSender())
	deserializer := defaultDeserializer()
	repository.SetDeserializer(deserializer)

	_, err := repository.SearchArtist(defaultSearchName(), defaultLimit())

	testtools.AssertErrorIsNil(t, err)
	testtools.AssertEqual(t, deserializer.GetArgs(), *defaultSenderResponse())
}

func TestSearchArtistsReturnsErrorOnResponseBodyDeserializationError(t *testing.T) {
	repository := spotifySongRepository(defaultSender())
	deserializer := defaultDeserializer()
	deserializer.SetError(errors.New("test error"))
	repository.SetDeserializer(deserializer)

	_, err := repository.SearchArtist(defaultSearchName(), defaultLimit())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestSearchArtistReturnsDeserializedArtists(t *testing.T) {
	repository := spotifySongRepository(defaultSender())

	artists, _ := repository.SearchArtist(defaultSearchName(), defaultLimit())

	testtools.AssertEqual(t, *artists, defaultArtists())
}

func TestSearchArtistReturnsEmptyIfNoneFound(t *testing.T) {
	repository := spotifySongRepository(defaultSender())
	deserializer := defaultDeserializer()
	deserializer.SetResponse(&[]artist.Artist{})
	repository.SetDeserializer(deserializer)

	artists, _ := repository.SearchArtist(defaultSearchName(), defaultLimit())

	testtools.AssertEqual(t, *artists, []artist.Artist{})
}