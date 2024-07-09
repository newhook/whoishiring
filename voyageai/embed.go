package voyageai

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
	"io"
	"math"
	"net/http"
	"os"
)

const voyageURL = "https://api.voyageai.com/v1/embeddings"

type EmbeddingModel string

const (
	Voyage2Model             EmbeddingModel = "voyage-2"
	VoyageLarge2Model        EmbeddingModel = "voyage-large-2"
	VoyageFinance2Model      EmbeddingModel = "voyage-finance-2"
	VoyageMultilingual2Model EmbeddingModel = "voyage-multilingual-2"
	VoyageLaw2Model          EmbeddingModel = "voyage-law-2"
	VoyageCode2Model         EmbeddingModel = "voyage-code-2"
)

var apiKey = os.Getenv("VOYAGE_API_KEY")

type ApiResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

type RateLimitedClient struct {
	client  *http.Client
	limiter *rate.Limiter
}

func NewRateLimitedClient(requestsPerSecond float64) *RateLimitedClient {
	return &RateLimitedClient{
		client:  http.DefaultClient,
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
	}
}

func (c *RateLimitedClient) Do(req *http.Request) (*http.Response, error) {
	err := c.limiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

var client = NewRateLimitedClient(4) // 5 per second is 300 per minute, we'll go slightly lower.

func Embedding(model EmbeddingModel) func(ctx context.Context, text string) ([]float32, error) {
	return func(ctx context.Context, text string) ([]float32, error) {
		// Prepare the request body.
		reqBody, err := json.Marshal(map[string]any{
			"model": model,
			"input": []string{text},
		})
		if err != nil {
			return nil, errors.Errorf("couldn't marshal request body: %w", err)
		}

		// Create the request. Creating it with context is important for a timeout
		// to be possible, because the client is configured without a timeout.
		req, err := http.NewRequestWithContext(ctx, "POST", voyageURL, bytes.NewBuffer(reqBody))
		if err != nil {
			return nil, errors.Errorf("couldn't create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		// Send the request.
		resp, err := client.Do(req)
		if err != nil {
			return nil, errors.Errorf("couldn't send request: %w", err)
		}
		defer resp.Body.Close()

		// Check the response status.
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("error response from the embedding API: " + resp.Status)
		}

		// Read and decode the response body.
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Errorf("couldn't read response body: %w", err)
		}
		var embeddingResponse ApiResponse
		err = json.Unmarshal(body, &embeddingResponse)
		if err != nil {
			return nil, errors.Errorf("couldn't unmarshal response body: %w", err)
		}

		// Check if the response contains embeddings.
		if len(embeddingResponse.Data) == 0 {
			return nil, errors.New("no embeddings found in the response")
		}

		return embeddingResponse.Data[0].Embedding, nil
	}
}

const isNormalizedPrecisionTolerance = 1e-6

func isNormalized(v []float32) bool {
	var sqSum float64
	for _, val := range v {
		sqSum += float64(val) * float64(val)
	}
	magnitude := math.Sqrt(sqSum)
	return math.Abs(magnitude-1) < isNormalizedPrecisionTolerance
}

func normalizeVector(v []float32) []float32 {
	var norm float32
	for _, val := range v {
		norm += val * val
	}
	norm = float32(math.Sqrt(float64(norm)))

	res := make([]float32, len(v))
	for i, val := range v {
		res[i] = val / norm
	}

	return res
}
