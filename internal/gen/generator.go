package gen

import (
	"context"
	"errors"

	"google.golang.org/api/option"
	"github.com/google/generative-ai-go/genai"
)

type Generator struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

type Options struct {
	APIKey      string
	Model       string
	MaxTokens   int32
	Temperature float32
	TopP        float32
	Lang        string
}

func New(ctx context.Context, opts Options) (*Generator, error) {
	if opts.APIKey == "" {
		return nil, errors.New("missing Gemini API key")
	}

	// Create Gemini API client with API key
	client, err := genai.NewClient(ctx, option.WithAPIKey(opts.APIKey))
	if err != nil {
		return nil, err
	}

	model := client.GenerativeModel(opts.Model)
	if opts.MaxTokens > 0 {
		model.GenerationConfig = genai.GenerationConfig{
			MaxOutputTokens: ptrInt32(opts.MaxTokens),
			Temperature:     ptrFloat32(opts.Temperature),
			TopP:            ptrFloat32(opts.TopP),
		}
	}

	return &Generator{
		client: client,
		model:  model,
	}, nil
}

func (g *Generator) Close() {
	if g.client != nil {
		g.client.Close()
	}
}

func (g *Generator) ComposeTweet(ctx context.Context, topic, style string) (string, error) {
	resp, err := g.model.GenerateContent(ctx, genai.Text(
		"Write a short, engaging tweet about "+topic+" in a "+style+" style.",
	))
	if err != nil {
		return "", err
	}
	return extractText(resp), nil
}

func (g *Generator) ComposeReply(ctx context.Context, tweetText, author string) (string, error) {
	resp, err := g.model.GenerateContent(ctx, genai.Text(
		"Reply to the following tweet by "+author+" in a friendly and concise manner:\n\n"+tweetText,
	))
	if err != nil {
		return "", err
	}
	return extractText(resp), nil
}

func extractText(resp *genai.GenerateContentResponse) string {
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		parts := resp.Candidates[0].Content.Parts
		if len(parts) > 0 {
			if text, ok := parts[0].(genai.Text); ok {
				return string(text)
			}
		}
	}
	return ""
}

func ptrInt32(v int32) *int32       { return &v }
func ptrFloat32(v float32) *float32 { return &v }

