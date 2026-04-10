# ems2sns

EMS package tracking with push notifications to Telegram and Discord.

Polls tracking data from Japan Post and 17track, detects status changes, and sends updates to your chats automatically.

Supported carriers:
- Japan Post (JP/EN)
- China EMS via 17track
- CN-suffix tracking numbers auto-merge JP + CN + EN segments

[中文](README_zh.md)

## Quick Start

Create a `docker-compose.yaml`:

```yaml
services:
  ems2sns:
    image: ghcr.io/channinghe/ems2sns:latest
    container_name: ems2sns
    restart: unless-stopped
    environment:
      - EMS2SNS_APP_LANGUAGE=en
      - EMS2SNS_TELEGRAM_ENABLED=true
      - EMS2SNS_TELEGRAM_BOT_TOKEN=your-telegram-bot-token
      - EMS2SNS_TRACKING_SEVENTEEN_TRACK_TOKEN=your-17track-token
      - EMS2SNS_STORAGE_PATH=/app/data/subscriptions.json
      - TZ=Asia/Tokyo
    volumes:
      - ./data:/app/data
```

```bash
docker compose up -d
```

## Commands

**Telegram** -- slash commands or inline buttons via `/start`:

| Command | Description |
|---------|-------------|
| `/sub <tracking_number>` | Subscribe to tracking |
| `/unsub <tracking_number>` | Unsubscribe |
| `/list` | List active subscriptions |
| `/check <tracking_number>` | Query current status |
| `/push` | Force check all subscriptions now |
| `/help` | Show help |

**Discord** -- slash commands:

`/sub`, `/unsub`, `/list`, `/check`, `/push`, `/emshelp`

## Configuration

Supports YAML config file and/or environment variables (`EMS2SNS_` prefix). Env vars override config file values.

Mapping rule: `telegram.bot_token` -> `EMS2SNS_TELEGRAM_BOT_TOKEN`

<details>
<summary>Config file reference (config.example.yaml)</summary>

```yaml
app:
  log_level: "info"           # "info" or "debug"

tracking:
  poll_interval: "30m"        # supports s/m/h
  seventeen_track_token: ""   # 17track.net API token
  request_delay: "2s"         # delay between provider requests

storage:
  path: "data/subscriptions.json"

telegram:
  enabled: true
  bot_token: ""
  allowed_user_ids: []        # Telegram user IDs for private chat access
  allowed_chat_ids: []        # Group/supergroup chat IDs allowed
  push_chat_ids: []           # Centralized push targets (optional)

discord:
  enabled: false
  bot_token: ""
  allowed_guild_ids: []       # Discord server IDs allowed
  allowed_channel_ids: []     # Channel IDs allowed
  push_channel_ids: []        # Centralized push targets (optional)

cross_platform:
  enabled: false
  mirrors:
    - from_platform: "telegram"
      from_channel: "123456789"
      to_platform: "discord"
      to_channel: "987654321"
```

</details>

<details>
<summary>Environment variables</summary>

| Variable | Default | Description |
|----------|---------|-------------|
| `EMS2SNS_APP_LOG_LEVEL` | `info` | Log level (`info`, `debug`) |
| `EMS2SNS_TRACKING_POLL_INTERVAL` | `30m` | Polling interval |
| `EMS2SNS_TRACKING_SEVENTEEN_TRACK_TOKEN` | | 17track API token |
| `EMS2SNS_TRACKING_REQUEST_DELAY` | `2s` | Delay between requests |
| `EMS2SNS_STORAGE_PATH` | `data/subscriptions.json` | Subscription data path |
| `EMS2SNS_TELEGRAM_ENABLED` | `false` | Enable Telegram bot |
| `EMS2SNS_TELEGRAM_BOT_TOKEN` | | Telegram bot token |
| `EMS2SNS_TELEGRAM_ALLOWED_USER_IDS` | | Comma-separated user IDs |
| `EMS2SNS_TELEGRAM_ALLOWED_CHAT_IDS` | | Comma-separated chat IDs |
| `EMS2SNS_TELEGRAM_PUSH_CHAT_IDS` | | Centralized push targets |
| `EMS2SNS_DISCORD_ENABLED` | `false` | Enable Discord bot |
| `EMS2SNS_DISCORD_BOT_TOKEN` | | Discord bot token |
| `EMS2SNS_DISCORD_ALLOWED_GUILD_IDS` | | Comma-separated guild IDs |
| `EMS2SNS_DISCORD_ALLOWED_CHANNEL_IDS` | | Comma-separated channel IDs |
| `EMS2SNS_DISCORD_PUSH_CHANNEL_IDS` | | Centralized push targets |
| `EMS2SNS_CROSS_PLATFORM_ENABLED` | `false` | Enable cross-platform mirroring |

</details>

## Debug

Set `log_level` to `debug` or `EMS2SNS_APP_LOG_LEVEL=debug` to enable verbose logging. In debug mode, the Telegram bot logs every raw update received and enables the underlying API client's debug output.

## TODO

- [ ] Discord integration testing
