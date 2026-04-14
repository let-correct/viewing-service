package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awseventbridge "github.com/aws/aws-sdk-go-v2/service/eventbridge"

	ebviewing "github.com/troysnowden/viewing-service/internal/adapters/eventbridge/viewing"
	appviewing "github.com/troysnowden/viewing-service/internal/application/viewing"
	"github.com/troysnowden/viewing-service/internal/config"
	viewinghandler "github.com/troysnowden/viewing-service/internal/handler/viewing"
)

func main() {
	cfg, err := config.LoadViewingWorker()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel(cfg.LogLevel),
	}))

	ctx := context.Background()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "failed to load AWS config", "error", err)
		os.Exit(1)
	}

	eventbridgeClient := awseventbridge.NewFromConfig(awsCfg)
	publisher := ebviewing.New(eventbridgeClient, cfg.EventBusName)
	scheduler := appviewing.NewProcessAppointment(publisher)
	h := viewinghandler.New(logger, scheduler)

	lambda.StartWithOptions(h.Handle, lambda.WithContext(ctx))
}

func logLevel(level string) slog.Level {
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
