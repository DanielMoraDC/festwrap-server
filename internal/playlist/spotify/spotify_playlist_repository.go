package spotify

import (
	"context"
	"fmt"
	"net/url"

	types "festwrap/internal"
	httpsender "festwrap/internal/http/sender"
	"festwrap/internal/playlist"
	"festwrap/internal/playlist/errors"
	"festwrap/internal/serialization"
	"festwrap/internal/song"
)

type SpotifyPlaylistRepository struct {
	songsSerializer      serialization.Serializer[SpotifySongs]
	playlistSerializer   serialization.Serializer[SpotifyPlaylist]
	playlistDeserializer serialization.Deserializer[SpotifySearchPlaylistResponse]
	userIdKey            types.ContextKey
	tokenKey             types.ContextKey
	host                 string
	httpSender           httpsender.HTTPRequestSender
}

func NewSpotifyPlaylistRepository(httpSender httpsender.HTTPRequestSender) SpotifyPlaylistRepository {
	playlistSerializer := serialization.NewJsonSerializer[SpotifyPlaylist]()
	songSerializer := serialization.NewJsonSerializer[SpotifySongs]()
	playlistDeserializer := serialization.NewJsonDeserializer[SpotifySearchPlaylistResponse]()
	return SpotifyPlaylistRepository{
		tokenKey:             "token",
		userIdKey:            "user_id",
		host:                 "api.spotify.com",
		httpSender:           httpSender,
		songsSerializer:      &songSerializer,
		playlistSerializer:   &playlistSerializer,
		playlistDeserializer: &playlistDeserializer,
	}
}

func (r *SpotifyPlaylistRepository) AddSongs(ctx context.Context, playlistId string, songs []song.Song) error {
	if len(songs) == 0 {
		return errors.NewCannotAddSongsToPlaylistError("no songs provided")
	}

	token, ok := ctx.Value(r.tokenKey).(string)
	if !ok {
		return errors.NewCannotAddSongsToPlaylistError("Could not retrieve token from context")
	}

	body, err := r.songsSerializer.Serialize(NewSpotifySongs(songs))
	if err != nil {
		errorMsg := fmt.Sprintf("could not serialize songs: %v", err.Error())
		return errors.NewCannotAddSongsToPlaylistError(errorMsg)
	}

	httpOptions := r.addSongsHttpOptions(playlistId, body, token)
	_, err = r.httpSender.Send(httpOptions)
	if err != nil {
		return errors.NewCannotAddSongsToPlaylistError(err.Error())
	}

	return nil
}

func (r *SpotifyPlaylistRepository) CreatePlaylist(ctx context.Context, userId string, playlist playlist.Playlist) error {
	token, ok := ctx.Value(r.tokenKey).(string)
	if !ok {
		return errors.NewCannotCreatePlaylistError("Could not retrieve token from context")
	}

	body, err := r.playlistSerializer.Serialize(
		SpotifyPlaylist{
			Name:        playlist.Name,
			Description: playlist.Description,
			IsPublic:    playlist.IsPublic,
		},
	)
	if err != nil {
		errorMsg := fmt.Sprintf("could not serialize playlist: %v", err.Error())
		return errors.NewCannotCreatePlaylistError(errorMsg)
	}

	httpOptions := r.createPlaylistOptions(userId, body, token)
	_, err = r.httpSender.Send(httpOptions)
	if err != nil {
		return errors.NewCannotCreatePlaylistError(err.Error())
	}

	return nil
}

func (r *SpotifyPlaylistRepository) SearchPlaylist(ctx context.Context, name string, limit int) ([]playlist.Playlist, error) {
	emptyResponse := []playlist.Playlist{}
	token, ok := ctx.Value(r.tokenKey).(string)
	if !ok {
		return emptyResponse, errors.NewCannotSearchPlaylistError("Could not retrieve token from context")
	}

	userId, ok := ctx.Value(r.userIdKey).(string)
	if !ok {
		return emptyResponse, errors.NewCannotSearchPlaylistError("Could not retrieve user id from context")
	}

	httpOptions := r.searchPlaylistOptions(name, limit, token)
	response, err := r.httpSender.Send(httpOptions)
	if err != nil {
		return emptyResponse, errors.NewCannotSearchPlaylistError(err.Error())
	}

	var searchedPlaylist SpotifySearchPlaylistResponse
	err = r.playlistDeserializer.Deserialize(*response, &searchedPlaylist)
	if err != nil {
		return nil, errors.NewCannotSearchPlaylistError(err.Error())
	}

	return filterPlaylistByUser(searchedPlaylist.Playlists.Items, userId), nil
}

