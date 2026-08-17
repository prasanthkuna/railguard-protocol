package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/railguard/signgate/internal/api"
	"github.com/railguard/signgate/internal/config"
	"github.com/railguard/signgate/internal/logger"
	"github.com/railguard/signgate/internal/policy"
	"github.com/railguard/signgate/internal/reservation"
	"github.com/railguard/signgate/internal/store"
	"github.com/railguard/signgate/internal/watcher"
)

func main() {
	cfg := config.Load()
	log := logger.New()

	if err := cfg.Validate(); err != nil {
		log.Fatal().Err(err).Msg("invalid configuration")
	}
	if cfg.KMSSignerEnabled && cfg.KMSSignerKeyID == "" {
		log.Fatal().Msg("KMS_SIGNER_ENABLED requires KMS_SIGNER_KEY_ID")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pe, err := policy.New(cfg.OPAPolicyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("load opa policy")
	}

	var repo store.Repository
	if db, err := store.New(ctx, cfg.PostgresURL); err != nil {
		if cfg.IsLocal() && cfg.AllowNoopStore {
			log.Warn().Err(err).Msg("postgres not reachable; ALLOW_NOOP_STORE enabled")
			repo = store.NewNoop()
		} else {
			log.Fatal().Err(err).Msg("postgres required")
		}
	} else {
		defer db.Close()
		repo = db
	}

	var budget reservation.Authority
	if cfg.UsePostgresBudgetAuthority() {
		log.Info().Msg("using postgres budget authority")
		budget = reservation.NewPostgres(repo)
	} else {
		redisBudget := reservation.New(cfg.RedisAddr)
		if err := redisBudget.Ping(ctx); err != nil {
			log.Warn().Err(err).Msg("redis not reachable; reservation endpoints may fail")
		}
		budget = redisBudget
	}

	srv := api.New(log, cfg, pe, budget, repo)

	if cfg.WatcherEnabled {
		w := watcher.New(log, repo, budget, cfg)
		go w.Run(ctx)
	}

	if err := api.ListenAndServe(ctx, log, cfg, srv); err != nil {
		log.Fatal().Err(err).Msg("server stopped")
	}
}
