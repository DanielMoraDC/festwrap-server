package spotify

import (
	"context"
	"errors"
	"fmt"
	"testing"

	types "festwrap/internal"
	httpsender "festwrap/internal/http/sender"
	"festwrap/internal/playlist"
	"festwrap/internal/serialization"
	"festwrap/internal/song"
	"festwrap/internal/testtools"
)

func fakeSender() *httpsender.FakeHTTPSender {
	sender := httpsender.FakeHTTPSender{}
	emptyResponse := []byte("")
	sender.SetResponse(&emptyResponse)
	return &sender
}

func errorSender() *httpsender.FakeHTTPSender {
	sender := httpsender.FakeHTTPSender{}
	sender.SetError(errors.New("test send error"))
	return &sender
}

func defaultPlaylistId() string {
	return "test_id"
}

func defaultPlaylistName() string {
	return "myPlaylist"
}

func defaultSearchPlaylistLimit() int {
	return 5
}

func defaultSongs() []song.Song {
	return []song.Song{song.NewSong("uri1"), song.NewSong("uri2")}
}

func defaultPlaylist() playlist.Playlist {
	return playlist.Playlist{Name: "my-playlist", Description: "some playlist", IsPublic: false}
}

func defaultSongsBody() []byte {
	return []byte(`{"uris":["uri1","uri2"]}`)
}

func defaultPlaylistBody() []byte {
	return []byte(`{"name":"my-playlist","description":"some playlist","is_public":false}`)
}

func defaultSearchedPlaylistsResponse() []byte {
	return []byte(`
		{
			"playlists": {
				"items": [
					{
						"name":"first playlist",
						"description":"First description",
						"public":true,
						"owner":{"id":"qrRwLBFxQL9fknW8NzBn4JprRNgS"}
					},
					{
						"name":"second playlist",
						"description":"Second description",
						"public":false,
						"owner":{"id":"another_owner_id"}
					}
				]
			}
		}
	`)
}

func defaultSearchedPlaylists() SpotifySearchPlaylistResponse {
	return SpotifySearchPlaylistResponse{
		Playlists: SpotifySearchPlaylists{
			Items: []SpotifySearchPlaylist{
				{
					Name:          "first playlist",
					Description:   "First description",
					Public:        true,
					OwnerMetadata: SpotifyPlaylistOwnerMetadata{Id: defaultUserId()},
				},
				{
					Name:          "second playlist",
					Description:   "Second description",
					Public:        false,
					OwnerMetadata: SpotifyPlaylistOwnerMetadata{Id: "another_owner_id"},
				},
			},
		},
	}
}

func defaultSearchedFilteredPlaylists() []playlist.Playlist {
	return []playlist.Playlist{
		{
			Name:        "first playlist",
			Description: "First description",
			IsPublic:    true,
		},
	}
}

func defaultSongsSerializer() *serialization.FakeSerializer[SpotifySongs] {
	serializer := serialization.FakeSerializer[SpotifySongs]{}
	serializer.SetResponse(defaultSongsBody())
	return &serializer
}

func defaultPlaylistSerializer() *serialization.FakeSerializer[SpotifyPlaylist] {
	serializer := serialization.FakeSerializer[SpotifyPlaylist]{}
	serializer.SetResponse(defaultPlaylistBody())
	return &serializer
}

func defaultPlaylistDeserializer() *serialization.FakeDeserializer[SpotifySearchPlaylistResponse] {
	deserializer := serialization.FakeDeserializer[SpotifySearchPlaylistResponse]{}
	deserializer.SetResponse(defaultSearchedPlaylists())
	return &deserializer
}

func defaultToken() string {
	return "abcdefg12345"
}

func defaultTokenKey() types.ContextKey {
	return "token"
}

func defaultUserId() string {
	return "qrRwLBFxQL9fknW8NzBn4JprRNgS"
}

func defaultUserIdKey() types.ContextKey {
	return "userId"
}

func defaultContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, defaultTokenKey(), defaultToken())
	ctx = context.WithValue(ctx, defaultUserIdKey(), defaultUserId())
	return ctx
}

func expectedAddSongsHttpOptions() httpsender.HTTPRequestOptions {
	options := httpsender.NewHTTPRequestOptions("https://api.spotify.com/v1/playlists/test_id/tracks", httpsender.POST, 201)
	options.SetHeaders(defaultHeaders())
	options.SetBody(defaultSongsBody())
	return options
}

func expectedCreatePlaylistHttpOptions() httpsender.HTTPRequestOptions {
	url := fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", defaultUserId())
	options := httpsender.NewHTTPRequestOptions(url, httpsender.POST, 201)
	options.SetHeaders(defaultHeaders())
	options.SetBody(defaultPlaylistBody())
	return options
}

func expectedSearchPlaylistHttpOptions() httpsender.HTTPRequestOptions {
	url := fmt.Sprintf(
		"https://api.spotify.com/v1/search?limit=%d&q=%s&type=playlist",
		defaultSearchPlaylistLimit(),
		defaultPlaylistName(),
	)
	options := httpsender.NewHTTPRequestOptions(url, httpsender.POST, 201)
	options.SetHeaders(defaultHeaders())
	return options
}

