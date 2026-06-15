package embeddings

import "context"

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

type NoopEmbedder struct{}

func (NoopEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	return nil, nil
}
