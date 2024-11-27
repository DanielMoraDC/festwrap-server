package errors

type CannotSearchPlaylistError struct {
	message string
}

func NewCannotSearchPlaylistError(message string) error {
	return &CannotSearchPlaylistError{message: message}
}

func (e *CannotSearchPlaylistError) Error() string {
	return e.message
}
