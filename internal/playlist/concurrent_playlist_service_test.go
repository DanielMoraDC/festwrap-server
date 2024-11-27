package playlist

import (
	"context"
	"errors"
	"testing"

	"festwrap/internal/setlist"
	"festwrap/internal/song"
	"festwrap/internal/testtools"
)

func defaultContext() context.Context {
	return context.Background()
}

func defaultPlaylistId() string {
	return "myPlaylist"
}

func defaultArtist() string {
	return "myArtist"
}

func defaultSongs() []interface{} {
	return []interface{}{
		song.NewSong("some_uri"),
		song.NewSong("another_uri"),
	}
}

func songsWithErrors() []interface{} {
	return []interface{}{
		errors.New("Some error"),
		song.NewSong("another_uri"),
	}
}

func errorSongs() []interface{} {
	return []interface{}{
		errors.New("Some error"),
		errors.New("Some other error"),
	}
}

func defaultSetlist() setlist.Setlist {
	songs := []setlist.Song{
		setlist.NewSong("My song"),
		setlist.NewSong("My other song"),
	}
	return setlist.NewSetlist(defaultArtist(), songs)
}

func emptySetlist() setlist.Setlist {
	return setlist.NewSetlist(defaultArtist(), []setlist.Song{})
}

func defaultGetSongArgs() []song.GetSongArgs {
	return []song.GetSongArgs{
		{Artist: defaultArtist(), Title: "My song"},
		{Artist: defaultArtist(), Title: "My other song"},
	}
}

func defaultAddSongsArgs() AddSongsArgs {
	return AddSongsArgs{
		Context:    defaultContext(),
		PlaylistId: defaultPlaylistId(),
		Songs: []song.Song{
			song.NewSong("some_uri"),
			song.NewSong("another_uri"),
		},
	}
}

func addSongsArgsWithErrors() AddSongsArgs {
	return AddSongsArgs{
		Context:    defaultContext(),
		PlaylistId: defaultPlaylistId(),
		Songs: []song.Song{
			song.NewSong("another_uri"),
		},
	}
}

func newFakeSetlistRepository() setlist.FakeSetlistRepository {
	repository := setlist.NewFakeSetlistRepository()
	repository.SetReturnValue(defaultSetlist())
	return repository
}

func newFakeSongRepository() song.FakeSongRepository {
	repository := song.NewFakeSongRepository()
	repository.SetSongs(defaultSongs())
	return repository
}

func testSetup() (FakePlaylistRepository, setlist.FakeSetlistRepository, song.FakeSongRepository) {
	playlistRepository := NewFakePlaylistRepository()
	setlistRepository := newFakeSetlistRepository()
	songRepository := newFakeSongRepository()
	return playlistRepository, setlistRepository, songRepository
}

func TestAddSetlistSetlistRepositoryCalledWithArgs(t *testing.T) {
	artist := defaultArtist()
	minSongs := 6
	playlistRepository, setlistRepository, songRepository := testSetup()

	service := NewConcurrentPlaylistService(&playlistRepository, &setlistRepository, &songRepository)
	service.SetMinSongs(minSongs)

	err := service.AddSetlist(defaultContext(), defaultPlaylistId(), artist)

	actual := setlistRepository.GetGetSetlistArgs()
	expected := setlist.GetSetlistArgs{Artist: artist, MinSongs: minSongs}
	testtools.AssertErrorIsNil(t, err)
	testtools.AssertEqual(t, actual, expected)
}

func TestAddSetlistReturnsErrorOnSetlistRepositoryError(t *testing.T) {
	playlistRepository, setlistRepository, songRepository := testSetup()
	returnError := errors.New("test error")
	setlistRepository.SetError(returnError)
	service := NewConcurrentPlaylistService(&playlistRepository, &setlistRepository, &songRepository)

	err := service.AddSetlist(defaultContext(), defaultPlaylistId(), defaultArtist())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestAddSetlistSongRepositoryCalledWithSetlistSongs(t *testing.T) {
	playlistRepository, setlistRepository, songRepository := testSetup()
	service := NewConcurrentPlaylistService(&playlistRepository, &setlistRepository, &songRepository)

	err := service.AddSetlist(defaultContext(), defaultPlaylistId(), defaultArtist())

	actual := songRepository.GetGetSongArgs()
	expected := defaultGetSongArgs()
	testtools.AssertErrorIsNil(t, err)
	if !testtools.HaveSameElements(actual, expected) {
		t.Errorf("Expected called songs %v, found %v", expected, actual)
	}
}

func TestAddSetlistAddsSongsFetched(t *testing.T) {
	playlistRepository, setlistRepository, songRepository := testSetup()
	songRepository.SetSongs(defaultSongs())
	service := NewConcurrentPlaylistService(&playlistRepository, &setlistRepository, &songRepository)

	err := service.AddSetlist(defaultContext(), defaultPlaylistId(), defaultArtist())

	actual := playlistRepository.GetAddSongArgs()
	expected := defaultAddSongsArgs()
	testtools.AssertErrorIsNil(t, err)
	testtools.AssertEqual(t, actual, expected)
}

func TestAddSetlistAddsOnlySongsFetchedWithoutError(t *testing.T) {
	playlistRepository, setlistRepository, songRepository := testSetup()
	songRepository.SetSongs(songsWithErrors())
	service := NewConcurrentPlaylistService(&playlistRepository, &setlistRepository, &songRepository)

	err := service.AddSetlist(defaultContext(), "myPlaylist", defaultArtist())

	actual := playlistRepository.GetAddSongArgs()
	expected := addSongsArgsWithErrors()
	testtools.AssertErrorIsNil(t, err)
	testtools.AssertEqual(t, actual, expected)
}

func TestAddSetlistSetlistRaisesErrorIfSetlistEmpty(t *testing.T) {
	playlistRepository, setlistRepository, songRepository := testSetup()
	songRepository.SetSongs(errorSongs())
	service := NewConcurrentPlaylistService(&playlistRepository, &setlistRepository, &songRepository)

	err := service.AddSetlist(defaultContext(), defaultPlaylistId(), defaultArtist())

	testtools.AssertErrorIsNotNil(t, err)
}

func TestAddSetlistSetlistRaisesErrorIfNoSongsFound(t *testing.T) {
	playlistRepository, setlistRepository, songRepository := testSetup()
	setlistRepository.SetReturnValue(emptySetlist())
	service := NewConcurrentPlaylistService(&playlistRepository, &setlistRepository, &songRepository)

	err := service.AddSetlist(defaultContext(), defaultPlaylistId(), defaultArtist())

	testtools.AssertErrorIsNotNil(t, err)
}
