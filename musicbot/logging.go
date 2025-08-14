package musicbot

import (
	"log/slog"
)

func LogCommand(command string, guildID string, userID string) {
	/* log a command execution */
	slog.Info("command executed",
		slog.String("command", command),
		slog.String("guild_id", guildID),
		slog.String("user_id", userID),
	)
}

func LogSendError(err error, guildID string, userID string, ephemeral bool) {
	/* log an error when a failed attempt to send a message occurs. */
	slog.Error("failed to send message",
		slog.Any("error", err),
		slog.String("guild_id", guildID),
		slog.String("user_id", userID),
		slog.Bool("ephemeral", ephemeral),
	)
}

func LogCommandError(err error, command string, guildID string, userID string) {
	/* log a error when a failed attempt to run a command */
	slog.Error("failed to run command",
		slog.Any("error", err),
		slog.String("command", command),
		slog.String("guild_id", guildID),
		slog.String("user_id", userID),
	)
}
