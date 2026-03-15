package usecase

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

var baseRetrievalQueries = []string{
	"What activities, hobbies, passions, and interests does this person mention wanting to pursue or already enjoying?",
	"What personality traits, emotional patterns, and communication style does this person show? What makes them happy or laugh?",
	"What things has this person explicitly said they want, wish for, or plan to do someday? Any specific items or experiences mentioned?",
	"What is the nature of this relationship and what recurring shared themes or experiences do they have together?",
}

func BuildRetrievalQueries(recipient domain.RecipientDetails, numQueries int) []string {
	prefix := fmt.Sprintf("For a %s gift for %s: ", recipient.Occasion, recipient.Relation)
	queries := make([]string, 0, numQueries)
	for i := 0; i < numQueries && i < len(baseRetrievalQueries); i++ {
		queries = append(queries, prefix+baseRetrievalQueries[i])
	}
	for len(queries) < numQueries {
		queries = append(queries, prefix+baseRetrievalQueries[0])
	}
	return queries
}

func RetrieveAndRerank(
	ctx context.Context,
	sessionID string,
	queries []string,
	chunksByID map[string]domain.Chunk,
	embedder port.Embedder,
	store port.VectorStore,
	topK int,
) ([]domain.Chunk, error) {
	queryVectors, err := embedder.Embed(ctx, queries)
	if err != nil {
		return nil, fmt.Errorf("embed retrieval queries: %w", err)
	}

	allResults := make([][]domain.Chunk, len(queries))
	errCh := make(chan error, len(queries))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, vec := range queryVectors {
		wg.Add(1)
		go func(idx int, v []float32) {
			defer wg.Done()
			chunks, qErr := store.Query(ctx, sessionID, v, topK, port.MetadataFilter{})
			if qErr != nil {
				errCh <- fmt.Errorf("query %d failed: %w", idx, qErr)
				return
			}
			mu.Lock()
			allResults[idx] = chunks
			mu.Unlock()
		}(i, vec)
	}

	wg.Wait()
	close(errCh)

	if qErr := <-errCh; qErr != nil {
		return nil, qErr
	}

	return dedupAndRerank(allResults, chunksByID), nil
}

func dedupAndRerank(allResults [][]domain.Chunk, chunksByID map[string]domain.Chunk) []domain.Chunk {
	seen := make(map[string]int)
	var unique []domain.Chunk

	for _, chunks := range allResults {
		for _, c := range chunks {
			if _, exists := seen[c.ID]; !exists {
				seen[c.ID] = 0
				if full, ok := chunksByID[c.ID]; ok {
					unique = append(unique, full)
				} else {
					unique = append(unique, c)
				}
			}
			seen[c.ID]++
		}
	}

	sort.SliceStable(unique, func(i, j int) bool {
		hi := unique[i].Metadata.HasPreference || unique[i].Metadata.HasWish
		hj := unique[j].Metadata.HasPreference || unique[j].Metadata.HasWish
		if hi != hj {
			return hi
		}
		return seen[unique[i].ID] > seen[unique[j].ID]
	})

	return unique
}