func filterPlaylistByUser(playlists []SpotifySearchPlaylist, userId string) []playlist.Playlist {
	var userPlaylists []playlist.Playlist
	for _, currentPlaylist := range playlists {
		if currentPlaylist.OwnerMetadata.Id == userId {
			playlistObj := playlist.Playlist{
				Name:        currentPlaylist.Name,
				Description: currentPlaylist.Description,
				IsPublic:    currentPlaylist.Public,
			}
			userPlaylists = append(userPlaylists, playlistObj)
		}
	}
	return userPlaylists
}

func (r *SpotifyPlaylistRepository) SetUserIdKey(key types.ContextKey) {
	r.userIdKey = key
}

func (r *SpotifyPlaylistRepository) SetTokenKey(key types.ContextKey) {
	r.tokenKey = key
}

func (r *SpotifyPlaylistRepository) GetHTTPSender() httpsender.HTTPRequestSender {
	return r.httpSender
}

func (r *SpotifyPlaylistRepository) SetHTTPSender(httpSender httpsender.HTTPRequestSender) {
	r.httpSender = httpSender
}

func (r *SpotifyPlaylistRepository) SetSongSerializer(serializer serialization.Serializer[SpotifySongs]) {
	r.songsSerializer = serializer
}

func (r *SpotifyPlaylistRepository) GetSongSerializer() serialization.Serializer[SpotifySongs] {
	return r.songsSerializer
}

func (r *SpotifyPlaylistRepository) GetPlaylistSerializer() serialization.Serializer[SpotifyPlaylist] {
	return r.playlistSerializer
}

func (r *SpotifyPlaylistRepository) SetPlaylistSerializer(serializer serialization.Serializer[SpotifyPlaylist]) {
	r.playlistSerializer = serializer
}

func (r *SpotifyPlaylistRepository) SetPlaylistDeserializer(deserializer serialization.Deserializer[SpotifySearchPlaylistResponse]) {
	r.playlistDeserializer = deserializer
}

func (r *SpotifyPlaylistRepository) addSongsHttpOptions(
	playlistId string, body []byte, token string,
) httpsender.HTTPRequestOptions {
	url := fmt.Sprintf("https://%s/v1/playlists/%s/tracks", r.host, playlistId)
	httpOptions := httpsender.NewHTTPRequestOptions(url, httpsender.POST, 201)
	httpOptions.SetBody(body)
	httpOptions.SetHeaders(r.GetSpotifyBaseHeaders(token))
	return httpOptions
}

func (r *SpotifyPlaylistRepository) createPlaylistOptions(
	userId string, body []byte, token string,
) httpsender.HTTPRequestOptions {
	url := fmt.Sprintf("https://%s/v1/users/%s/playlists", r.host, userId)
	httpOptions := httpsender.NewHTTPRequestOptions(url, httpsender.POST, 201)
	httpOptions.SetBody(body)
	httpOptions.SetHeaders(r.GetSpotifyBaseHeaders(token))
	return httpOptions
}

func (r *SpotifyPlaylistRepository) searchPlaylistOptions(
	playlistName string, limit int, token string,
) httpsender.HTTPRequestOptions {
	queryParams := url.Values{}
	queryParams.Set("q", playlistName)
	queryParams.Set("type", "playlist")
	queryParams.Set("limit", fmt.Sprintf("%d", limit))
	url := fmt.Sprintf("https://%s/v1/search?%s", r.host, queryParams.Encode())
	httpOptions := httpsender.NewHTTPRequestOptions(url, httpsender.POST, 201)
	httpOptions.SetHeaders(r.GetSpotifyBaseHeaders(token))
	return httpOptions
}

func (r *SpotifyPlaylistRepository) GetSpotifyBaseHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Content-Type":  "application/json",
	}
}
