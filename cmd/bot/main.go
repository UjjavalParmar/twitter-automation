package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

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

const indexHTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8"/>
<title>Twitter Automation Dashboard</title>
<style>
body{font-family:Arial,sans-serif;margin:24px;}
.card{border:1px solid #ddd;border-radius:8px;padding:16px;margin-bottom:16px;}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:12px}
.stat{font-size:28px;font-weight:bold}
label{display:block;margin:8px 0 4px}
button{padding:8px 14px}
.bad{color:#b00}
.ok{color:#080}
</style>
</head>
<body>
<h2>Twitter Automation Dashboard</h2>
<div class="grid">
  <div class="card">
    <div>Posted Tweets</div>
    <div id="posted" class="stat">-</div>
  </div>
  <div class="card">
    <div>Replies Sent</div>
    <div id="replies" class="stat">-</div>
  </div>
  <div class="card">
    <div>Total Likes (recent)</div>
    <div id="likes" class="stat">-</div>
  </div>
  <div class="card">
    <div>Total Replies (recent)</div>
    <div id="replies_total" class="stat">-</div>
  </div>
</div>

<div class="card">
  <h3>Compose Tweet</h3>
  <div id="topics"></div>
  <label for="style">Style</label>
  <select id="style"></select>
  <div style="margin-top:12px">
    <button id="generate">Generate</button>
    <button id="post" disabled>Post</button>
    <button id="discard" disabled>Discard</button>
  </div>
  <div style="margin-top:10px">
    <label for="preview">Preview</label>
    <textarea id="preview" rows="5" style="width:100%" placeholder="Generated tweet will appear here..." disabled></textarea>
  </div>
  <div id="result" style="margin-top:10px"></div>
</div>

<script>
async function loadMeta(){
  const res = await fetch('/api/topics');
  const data = await res.json();
  const tEl = document.getElementById('topics');
  tEl.innerHTML = '';
  data.topics.forEach(function(grp, i){
    var div = document.createElement('div');
    div.style.margin='6px 0';
    var html = '<label>Topics Group ' + (i+1) + '</label>';
    for (var j=0;j<grp.length;j++) {
      var t = grp[j];
      html += '<label><input type="checkbox" name="topic" value="' + t + '"> ' + t + '</label> ';
    }
    div.innerHTML = html;
    tEl.appendChild(div);
  });
  const sEl = document.getElementById('style');
  var shtml = '';
  for (var k=0;k<data.styles.length;k++) { shtml += '<option>' + data.styles[k] + '</option>'; }
  sEl.innerHTML = shtml;
}
async function loadStats(){
  const res = await fetch('/api/stats');
  const s = await res.json();
  document.getElementById('posted').textContent = s.posted_count;
  document.getElementById('replies').textContent = s.reply_count;
  document.getElementById('likes').textContent = s.likes_total;
  document.getElementById('replies_total').textContent = s.replies_total;
}
let generated = '';
async function generateTweet(){
  var tops = document.querySelectorAll('input[name="topic"]:checked');
  var topics = [];
  for (var i=0;i<tops.length;i++){ topics.push(tops[i].value); }
  var style = document.getElementById('style').value;
  var result = document.getElementById('result');
  var preview = document.getElementById('preview');
  var postBtn = document.getElementById('post');
  var discardBtn = document.getElementById('discard');
  result.textContent = 'Generating...';
  try{
    var res = await fetch('/api/generate', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({topics: topics, style: style})});
    if(!res.ok){
      var t = await res.text();
      result.innerHTML = '<span class="bad">Failed: ' + t + '</span>';
      return;
    }
    var data = await res.json();
    generated = data.text || '';
    preview.value = generated;
    preview.disabled = false;
    postBtn.disabled = !generated;
    discardBtn.disabled = !generated;
    result.innerHTML = '<span class="ok">Preview generated. Review and click Post to publish.</span>';
  }catch(e){
    result.innerHTML = '<span class="bad">Error: ' + e + '</span>';
  }
}
async function postTweet(){
  var preview = document.getElementById('preview');
  var text = preview.value.trim();
  var result = document.getElementById('result');
  if(!text){ result.innerHTML = '<span class="bad">Nothing to post.</span>'; return; }
  result.textContent = 'Posting...';
  try{
    var res = await fetch('/api/post', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({text: text})});
    if(!res.ok){
      var t = await res.text();
      result.innerHTML = '<span class="bad">Failed: ' + t + '</span>';
      return;
    }
    var data = await res.json();
    result.innerHTML = '<span class="ok">Posted!</span> ID: ' + data.id + '<br/>Text: ' + String(text).replace(/</g,'&lt;');
    generated = '';
    preview.value = '';
    preview.disabled = true;
    document.getElementById('post').disabled = true;
    document.getElementById('discard').disabled = true;
    loadStats();
  }catch(e){
    result.innerHTML = '<span class="bad">Error: ' + e + '</span>';
  }
}
function discardTweet(){
  generated = '';
  var preview = document.getElementById('preview');
  preview.value = '';
  preview.disabled = true;
  document.getElementById('post').disabled = true;
  document.getElementById('discard').disabled = true;
  document.getElementById('result').textContent = 'Draft discarded.';
}
loadMeta();
loadStats();
setInterval(loadStats, 10000);
document.addEventListener('click', function(e){ 
  if(e.target && e.target.id==='generate'){ generateTweet(); }
  if(e.target && e.target.id==='post'){ postTweet(); }
  if(e.target && e.target.id==='discard'){ discardTweet(); }
});
</script>
</body>
</html>`

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

	// HTTP server for simple frontend
	mux := http.NewServeMux()
	mux.HandleFunc("/api/topics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		resp := map[string]any{
			"topics": selector.AllTopics(),
			"styles": selector.AllStyles(),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		postedCount, _ := store.CountPrefix("postedid:")
		replyCount, _ := store.CountPrefix("replyid:")
		// Aggregate metrics for recent posted tweets
		ids, _ := store.ListIDs("postedid:", 50)
		likes := 0
		replies := 0
		if len(ids) > 0 {
			tweets, err := x.GetTweets(ids)
			if err == nil {
				for _, t := range tweets {
					likes += t.PublicMetrics.LikeCount
					replies += t.PublicMetrics.ReplyCount
				}
			}
		}
		resp := map[string]any{
			"posted_count":  postedCount,
			"reply_count":   replyCount,
			"likes_total":   likes,
			"replies_total": replies,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	// Legacy single-shot generate+post endpoint (kept for backward compatibility)
	mux.HandleFunc("/api/tweet", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Topics []string `json:"topics"`
			Style  string   `json:"style"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid json"))
			return
		}
		if len(body.Topics) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("topics required"))
			return
		}
		if body.Style == "" {
			body.Style = selector.RandomStyle()
		}
		text, err := genr.ComposeTweet(r.Context(), strings.Join(body.Topics, ", "), body.Style)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to compose tweet"))
			return
		}
		id, err := x.PostTweet(text)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("failed to post tweet"))
			return
		}
		_ = store.AddPostedID(id)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "text": text})
	})
	// New two-step compose flow: generate -> post
	mux.HandleFunc("/api/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Topics []string `json:"topics"`
			Style  string   `json:"style"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid json"))
			return
		}
		if len(body.Topics) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("topics required"))
			return
		}
		if body.Style == "" {
			body.Style = selector.RandomStyle()
		}
		text, err := genr.ComposeTweet(r.Context(), strings.Join(body.Topics, ", "), body.Style)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to compose tweet"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"text": text})
	})
	mux.HandleFunc("/api/post", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid json"))
			return
		}
		text := gen.CleanTweetText(strings.TrimSpace(body.Text))
		if text == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("text required"))
			return
		}
		id, err := x.PostTweet(text)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("failed to post tweet"))
			return
		}
		_ = store.AddPostedID(id)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": id})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, indexHTML)
	})
	go func() {
		addr := ":8080"
		log.Info().Str("addr", addr).Msg("starting frontend server")
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Error().Err(err).Msg("http server stopped")
		}
	}()

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

	text, err := genr.ComposeTweet(ctx, strings.Join(topicSet, ", "), style)
	if err != nil {
		return err
	}

	id, err := x.PostTweet(text)
	if err != nil {
		return err
	}

	log.Info().Str("id", id).Msg("posted tweet")
	_ = store.AddPostedID(id)
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

		rid, err := x.Reply(t.ID, reply)
		if err != nil {
			log.Error().Err(err).Str("tid", t.ID).Msg("reply failed")
			continue
		}
		_ = store.SeenTweet(t.ID)
		_ = store.AddReplyID(rid)
		log.Info().Str("tid", t.ID).Str("rid", rid).Msg("replied")
		count++
	}

	if count > 0 {
		log.Info().Int("replies", count).Msg("reply pass")
	}
	return nil
}
