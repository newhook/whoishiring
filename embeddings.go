package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"github.com/newhook/ai-hack/ollama"
	"github.com/newhook/ai-hack/openai"
	"github.com/newhook/ai-hack/queries"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"os"
	"time"
)

type Embedding struct {
	Model     string
	Embedding func(ctx context.Context, text string) ([]float32, error)
}

const (
	Nomic        = "nomic-embed-text"
	Gemma        = "gemma:2b"
	OpenAI3Small = string(openai.EmbeddingModelOpenAI3Small)
)

var embeddings = map[string]Embedding{
	Nomic: {
		Model:     Nomic,
		Embedding: ollama.Embedding(Nomic, ""),
	},
	Gemma: {
		Model:     Gemma,
		Embedding: ollama.Embedding(Gemma, ""),
	},
	OpenAI3Small: {
		Model:     OpenAI3Small,
		Embedding: openai.Embedding(os.Getenv("OPENAI_API_KEY"), openai.EmbeddingModelOpenAI(OpenAI3Small)),
	},
}

func GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	return embeddings[*embeddingModel].Embedding(ctx, text)
}

func CreateEmbeddings(ctx context.Context, l *slog.Logger, model string) error {
	q := queries.New(db)
	posts, err := q.GetItemsWithTitle(ctx, whoIsHiring)
	if err != nil {
		return err
	}

	var create []queries.Item
	// last three months.
	for _, post := range posts[:MaxWindow] {
		children, err := q.GetItemsForParent(ctx, post.ID)
		if err != nil {
			return err
		}
		embeddings, err := q.GetEmbeddingsByParent(ctx, queries.GetEmbeddingsByParentParams{
			Model:  model,
			Parent: post.ID,
		})
		if err != nil {
			return err
		}
		set := NewSet[int]()
		for _, e := range embeddings {
			set.Add(e.ItemID)
		}
		for _, c := range children {
			if !set.Contains(c.ID) {
				create = append(create, c)
			}
		}
	}

	l.Info("creating embeddings", slog.Int("count", len(create)))
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)
	for _, comment := range create {
		if comment.Text == "" {
			continue
		}

		g.Go(func() error {
			var blob []byte
			vector, err := embeddings[model].Embedding(ctx, comment.Text)
			if err != nil {
				return err
			}
			blob, err = MarshalFloat32ArrayWithLength(vector)
			if err != nil {
				return err
			}
			now := int(time.Now().Unix())
			err = q.InsertEmbedding(ctx, queries.InsertEmbeddingParams{
				ItemID:    comment.ID,
				Model:     model,
				Embedding: blob,
				CreatedAt: now,
				UpdatedAt: now,
			})
			if err != nil {
				return err
			}
			return nil
		})
	}
	return g.Wait()
}

// MarshalFloat32ArrayWithLength marshals an array of float32 values to a binary blob, including the length of the array at the beginning.
func MarshalFloat32ArrayWithLength(floats []float32) ([]byte, error) {
	buf := new(bytes.Buffer)
	// Write the length of the array first
	length := int32(len(floats))
	if err := binary.Write(buf, binary.LittleEndian, length); err != nil {
		return nil, err
	}
	// Write the float32 values
	for _, f := range floats {
		if err := binary.Write(buf, binary.LittleEndian, f); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// UnmarshalFloat32ArrayWithLength unmarshals a binary blob back into an array of float32, assuming the first value is the length of the array.
func UnmarshalFloat32ArrayWithLength(data []byte) ([]float32, error) {
	buf := bytes.NewReader(data)
	// Read the length of the array first
	var length int32
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	floats := make([]float32, length)
	for i := 0; i < int(length); i++ {
		if err := binary.Read(buf, binary.LittleEndian, &floats[i]); err != nil {
			return nil, err
		}
	}
	return floats, nil
}
