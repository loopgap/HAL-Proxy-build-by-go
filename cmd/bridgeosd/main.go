package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"

	"hal-proxy/internal/api"
	"hal-proxy/internal/config"
	"hal-proxy/internal/core"
	"hal-proxy/internal/store"
)

func main() {
	cfg, err := config.Load(os.Getenv("HAL_PROXY_CONFIG"))
	if err != nil {
		log.Fatal(err)
	}

	repo, err := store.NewSQLiteRepository(cfg.Database.Path)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	// Apply connection pool settings from config
	repo.DB().SetMaxOpenConns(cfg.Database.MaxOpenConns)
	repo.DB().SetMaxIdleConns(cfg.Database.MaxIdleConns)

	// Setup OTLP trace exporter (optional - won't error if collector not available)
	ctx := context.Background()
	exporter, err := otlptracehttp.New(ctx)
	if err == nil {
		tp := trace.NewTracerProvider(trace.WithBatcher(exporter))
		otel.SetTracerProvider(tp)
	}

	svc := core.NewService(repo, cfg.App.ArtifactsDir)
	if err := svc.Init(context.Background()); err != nil {
		log.Fatal(err)
	}

	server := api.NewServer(svc, repo.DB(), repo.Blacklist, cfg.Auth.JWTSecret, cfg.Auth.JWTExpiryHours, cfg.Auth.JWTIssuer, cfg.Auth.TrustedProxies)

	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      server.Handler(),
		ReadTimeout:  cfg.Server.GetReadTimeout(),
		WriteTimeout: cfg.Server.GetWriteTimeout(),
		IdleTimeout:  cfg.Server.GetIdleTimeout(),
	}

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("received shutdown signal, gracefully shutting down...")
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("hal-proxyd listening on %s", cfg.Server.Address)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
	log.Println("server stopped")
}
