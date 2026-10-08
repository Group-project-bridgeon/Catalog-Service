package service

const (
	DefaultPageLimit = 50
	MaxPageLimit     = 100
)

type PageInput struct {
	Limit  int
	Cursor string
}

type Page[T any] struct {
	Items      []T
	NextCursor *string
}

func normalizePageInput(
	input PageInput,
) (PageInput, error) {
	if input.Limit == 0 {
		input.Limit = DefaultPageLimit
	}

	if input.Limit < 1 ||
		input.Limit > MaxPageLimit {
		return PageInput{}, ErrInvalidArgument
	}

	return input, nil
}

func trimPage[T any](
	items []T,
	limit int,
) ([]T, bool) {
	if len(items) <= limit {
		return items, false
	}

	return items[:limit], true
}
