package ai

import "context"

type Suggester interface {
	Suggest(ctx context.Context, input string) (string, error)
}
