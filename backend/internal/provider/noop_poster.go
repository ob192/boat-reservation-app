package provider

import "context"

// NoopPosterClient is used when Poster is not configured.
type NoopPosterClient struct{}

func (NoopPosterClient) CreateIncomingOrder(context.Context, PosterOrder) (PosterOrderResult, error) {
	return PosterOrderResult{}, nil
}