func defaultHeaders() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", defaultToken()),
		"Content-Type":  "application/json",
	}
}

func spotifyPlaylistRepository() SpotifyPlaylistRepository {
	repository := NewSpotifyPlaylistRepository(fakeSender())

	repository.SetTokenKey(defaultTokenKey())
	repository.SetUserIdKey(defaultUserIdKey())

	repository.SetSongSerializer(defaultSongsSerializer())
	repository.SetPlaylistSerializer(defaultPlaylistSerializer())
	repository.SetPlaylistDeserializer(defaultPlaylistDeserializer())

	return repository
}

func TestAddSongsReturnsErrorWhenNoSongsProvided(t *testing.T) {
	repository := spotifyPlaylistRepository()

	err := repository.AddSongs(defaultContext(), defaultPlaylistId(), []song.Song{})

	testtools.AssertErrorIsNotNil(t, err)
}

func TestAddSongsSerializesInputSongs(t *testing.T) {
	repository := spotifyPlaylistRepository()

	repository.AddSongs(defaultContext(), defaultPlaylistId(), defaultSongs())

	expected := SpotifySongs{Uris: []string{"uri1", "uri2"}}
	actual := repository.GetSongSerializer().(*serialization.FakeSerializer[SpotifySongs]).GetArgs()
	testtools.AssertEqual(t, actual, expected)
}

