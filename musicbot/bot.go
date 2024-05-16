package musicbot

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgolink/v3/disgolink"
)

type Bot struct {
	Client        bot.Client
	Lavalink      disgolink.Client
	PlayerManager PlayerManager
}

func (b *Bot) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := b.Client.OpenGateway(ctx); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func(node disgolink.NodeConfig) {
		defer wg.Done()
		if _, err := b.Lavalink.AddNode(ctx, node); err != nil {
			slog.Error("failed to add lavalink node", slog.String("node", node.Name), slog.Any("error", err))
		} else {
			slog.Info("added lavalink node", slog.String("node", node.Name))
		}
	}(disgolink.NodeConfig{
		Name:      "test-node",
		Address:   "localhost:2333",
		Password:  "youshallnotpass",
		Secure:    false,
		SessionID: "",
	})

	wg.Wait()

	if node := b.Lavalink.BestNode(); node == nil {
		slog.Error("no node connected")
		os.Exit(-1)
	}

	return nil
}
