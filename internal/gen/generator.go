package gen

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
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
	return CleanTweetText(extractText(resp)), nil
}

func (g *Generator) ComposeReply(ctx context.Context, tweetText, author string) (string, error) {
	resp, err := g.model.GenerateContent(ctx, genai.Text(
		"Reply to the following tweet by "+author+" in a friendly and concise manner:\n\n"+tweetText,
	))
	if err != nil {
		return "", err
	}
	return CleanTweetText(extractText(resp)), nil
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

// CleanTweetText normalizes model outputs into human-like, postable tweets.
// It strips Markdown formatting (**, __, *, _, backticks, code fences), headings,
// extra quotes, collapses whitespace/newlines, and trims to 280 chars.
func CleanTweetText(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	// Remove common prefixes
	for _, p := range []string{"Tweet:", "Draft:", "Suggestion:", "Here is a tweet:", "Here’s a tweet:", "Possible tweet:"} {
		if strings.HasPrefix(strings.ToLower(s), strings.ToLower(p)) {
			s = strings.TrimSpace(s[len(p):])
			break
		}
	}
	// Remove code fences and inline backticks
	reCodeFence := regexp.MustCompile("```[\\s\\S]*?```")
	s = reCodeFence.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "`", "")
	// Strip Markdown bold/italics markers
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "_", "")
	// Remove leading '#' heading markers per line
	reHeading := regexp.MustCompile(`(?m)^#+\\s*`)
	s = reHeading.ReplaceAllString(s, "")
	// Collapse multiple spaces and newlines
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	reMultiNewline := regexp.MustCompile("\n{3,}")
	s = reMultiNewline.ReplaceAllString(s, "\n\n")
	reMultiSpace := regexp.MustCompile("[ \t]{2,}")
	s = reMultiSpace.ReplaceAllString(s, " ")
	// Trim surrounding quotes
	s = strings.Trim(s, "\"'\n ")
	// Final trim
	s = strings.TrimSpace(s)
	// Ensure within 280 chars (UTF-8 safe)
	if utf8.RuneCountInString(s) > 280 {
		r := []rune(s)
		s = string(r[:280])
		// avoid cutting in the middle of a word by trimming to last space if feasible
		if i := strings.LastIndex(s, " "); i >= 200 { // try not to over-trim small texts
			s = strings.TrimSpace(s[:i])
		}
	}
	return s
}

func ptrInt32(v int32) *int32       { return &v }
func ptrFloat32(v float32) *float32 { return &v }
