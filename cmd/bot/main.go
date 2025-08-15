package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"
	"strings"

	"github.com/UjjavalParmar/twitter-automation/internal/config"
	"github.com/UjjavalParmar/twitter-automation/internal/gen"
	"github.com/UjjavalParmar/twitter-automation/internal/logging"
	"github.com/UjjavalParmar/twitter-automation/internal/scheduler"
	"github.com/UjjavalParmar/twitter-automation/internal/selector"
	"github.com/UjjavalParmar/twitter-automation/internal/storage"
	"github.com/UjjavalParmar/twitter-automation/internal/xclient"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	log := logging.New()
	cfg := config.Load()



	loc, err := time.LoadLocation(cfg.TZ)
	if err != nil {
		log.Fatal().Err(err).Msg("load TZ")
	}

	store, err := storage.Open(cfg.DataDir)
	if err != nil {
		log.Fatal().Err(err).Msg("open store")
	}
	defer store.Close()

	ctx := context.Background()
	genr, err := gen.New(ctx, gen.Options{
		APIKey:      cfg.GeminiKey,
		Model:       cfg.Model,
		MaxTokens:   int32(cfg.MaxTokens), // match latest API type
		Temperature: cfg.Temperature,
		TopP:        cfg.TopP,
		Lang:        cfg.Lang,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("genai client")
	}
	defer genr.Close()

	// log.Info().Msg("Running immediate test post...")
	//
	// tweet, err := genr.ComposeTweet(
	// 		context.Background(),
	// 		"AI in DevOps and job openings in india", // topic
	// 		"casual",       // style
	// )
	// if err != nil {
	// 		log.Fatal().Err(err).Msg("failed to compose tweet")
	// }
	//
	// log.Info().Str("tweet", tweet).Msg("Generated test tweet")


	// Prepare Twitter client with creds
	httpClient := &http.Client{}
	x := xclient.NewWithCreds(httpClient, xclient.Creds{
		APIKey:       cfg.XApiKey,
		APISecret:    cfg.XApiSecret,
		AccessToken:  cfg.XAccessToken,
		AccessSecret: cfg.XAccessSecret,
	})
	// schedule today’s slots
	slots, err := scheduler.DailyRandomSlots(loc, cfg.PostsPerDay, cfg.PostWindowStart, cfg.PostWindowEnd)
	if err != nil {
		log.Fatal().Err(err).Msg("schedule")
	}
	for _, s := range slots {
		log.Info().Time("time", s.Time).Str("key", s.Key).Msg("post slot")
	}

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	replyTicker := time.NewTicker(cfg.ReplyScanInterval)
	defer replyTicker.Stop()

	for {
		select {
		case <-stop:
			log.Info().Msg("shutting down")
			return

		case <-ticker.C:
			now := time.Now().In(loc)
			for _, s := range slots {
				if now.After(s.Time) {
					posted, _ := store.WasPosted(s.Key)
					if posted {
						continue
					}
					// generate & post
					go func(slot scheduler.Slot) {
						if err := doPost(ctx, log, genr, x, store, slot); err != nil {
							log.Error().Err(err).Str("slot", slot.Key).Msg("post failed")
						}
					}(s)
				}
			}

		case <-replyTicker.C:
			go func() {
				if err := doReplies(ctx, log, genr, x, store, cfg); err != nil {
					log.Error().Err(err).Msg("reply scan failed")
				}
			}()
		}
	}
}

func doPost(ctx context.Context, log zerolog.Logger, genr *gen.Generator, x *xclient.Client, store *storage.Store, slot scheduler.Slot) error {
	topicSet := selector.RandomTopicSet()
	style := selector.RandomStyle()

	text, err := genr.ComposeTweet(ctx,strings.Join(topicSet, ", "), style)
	if err != nil {
		return err
	}

	id, err := x.PostTweet(text)
	if err != nil {
		return err
	}

	log.Info().Str("id", id).Msg("posted tweet")
	return store.MarkPosted(slot.Key)
}

func doReplies(ctx context.Context, log zerolog.Logger, genr *gen.Generator, x *xclient.Client, store *storage.Store, cfg *config.Config) error {
	ts, err := x.SearchDevOpsRecent(100)
	if err != nil {
		return err
	}

	sort.Slice(ts, func(i, j int) bool {
		a := ts[i].PublicMetrics.LikeCount + 2*ts[i].PublicMetrics.RetweetCount + ts[i].PublicMetrics.ReplyCount
		b := ts[j].PublicMetrics.LikeCount + 2*ts[j].PublicMetrics.RetweetCount + ts[j].PublicMetrics.ReplyCount
		return a > b
	})

	count := 0
	for _, t := range ts {
		if count >= cfg.ReplyMaxPerScan {
			break
		}
		if t.PublicMetrics.LikeCount < cfg.ReplyMinLikes || t.PublicMetrics.RetweetCount < cfg.ReplyMinRetweets {
			continue
		}
		seen, _ := store.IsSeen(t.ID)
		if seen {
			continue
		}

		reply, err := genr.ComposeReply(ctx, t.Text, "author")
		if err != nil {
			log.Error().Err(err).Msg("gen reply")
			continue
		}

		if _, err := x.Reply(t.ID, reply); err != nil {
			log.Error().Err(err).Str("tid", t.ID).Msg("reply failed")
			continue
		}
		_ = store.SeenTweet(t.ID)
		log.Info().Str("tid", t.ID).Msg("replied")
		count++
	}

	if count > 0 {
		log.Info().Int("replies", count).Msg("reply pass")
	}
	return nil
}
