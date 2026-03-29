package embedder

import (
	"context"
	"fmt"
	"sync"

	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/pipelines"
)

type HugotEmbedder struct {
	session  *hugot.Session
	pipeline *pipelines.FeatureExtractionPipeline
	mu       sync.Mutex
}

func NewHugot(modelsDir string) (*HugotEmbedder, error) {
	session, err := hugot.NewGoSession()
	if err != nil {
		return nil, fmt.Errorf("creating hugot session: %w", err)
	}

	opts := hugot.NewDownloadOptions()
	opts.OnnxFilePath = "onnx/model.onnx"

	modelPath, err := hugot.DownloadModel(
		"sentence-transformers/all-MiniLM-L6-v2",
		modelsDir,
		opts,
	)
	if err != nil {
		return nil, fmt.Errorf("downloading model: %w", err)
	}

	pipeline, err := hugot.NewPipeline(session, hugot.FeatureExtractionConfig{
		ModelPath:    modelPath,
		Name:         "embedder",
		OnnxFilename: "onnx/model.onnx",
	})
	if err != nil {
		return nil, fmt.Errorf("creating pipeline: %w", err)
	}

	return &HugotEmbedder{session: session, pipeline: pipeline}, nil
}

func (h *HugotEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	result, err := h.pipeline.RunPipeline([]string{text})
	if err != nil {
		return nil, fmt.Errorf("embedding: %w", err)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return result.Embeddings[0], nil
}

func (h *HugotEmbedder) EmbedBatch(_ context.Context, texts []string) ([][]float32, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	result, err := h.pipeline.RunPipeline(texts)
	if err != nil {
		return nil, fmt.Errorf("embedding batch: %w", err)
	}

	return result.Embeddings, nil
}

func (h *HugotEmbedder) Close() error {
	return h.session.Destroy()
}
