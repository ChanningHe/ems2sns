# ems2sns

EMS 快递追踪，自动推送状态更新到 Telegram / Discord。

从日本邮政和 17track 轮询物流数据，检测到状态变化时自动推送通知。

[English](README.md)

支持的运营商：
- 日本邮政（日文/英文）
- 中国 EMS（通过 17track）
- CN 结尾的单号自动合并中日英三段物流信息

## 快速开始

创建 `docker-compose.yaml`：

```yaml
services:
  ems2sns:
    image: ghcr.io/channinghe/ems2sns:latest
    container_name: ems2sns
    restart: unless-stopped
    environment:
      - EMS2SNS_APP_LANGUAGE=cn
      - EMS2SNS_TELEGRAM_ENABLED=true
      - EMS2SNS_TELEGRAM_BOT_TOKEN=你的-telegram-bot-token
      - EMS2SNS_TRACKING_SEVENTEEN_TRACK_TOKEN=你的-17track-token
      - EMS2SNS_STORAGE_PATH=/app/data/subscriptions.json
      - TZ=Asia/Tokyo
    volumes:
      - ./data:/app/data
```

```bash
docker compose up -d
```

## 命令

**Telegram** -- 斜杠命令或通过 `/start` 使用按钮菜单：

| 命令 | 说明 |
|------|------|
| `/sub <追踪号>` | 订阅追踪 |
| `/unsub <追踪号>` | 取消订阅 |
| `/list` | 查看订阅列表 |
| `/check <追踪号>` | 查询当前状态 |
| `/push` | 立即检查所有订阅 |
| `/help` | 显示帮助 |

**Discord** -- 斜杠命令：

`/sub`, `/unsub`, `/list`, `/check`, `/push`, `/emshelp`

## 配置

支持 YAML 配置文件和环境变量（`EMS2SNS_` 前缀），环境变量优先级高于配置文件。

映射规则：`telegram.bot_token` -> `EMS2SNS_TELEGRAM_BOT_TOKEN`

配置项和环境变量的完整列表见 [README.md](README.md#configuration)。

## 调试

将 `log_level` 设为 `debug` 启用详细日志。debug 模式下 Telegram bot 会记录每一条收到的原始 update。

```bash
EMS2SNS_APP_LOG_LEVEL=debug docker compose up
```

## TODO

- [ ] Discord 集成测试
- [ ] Telegram Webhook 模式（替代长轮询）
