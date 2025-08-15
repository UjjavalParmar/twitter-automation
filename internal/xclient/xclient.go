package xclient

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	rest  *resty.Client
	creds Creds
}

type Creds struct {
	APIKey       string
	APISecret    string
	AccessToken  string
	AccessSecret string
}

// NewWithCreds initializes a new Client with OAuth1 signing.
func NewWithCreds(httpClient *http.Client, creds Creds) *Client {
	// OAuth1 config for X (Twitter)
	config := oauth1.NewConfig(creds.APIKey, creds.APISecret)
	token := oauth1.NewToken(creds.AccessToken, creds.AccessSecret)

	// Create an HTTP client that automatically signs requests
	oauthClient := config.Client(oauth1.NoContext, token)

	rc := resty.NewWithClient(oauthClient).
		SetBaseURL("https://api.twitter.com/2"). // Correct base URL for X API
		SetRetryCount(4).
		SetRetryWaitTime(2 * time.Second).
		SetRetryMaxWaitTime(20 * time.Second)

	return &Client{rest: rc, creds: creds}
}

// PostTweet posts a new tweet.
func (c *Client) PostTweet(text string) (string, error) {
	var resp struct {
		Data struct{ ID, Text string } `json:"data"`
	}
	r, err := c.rest.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"text": text}).
		SetResult(&resp).
		Post("/tweets")
	if err != nil {
		return "", err
	}
	if r.StatusCode() != http.StatusCreated && r.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("post tweet failed: %s - %s", r.Status(), r.String())
	}
	return resp.Data.ID, nil
}

type Tweet struct {
	ID            string `json:"id"`
	Text          string `json:"text"`
	AuthorID      string `json:"author_id"`
	PublicMetrics struct {
		RetweetCount int `json:"retweet_count"`
		ReplyCount   int `json:"reply_count"`
		LikeCount    int `json:"like_count"`
		QuoteCount   int `json:"quote_count"`
	} `json:"public_metrics"`
}

type searchResp struct {
	Data []Tweet `json:"data"`
	Meta struct {
		NextToken string `json:"next_token"`
	} `json:"meta"`
}

// SearchDevOpsRecent searches for recent DevOps tweets.
func (c *Client) SearchDevOpsRecent(max int) ([]Tweet, error) {
	q := `("Kubernetes" OR K8s OR "CI/CD" OR "SRE" OR "Terraform" OR "OpenTofu" OR "ArgoCD" OR "OpenTelemetry" OR "Istio" OR "FinOps" OR "supply chain security") lang:en -is:retweet -is:quote`

	var out []Tweet
	params := map[string]string{
		"query":        q,
		"max_results":  "50",
		"tweet.fields": "public_metrics,author_id,created_at",
	}

	var resp searchResp
	r, err := c.rest.R().
		SetQueryParams(params).
		SetResult(&resp).
		Get("/tweets/search/recent")
	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, fmt.Errorf("search failed: %s - %s", r.Status(), r.String())
	}
	for _, t := range resp.Data {
		out = append(out, t)
		if len(out) >= max {
			break
		}
	}
	return out, nil
}

// Reply posts a reply to a tweet.
func (c *Client) Reply(tweetID string, text string) (string, error) {
	payload := map[string]any{
		"text": text,
		"reply": map[string]string{
			"in_reply_to_tweet_id": tweetID,
		},
	}
	var resp struct {
		Data struct{ ID, Text string } `json:"data"`
	}
	r, err := c.rest.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		SetResult(&resp).
		Post("/tweets")
	if err != nil {
		return "", err
	}
	if r.IsError() {
		return "", fmt.Errorf("reply failed: %s - %s", r.Status(), r.String())
	}
	return resp.Data.ID, nil
}

