# Twitter Automation Bot

A Go-based automation tool that generates and posts tweets using the Gemini API, schedules posting times, and replies to trending tweets in the DevOps space.
The bot uses AI to compose tweets and replies, stores posting history to avoid duplicates, and runs continuously with a scheduler.

---

## ✨ Features

- **AI-Powered Tweet Generation**
  Uses Google Gemini API to compose tweets and replies in different styles and languages.

- **Automated Tweet Posting**
  Schedules daily tweet slots within a configurable posting window.

- **Auto Replies to Trending Tweets**
  Monitors recent DevOps-related tweets, filters by popularity, and posts AI-generated replies.

- **Persistent Storage**
  Keeps track of posted tweets and replied tweets to avoid repetition.

- **Graceful Shutdown**
  Handles `SIGINT` and `SIGTERM` for safe exit.

---

## ⚙️ Requirements

- Go 1.24+
- Twitter/X Developer API credentials
- Google Gemini API key

## 📄 License

This project is licensed under the MIT License.
