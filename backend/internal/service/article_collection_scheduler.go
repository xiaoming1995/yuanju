package service

import (
	"context"
	"errors"
	"log"
	"time"
	"yuanju/internal/model"
	"yuanju/internal/repository"
)

func StartArticleCollectionScheduler(ctx context.Context, provider ArticleSearchProvider) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	runScheduledArticleCollection(ctx, provider)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runScheduledArticleCollection(ctx, provider)
		}
	}
}

func runScheduledArticleCollection(ctx context.Context, provider ArticleSearchProvider) {
	if err := ensureArticleModuleEnabledForCollection(); err != nil {
		if !errors.Is(err, ErrArticleModuleDisabled) {
			log.Printf("[article-collection] load module setting failed: %v", err)
		}
		return
	}
	cfg, err := repository.GetArticleCollectionConfig()
	if err != nil {
		log.Printf("[article-collection] load config failed: %v", err)
		return
	}
	if cfg == nil || !cfg.Enabled || !articleCollectionDue(cfg) {
		return
	}
	if _, err := RunManualArticleCollectionTask(ctx, configureArticleSearchProvider(provider, cfg), cfg.MaxResultsPerRun); err != nil {
		log.Printf("[article-collection] scheduled run failed: %v", err)
	}
	if err := repository.MarkArticleCollectionConfigLastRun(); err != nil {
		log.Printf("[article-collection] mark last_run failed: %v", err)
	}
}

func configureArticleSearchProvider(provider ArticleSearchProvider, cfg *model.ArticleCollectionConfig) ArticleSearchProvider {
	if cfg == nil {
		return provider
	}
	if sogou, ok := provider.(*SogouWeChatProvider); ok {
		return sogou.WithSearchPageRange(cfg.SearchPageMin, cfg.SearchPageMax)
	}
	return provider
}

func articleCollectionDue(cfg *model.ArticleCollectionConfig) bool {
	if cfg.LastRunAt == nil {
		return true
	}
	interval := cfg.ScheduleInterval
	if interval <= 0 {
		if cfg.Frequency == "weekly" {
			interval = 10080
		} else {
			interval = 1440
		}
	}
	if interval < 1 {
		interval = 1
	}
	return time.Since(*cfg.LastRunAt) >= time.Duration(interval)*time.Minute
}