func TestAddSongsReturnsErrorOnNonSerializationError(t *testing.T) {
	repository := spotifyPlaylistRepository()
	serializer := defaultSongsSerializer()
	serializer.SetError(errors.New("test songs error"))
	repository.SetSongSerializer(serializer)

	err := repository.AddSongs(defaultContext(), defaultPlaylistId(), defaultSongs())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestAddSongsSendsRequestUsingProperOptions(t *testing.T) {
	repository := spotifyPlaylistRepository()

	repository.AddSongs(defaultContext(), defaultPlaylistId(), defaultSongs())

	actual := repository.GetHTTPSender().(*httpsender.FakeHTTPSender).GetSendArgs()
	testtools.AssertEqual(t, actual, expectedAddSongsHttpOptions())
}

func TestAddSongsReturnsErrorOnSendError(t *testing.T) {
	repository := spotifyPlaylistRepository()
	repository.SetHTTPSender(errorSender())

	err := repository.AddSongs(defaultContext(), defaultPlaylistId(), defaultSongs())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestAddSongsSerializesInputPlaylist(t *testing.T) {
	repository := spotifyPlaylistRepository()

	repository.CreatePlaylist(defaultContext(), defaultUserId(), defaultPlaylist())

	expected := SpotifyPlaylist{Name: "my-playlist", Description: "some playlist", IsPublic: false}
	actual := repository.GetPlaylistSerializer().(*serialization.FakeSerializer[SpotifyPlaylist]).GetArgs()
	testtools.AssertEqual(t, actual, expected)
}

func TestCreatePlaylistReturnsErrorOnPlaylistSerializationError(t *testing.T) {
	repository := spotifyPlaylistRepository()
	serializer := defaultPlaylistSerializer()
	serializer.SetError(errors.New("test playlist error"))
	repository.SetPlaylistSerializer(serializer)

	err := repository.CreatePlaylist(defaultContext(), defaultUserId(), defaultPlaylist())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestCreatePlaylistSendsCreateRequestWithOptions(t *testing.T) {
	repository := spotifyPlaylistRepository()

	repository.CreatePlaylist(defaultContext(), defaultUserId(), defaultPlaylist())

	actual := repository.GetHTTPSender().(*httpsender.FakeHTTPSender).GetSendArgs()
	testtools.AssertEqual(t, actual, expectedCreatePlaylistHttpOptions())
}

func TestCreatePlaylistReturnsErrorOnSendError(t *testing.T) {
	repository := spotifyPlaylistRepository()
	repository.SetHTTPSender(errorSender())

	err := repository.CreatePlaylist(defaultContext(), defaultUserId(), defaultPlaylist())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestSearchPlaylistSendsCreateRequestWithOptions(t *testing.T) {
	repository := spotifyPlaylistRepository()

	repository.SearchPlaylist(defaultContext(), defaultPlaylistName(), defaultSearchPlaylistLimit())

	actual := repository.GetHTTPSender().(*httpsender.FakeHTTPSender).GetSendArgs()
	testtools.AssertEqual(t, actual, expectedSearchPlaylistHttpOptions())
}

func TestSearchPlaylistReturnsErrorOnSendError(t *testing.T) {
	repository := spotifyPlaylistRepository()
	repository.SetHTTPSender(errorSender())

	_, err := repository.SearchPlaylist(defaultContext(), defaultPlaylistName(), defaultSearchPlaylistLimit())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestSearchPlaylistReturnsErrorOnDeserializationError(t *testing.T) {
	deserializer := defaultPlaylistDeserializer()
	deserializer.SetError(errors.New("test deserialization error"))
	repository := spotifyPlaylistRepository()
	repository.SetPlaylistDeserializer(deserializer)

	_, err := repository.SearchPlaylist(defaultContext(), defaultPlaylistName(), defaultSearchPlaylistLimit())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestSearchPlaylistReturnsDeserializedContent(t *testing.T) {
	repository := spotifyPlaylistRepository()

	actual, err := repository.SearchPlaylist(defaultContext(), defaultPlaylistName(), defaultSearchPlaylistLimit())

	testtools.AssertErrorIsNil(t, err)
	testtools.AssertEqual(t, actual, defaultSearchedFilteredPlaylists())
}

func TestAddSongsPlaylistSendsOptionsUsingSerializerIntegration(t *testing.T) {
	testtools.SkipOnShortRun(t)

	serializer := serialization.NewJsonSerializer[SpotifySongs]()
	repository := spotifyPlaylistRepository()
	repository.SetSongSerializer(&serializer)

	repository.AddSongs(defaultContext(), defaultPlaylistId(), defaultSongs())

	actual := repository.GetHTTPSender().(*httpsender.FakeHTTPSender).GetSendArgs()
	testtools.AssertEqual(t, actual, expectedAddSongsHttpOptions())
}

func TestCreatePlaylistSendsOptionsUsingSerializerIntegration(t *testing.T) {
	testtools.SkipOnShortRun(t)

	serializer := serialization.NewJsonSerializer[SpotifyPlaylist]()
	repository := spotifyPlaylistRepository()
	repository.SetPlaylistSerializer(&serializer)

	repository.CreatePlaylist(defaultContext(), defaultUserId(), defaultPlaylist())

	actual := repository.GetHTTPSender().(*httpsender.FakeHTTPSender).GetSendArgs()
	testtools.AssertEqual(t, actual, expectedCreatePlaylistHttpOptions())
}

func TestSearchPlaylistReturnsResultsUsingDeserializerIntegration(t *testing.T) {
	testtools.SkipOnShortRun(t)

	sender := fakeSender()
	response := defaultSearchedPlaylistsResponse()
	sender.SetResponse(&response)
	deserializer := serialization.NewJsonDeserializer[SpotifySearchPlaylistResponse]()
	repository := spotifyPlaylistRepository()
	repository.SetPlaylistDeserializer(deserializer)
	repository.SetHTTPSender(sender)

	actual, err := repository.SearchPlaylist(defaultContext(), defaultPlaylistName(), defaultSearchPlaylistLimit())

	testtools.AssertEqual(t, actual, defaultSearchedFilteredPlaylists())
	testtools.AssertErrorIsNil(t, err)
}

func TestRepositoryMethodsReturnErrorWhenInvalidToken(t *testing.T) {
	tests := map[string]struct {
		repositoryTokenKey types.ContextKey
		tokenKey           types.ContextKey
		tokenVal           interface{}
	}{
		"returns error when token is wrong type": {
			repositoryTokenKey: "matchingKey",
			tokenKey:           "matchingKey",
			tokenVal:           1234,
		},
		"returns error when token is missing": {
			repositoryTokenKey: "someKey",
			tokenKey:           "otherKey",
			tokenVal:           "myToken",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ctx = context.WithValue(ctx, test.tokenKey, test.tokenVal)
			repository := spotifyPlaylistRepository()
			repository.SetTokenKey(test.repositoryTokenKey)

			err := repository.AddSongs(ctx, defaultPlaylistId(), defaultSongs())
			testtools.AssertErrorIsNotNil(t, err)

			err = repository.CreatePlaylist(ctx, defaultUserId(), defaultPlaylist())
			testtools.AssertErrorIsNotNil(t, err)

			_, err = repository.SearchPlaylist(ctx, defaultPlaylistName(), defaultSearchPlaylistLimit())
			testtools.AssertErrorIsNotNil(t, err)
		})
	}
}

func TestRepositoryMethodsReturnErrorWhenInvalidUserId(t *testing.T) {
	tests := map[string]struct {
		repositoryUserIdKey types.ContextKey
		userIdKey           types.ContextKey
		userIdVal           interface{}
	}{
		"returns error when userId is wrong type": {
			repositoryUserIdKey: "matchingKey",
			userIdKey:           "matchingKey",
			userIdVal:           1234,
		},
		"returns error when userId is missing": {
			repositoryUserIdKey: "someKey",
			userIdKey:           "otherKey",
			userIdVal:           "myUser",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ctx = context.WithValue(ctx, test.userIdKey, test.userIdVal)
			repository := spotifyPlaylistRepository()
			repository.SetUserIdKey(test.repositoryUserIdKey)

			_, err := repository.SearchPlaylist(ctx, defaultPlaylistName(), defaultSearchPlaylistLimit())
			testtools.AssertErrorIsNotNil(t, err)
		})
	}
}
