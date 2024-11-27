package setlistfm

import (
	"errors"
	httpsender "festwrap/internal/http/sender"
	"festwrap/internal/serialization"
	"festwrap/internal/setlist"
	"festwrap/internal/testtools"
	"path/filepath"
	"testing"
)

func defaultArtist() string {
	return "Boysetsfire"
}

func deserializedResponse() setlistFMResponse {
	artist := setlistfmArtist{Name: defaultArtist()}
	songs := []setlistfmSong{
		{Name: "Closure"},
		{Name: "Rookie"},
		{Name: "The Misery Index"},
		{Name: "One Match"},
		{Name: "Requiem"},
	}
	set := setlistfmSet{Songs: songs}
	setlists := []setlistFMSetlist{
		{Artist: artist, Sets: setlistFMSets{Sets: []setlistfmSet{set}}},
	}
	return setlistFMResponse{Body: setlists}
}

func defaultMinSongs() int {
	return 3
}

func senderResponse() []byte {
	return []byte("some response")
}

func defaultSender() *httpsender.FakeHTTPSender {
	sender := httpsender.FakeHTTPSender{}
	response := senderResponse()
	sender.SetResponse(&response)
	return &sender
}

func defaultDeserializer() serialization.FakeDeserializer[setlistFMResponse] {
	result := serialization.FakeDeserializer[setlistFMResponse]{}
	result.SetResponse(deserializedResponse())
	return result
}

func expectedHttpOptions() httpsender.HTTPRequestOptions {
	url := "https://api.setlist.fm/rest/1.0/search/setlists?artistName=Boysetsfire"
	options := httpsender.NewHTTPRequestOptions(url, httpsender.GET, 200)
	options.SetHeaders(
		map[string]string{
			"x-api-key": "some_api_key",
			"Accept":    "application/json",
		},
	)
	return options
}

func expectedSetlist() *setlist.Setlist {
	expected := setlist.NewSetlist(
		"Boysetsfire",
		[]setlist.Song{
			setlist.NewSong("Closure"),
			setlist.NewSong("Rookie"),
			setlist.NewSong("The Misery Index"),
			setlist.NewSong("One Match"),
			setlist.NewSong("Requiem"),
		},
	)
	return &expected
}

func setlistRepository(sender httpsender.HTTPRequestSender) SetlistFMRepository {
	repository := *NewSetlistFMSetlistRepository("some_api_key", sender)
	deserializer := defaultDeserializer()
	repository.SetDeserializer(&deserializer)
	return repository
}

func integrationResponse(t *testing.T) []byte {
	path := filepath.Join(testtools.GetParentDir(t), "testdata", "response.json")
	return testtools.LoadTestDataOrError(t, path)
}

func integrationSetlist() *setlist.Setlist {
	songs := []setlist.Song{
		setlist.NewSong("Walk of Life"),
		setlist.NewSong("Anna"),
		setlist.NewSong("Nice Things"),
		setlist.NewSong("America (You're Freaking Me Out)"),
		setlist.NewSong("The Obituaries"),
		setlist.NewSong("After the Party"),
		setlist.NewSong("Irish Goodbyes"),
		setlist.NewSong("Casey"),
		setlist.NewSong("Layla"),
	}
	setlist := setlist.NewSetlist("The Menzingers", songs)
	return &setlist
}

func TestGetSetlistSenderCalledWithProperOptions(t *testing.T) {
	sender := defaultSender()
	repository := setlistRepository(sender)

	repository.GetSetlist(defaultArtist(), defaultMinSongs())

	testtools.AssertEqual(t, sender.GetSendArgs(), expectedHttpOptions())
}

func TestGetSetlistReturnsErrorOnSenderError(t *testing.T) {
	sender := httpsender.FakeHTTPSender{}
	sender.SetError(errors.New("test error"))
	repository := setlistRepository(&sender)

	_, err := repository.GetSetlist(defaultArtist(), defaultMinSongs())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestGetSetlistDeserializerCalledWithSenderResponse(t *testing.T) {
	repository := setlistRepository(defaultSender())
	deserializer := defaultDeserializer()
	repository.SetDeserializer(&deserializer)

	repository.GetSetlist(defaultArtist(), defaultMinSongs())

	testtools.AssertEqual(t, deserializer.GetArgs(), senderResponse())
}

func TestGetSetlistReturnsErrorOnDeserializationError(t *testing.T) {
	repository := setlistRepository(defaultSender())
	deserializer := defaultDeserializer()
	deserializer.SetError(errors.New("test deserialization error"))
	repository.SetDeserializer(&deserializer)

	_, err := repository.GetSetlist(defaultArtist(), defaultMinSongs())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestGetSetlistReturnsErrorIfNoSetlistFound(t *testing.T) {
	repository := setlistRepository(defaultSender())
	deserializer := defaultDeserializer()
	emptyResponse := setlistFMResponse{Body: []setlistFMSetlist{}}
	deserializer.SetResponse(emptyResponse)
	repository.SetDeserializer(&deserializer)

	_, err := repository.GetSetlist(defaultArtist(), defaultMinSongs())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestGetSetlistReturnsSetlistFromDeserializedResponse(t *testing.T) {
	repository := setlistRepository(defaultSender())

	actual, _ := repository.GetSetlist(defaultArtist(), defaultMinSongs())

	testtools.AssertEqual(t, actual, expectedSetlist())
}

func TestGetSetlistRetrievesErrorWhenMinSongsNotReached(t *testing.T) {
	repository := setlistRepository(defaultSender())

	_, err := repository.GetSetlist(defaultArtist(), 50)

	testtools.AssertErrorIsNotNil(t, err)
}

func TestGetSetlistRetrievesSetlistFromResponseIntegration(t *testing.T) {
	testtools.SkipOnShortRun(t)

	sender := defaultSender()
	response := integrationResponse(t)
	sender.SetResponse(&response)
	repository := NewSetlistFMSetlistRepository("some_api_key", sender)

	actual, err := repository.GetSetlist(defaultArtist(), defaultMinSongs())

	testtools.AssertErrorIsNil(t, err)
	testtools.AssertEqual(t, actual, integrationSetlist())
}
