package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/channinghe/ems2sns/internal/config"
	"github.com/channinghe/ems2sns/internal/model"
	"github.com/channinghe/ems2sns/internal/notifier/discord"
	"github.com/channinghe/ems2sns/internal/notifier/telegram"
	"github.com/channinghe/ems2sns/internal/provider"
	"github.com/channinghe/ems2sns/internal/store"
	"github.com/channinghe/ems2sns/internal/tracker"
)

func main() {
	cfgFile := flag.String("config", "", "path to config file")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("ems2sns starting...")

	cfg, err := config.Load(*cfgFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config loaded (poll_interval=%v, telegram=%v, discord=%v)",
		cfg.Tracking.PollInterval, cfg.Telegram.Enabled, cfg.Discord.Enabled)

	// --- Build providers ---
	jpProvider := provider.NewJapanPostProvider("ja", model.SourceJapanPostJA)
	enProvider := provider.NewJapanPostProvider("en", model.SourceJapanPostEN)
	stProvider := provider.NewSeventeenTrackProvider(cfg.Tracking.SeventeenTrackToken)

	merger := provider.NewMerger(jpProvider, enProvider, stProvider)

	// --- Build store ---
	jsonStore := store.NewJSONStore(cfg.Storage.Path)

	// --- Build tracker ---
	trk := tracker.New(merger, stProvider, jsonStore, cfg)

	// --- Setup cross-platform mirror if enabled ---
	var tgBot *telegram.Bot
	var dcBot *discord.Bot

	if cfg.Telegram.Enabled {
		tgBot, err = telegram.New(cfg.Telegram, trk)
		if err != nil {
			log.Fatalf("Failed to create Telegram bot: %v", err)
		}
	}

	if cfg.Discord.Enabled {
		dcBot, err = discord.New(cfg.Discord, trk)
		if err != nil {
			log.Fatalf("Failed to create Discord bot: %v", err)
		}
	}

	// Register cross-platform mirror callbacks
	if cfg.CrossPlatform.Enabled {
		registerMirrors(cfg, tgBot, dcBot, trk)
	}

	// --- Start everything ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Tracker polling loop
	wg.Add(1)
	go func() {
		defer wg.Done()
		trk.Start(ctx)
	}()

	// Telegram bot
	if tgBot != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := tgBot.Start(ctx); err != nil {
				log.Printf("Telegram bot error: %v", err)
			}
		}()
		log.Println("Telegram bot started")
	}

	// Discord bot
	if dcBot != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := dcBot.Start(ctx); err != nil {
				log.Printf("Discord bot error: %v", err)
			}
		}()
		log.Println("Discord bot started")
	}

	log.Println("ems2sns is running. Press Ctrl+C to stop.")

	<-sigChan
	log.Println("Shutting down...")
	cancel()

	if tgBot != nil {
		tgBot.Stop()
	}
	if dcBot != nil {
		dcBot.Stop()
	}

	wg.Wait()
	log.Println("ems2sns stopped.")
}

// registerMirrors sets up cross-platform notification mirroring based on config rules
func registerMirrors(cfg *config.Config, tgBot *telegram.Bot, dcBot *discord.Bot, trk *tracker.Tracker) {
	for _, rule := range cfg.CrossPlatform.Mirrors {
		r := rule
		trk.RegisterNotifyFunc(func(sub *model.Subscription, info *model.TrackingInfo, delivered bool) {
			if sub.Platform != r.FromPlatform || sub.ChannelID != r.FromChannel {
				return
			}

			log.Printf("[mirror] %s:%s -> %s:%s for %s",
				r.FromPlatform, r.FromChannel, r.ToPlatform, r.ToChannel, sub.TrackingNumber)

			mirrorSub := &model.Subscription{
				TrackingNumber: sub.TrackingNumber,
				Platform:       r.ToPlatform,
				ChannelID:      r.ToChannel,
				UserID:         sub.UserID,
				Username:       sub.Username,
			}

			switch r.ToPlatform {
			case "telegram":
				if tgBot != nil {
					tgBot.SendUpdate(mirrorSub, info, delivered)
				}
			case "discord":
				if dcBot != nil {
					dcBot.SendUpdate(mirrorSub, info, delivered)
				}
			}
		})
	}

	log.Printf("Registered %d cross-platform mirror rules", len(cfg.CrossPlatform.Mirrors))
}
