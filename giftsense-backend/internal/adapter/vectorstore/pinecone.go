package vectorstore

import (
	"context"
	"fmt"

	sdk "github.com/pinecone-io/go-pinecone/pinecone"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

type PineconeStore struct {
	client      *sdk.Client
	indexName   string
	environment string
	dimensions  int
}

func NewPineconeStore(apiKey, indexName, environment string, dimensions int) (*PineconeStore, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Pinecone API key is required")
	}
	if indexName == "" {
		return nil, fmt.Errorf("Pinecone index name is required")
	}
	if environment == "" {
		return nil, fmt.Errorf("Pinecone environment is required")
	}

	client, err := sdk.NewClient(sdk.NewClientParams{ApiKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pinecone client: %w", err)
	}

	return &PineconeStore{
		client:      client,
		indexName:   indexName,
		environment: environment,
		dimensions:  dimensions,
	}, nil
}

func (p *PineconeStore) Upsert(ctx context.Context, sessionID string, chunks []domain.Chunk, vectors [][]float32) error {
	idxConn, err := p.indexConnection(ctx, sessionID)
	if err != nil {
		return err
	}
	defer idxConn.Close()

	records := make([]*sdk.Vector, len(chunks))
	for i, chunk := range chunks {
		meta, err := buildMetadata(chunk, i)
		if err != nil {
			return fmt.Errorf("build metadata for chunk %s: %w", chunk.ID, err)
		}
		records[i] = &sdk.Vector{
			Id:       chunk.ID,
			Values:   vectors[i],
			Metadata: meta,
		}
	}

	_, err = idxConn.UpsertVectors(ctx, records)
	if err != nil {
		return fmt.Errorf("upsert vectors: %w", err)
	}
	return nil
}

func (p *PineconeStore) Query(ctx context.Context, sessionID string, queryVector []float32, topK int, filter port.MetadataFilter) ([]domain.Chunk, error) {
	idxConn, err := p.indexConnection(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	defer idxConn.Close()

	resp, err := idxConn.QueryByVectorValues(ctx, &sdk.QueryByVectorValuesRequest{
		Vector:          queryVector,
		TopK:            uint32(topK),
		IncludeMetadata: true,
	})
	if err != nil {
		return nil, fmt.Errorf("query by vector values: %w", err)
	}

	return chunksFromMatches(resp.Matches, sessionID, filter), nil
}

func (p *PineconeStore) DeleteSession(ctx context.Context, sessionID string) error {
	idxConn, err := p.indexConnection(ctx, sessionID)
	if err != nil {
		return err
	}
	defer idxConn.Close()

	if err := idxConn.DeleteAllVectorsInNamespace(ctx); err != nil {
		return fmt.Errorf("delete namespace %s: %w", sessionID, err)
	}
	return nil
}

func (p *PineconeStore) indexConnection(ctx context.Context, namespace string) (*sdk.IndexConnection, error) {
	idx, err := p.client.DescribeIndex(ctx, p.indexName)
	if err != nil {
		return nil, fmt.Errorf("describe index %s: %w", p.indexName, err)
	}
	conn, err := p.client.Index(sdk.NewIndexConnParams{Host: idx.Host, Namespace: namespace})
	if err != nil {
		return nil, fmt.Errorf("connect to index: %w", err)
	}
	return conn, nil
}

func buildMetadata(chunk domain.Chunk, chunkIndex int) (*structpb.Struct, error) {
	return structpb.NewStruct(map[string]interface{}{
		"has_preference":    chunk.Metadata.HasPreference,
		"has_wish":          chunk.Metadata.HasWish,
		"topics":            toAnySlice(chunk.Metadata.Topics),
		"emotional_markers": toAnySlice(chunk.Metadata.EmotionalMarkers),
		"chunk_index":       chunkIndex,
		"message_start":     chunk.StartIndex,
		"message_end":       chunk.EndIndex,
	})
}

func toAnySlice(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

func chunksFromMatches(matches []*sdk.ScoredVector, sessionID string, filter port.MetadataFilter) []domain.Chunk {
	var chunks []domain.Chunk
	for _, m := range matches {
		if m.Vector == nil || m.Vector.Metadata == nil {
			continue
		}
		meta := m.Vector.Metadata.AsMap()
		chunk := domain.Chunk{
			ID:        m.Vector.Id,
			SessionID: sessionID,
		}
		if v, ok := meta["has_preference"].(bool); ok {
			chunk.Metadata.HasPreference = v
		}
		if v, ok := meta["has_wish"].(bool); ok {
			chunk.Metadata.HasWish = v
		}
		if !matchesFilter(chunk, filter) {
			continue
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}
