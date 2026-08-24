package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"yuanju/internal/model"
)

type ArticleSearchResult struct {
	Rank               int
	Title              string
	SourceName         string
	OriginalURL        string
	CoverURL           string
	PublishedAtSource  *time.Time
	SearchSnippet      string
	FullTextAuthorized bool
	BodyContent        string
	BodyFetchStatus    model.ArticleBodyFetchStatus
	BodyFetchError     string
}

type ArticleCollectionCandidate struct {
	Rank               int
	Title              string
	SourceName         string
	OriginalURL        string
	CanonicalURLHash   string
	CoverURL           string
	PublishedAtSource  *time.Time
	SearchSnippet      string
	FullTextAuthorized bool
	BodyContent        string
	BodyFetchStatus    model.ArticleBodyFetchStatus
	BodyFetchError     string
	QualityScore       int
	QualityReasons     []model.ArticleQualityReason
}

type ArticleSearchProvider interface {
	Search(ctx context.Context, keyword string, limit int) ([]ArticleSearchResult, error)
}

type ArticleCollectionStore interface {
	ListKeywords(ctx context.Context) ([]model.ArticleKeyword, error)
	InsertCandidate(ctx context.Context, input ArticleCollectionCandidate) (articleID string, inserted bool, err error)
	RecordItem(ctx context.Context, articleID, originalURL, keyword string, status model.CollectionItemStatus, errorMsg string) error
}

type ArticleCollectionRetryStore interface {
	ArticleCollectionStore
	ListFailedKeywords(ctx context.Context, taskID string) ([]string, error)
}

type ArticleQualityConfigStore interface {
	GetQualityConfig(ctx context.Context) (*model.ArticleQualityConfig, error)
}

type ArticleCollectionConfigStore interface {
	GetCollectionConfig(ctx context.Context) (*model.ArticleCollectionConfig, error)
}

type ArticleCollectionQualityItemStore interface {
	RecordItemWithQuality(ctx context.Context, articleID, originalURL, keyword string, status model.CollectionItemStatus, errorMsg string, quality ArticleQualityResult) error
}

type ArticleAutoPublishStore interface {
	AutoPublishArticle(ctx context.Context, articleID string, qualityScore, threshold int) error
}

type ArticleCollectionService struct {
	provider ArticleSearchProvider
	store    ArticleCollectionStore
}

func NewArticleCollectionService(provider ArticleSearchProvider, store ArticleCollectionStore) *ArticleCollectionService {
	return &ArticleCollectionService{provider: provider, store: store}
}

func (s *ArticleCollectionService) RunManual(ctx context.Context, maxResults int) (model.ArticleCollectionTask, error) {
	task := model.ArticleCollectionTask{
		TriggerType: "manual",
		Status:      model.CollectionTaskStatusSucceeded,
		StartedAt:   time.Now(),
	}
	if s == nil || s.provider == nil || s.store == nil {
		task.Status = model.CollectionTaskStatusFailed
		task.ErrorMsg = "collection service 未配置"
		return task, errors.New(task.ErrorMsg)
	}

	keywords, err := s.store.ListKeywords(ctx)
	if err != nil {
		task.Status = model.CollectionTaskStatusFailed
		task.ErrorMsg = err.Error()
		return task, err
	}
	activeKeywords := make([]model.ArticleKeyword, 0, len(keywords))
	for _, kw := range keywords {
		if kw.Active && strings.TrimSpace(kw.Keyword) != "" {
			activeKeywords = append(activeKeywords, kw)
		}
	}
	task.KeywordCount = len(activeKeywords)

	return s.runKeywords(ctx, "manual", activeKeywords, maxResults)
}

func (s *ArticleCollectionService) RetryFailedKeywords(ctx context.Context, taskID string, maxResults int) (model.ArticleCollectionTask, error) {
	task := model.ArticleCollectionTask{
		TriggerType: "retry",
		Status:      model.CollectionTaskStatusSucceeded,
		StartedAt:   time.Now(),
	}
	retryStore, ok := s.store.(ArticleCollectionRetryStore)
	if !ok {
		task.Status = model.CollectionTaskStatusFailed
		task.ErrorMsg = "collection store 不支持重试查询"
		return task, errors.New(task.ErrorMsg)
	}
	keywords, err := retryStore.ListFailedKeywords(ctx, taskID)
	if err != nil {
		task.Status = model.CollectionTaskStatusFailed
		task.ErrorMsg = err.Error()
		return task, err
	}
	items := make([]model.ArticleKeyword, 0, len(keywords))
	for _, keyword := range keywords {
		if strings.TrimSpace(keyword) != "" {
			items = append(items, model.ArticleKeyword{Keyword: keyword, Active: true})
		}
	}
	return s.runKeywords(ctx, "retry", items, maxResults)
}

func (s *ArticleCollectionService) runKeywords(ctx context.Context, triggerType string, activeKeywords []model.ArticleKeyword, maxResults int) (model.ArticleCollectionTask, error) {
	task := model.ArticleCollectionTask{
		TriggerType:  triggerType,
		Status:       model.CollectionTaskStatusSucceeded,
		StartedAt:    time.Now(),
		KeywordCount: len(activeKeywords),
	}
	var firstErr error
	qualityConfig := model.ArticleQualityConfig{AllowWithoutBody: true, MinQualityScore: 60}
	collectionConfig := model.ArticleCollectionConfig{
		AutoPublishMinQualityScore: 90,
		AutoPublishRequiresBody:    true,
		AutoPublishMaxPerRun:       3,
	}
	if qualityStore, ok := s.store.(ArticleQualityConfigStore); ok {
		if cfg, err := qualityStore.GetQualityConfig(ctx); err != nil {
			firstErr = err
			task.Status = model.CollectionTaskStatusFailed
			task.ErrorMsg = err.Error()
		} else if cfg != nil {
			qualityConfig = *cfg
		}
	}
	if configStore, ok := s.store.(ArticleCollectionConfigStore); ok {
		if cfg, err := configStore.GetCollectionConfig(ctx); err != nil {
			firstErr = err
			task.Status = model.CollectionTaskStatusFailed
			task.ErrorMsg = err.Error()
		} else if cfg != nil {
			collectionConfig = *cfg
		}
	}
	normalizeCollectionAutoPublishConfig(&collectionConfig)
	autoPublishedCount := 0
	remainingResults := maxResults
	for idx, kw := range activeKeywords {
		if maxResults > 0 && remainingResults <= 0 {
			break
		}
		keywordLimit := maxResults
		if maxResults > 0 {
			remainingKeywords := len(activeKeywords) - idx
			keywordLimit = (remainingResults + remainingKeywords - 1) / remainingKeywords
			if keywordLimit < 1 {
				keywordLimit = 1
			}
		}
		results, err := s.provider.Search(ctx, kw.Keyword, keywordLimit)
		if err != nil {
			task.FailedCount++
			task.Status = model.CollectionTaskStatusFailed
			if firstErr == nil {
				firstErr = err
			}
			if task.ErrorMsg == "" {
				task.ErrorMsg = err.Error()
			}
			_ = s.store.RecordItem(ctx, "", "", kw.Keyword, model.CollectionItemStatusFailed, err.Error())
			continue
		}
		if maxResults > 0 && len(results) > remainingResults {
			results = results[:remainingResults]
		}
		task.FoundCount += len(results)
		if maxResults > 0 {
			remainingResults -= len(results)
		}
		for _, result := range results {
			normalized, normErr := NormalizeArticleOriginalURL(result.OriginalURL)
			if normErr != nil {
				task.FailedCount++
				_ = s.store.RecordItem(ctx, "", result.OriginalURL, kw.Keyword, model.CollectionItemStatusFailed, normErr.Error())
				continue
			}
			candidate := ArticleCollectionCandidate{
				Rank:               result.Rank,
				Title:              strings.TrimSpace(result.Title),
				SourceName:         strings.TrimSpace(result.SourceName),
				OriginalURL:        normalized,
				CoverURL:           strings.TrimSpace(result.CoverURL),
				PublishedAtSource:  result.PublishedAtSource,
				SearchSnippet:      strings.TrimSpace(result.SearchSnippet),
				FullTextAuthorized: result.FullTextAuthorized && strings.TrimSpace(result.BodyContent) != "",
				BodyContent:        strings.TrimSpace(result.BodyContent),
				BodyFetchStatus:    normalizeArticleBodyFetchStatus(result.BodyFetchStatus, result.BodyContent),
				BodyFetchError:     strings.TrimSpace(result.BodyFetchError),
			}
			candidate.CanonicalURLHash = ArticleOriginalURLHash(normalized)
			if strings.Contains(normalized, "weixin.sogou.com/link") {
				candidate.CanonicalURLHash = stableSogouArticleHash(candidate.Title, candidate.SourceName, candidate.PublishedAtSource, normalized)
			}
			quality := ScoreArticleCandidate(ArticleQualityInput{
				Candidate: candidate,
				Keyword:   kw.Keyword,
				Rank:      result.Rank,
				Config:    qualityConfig,
			})
			candidate.QualityScore = quality.Score
			candidate.QualityReasons = quality.Reasons
			if quality.Skip {
				_ = s.recordItem(ctx, "", normalized, kw.Keyword, model.CollectionItemStatusSkipped, quality.SkipReason, quality)
				continue
			}
			articleID, inserted, insertErr := s.store.InsertCandidate(ctx, candidate)
			if insertErr != nil {
				task.FailedCount++
				if firstErr == nil {
					firstErr = insertErr
				}
				_ = s.recordItem(ctx, "", normalized, kw.Keyword, model.CollectionItemStatusFailed, insertErr.Error(), quality)
				continue
			}
			if inserted {
				task.InsertedCount++
				if s.shouldAutoPublish(candidate, quality, collectionConfig, autoPublishedCount) {
					quality.AutoPublishReason = autoPublishReason(quality.Score, collectionConfig.AutoPublishMinQualityScore)
					if autoPublishStore, ok := s.store.(ArticleAutoPublishStore); ok {
						if err := autoPublishStore.AutoPublishArticle(ctx, articleID, quality.Score, collectionConfig.AutoPublishMinQualityScore); err != nil {
							task.FailedCount++
							if firstErr == nil {
								firstErr = err
							}
							_ = s.recordItem(ctx, articleID, normalized, kw.Keyword, model.CollectionItemStatusFailed, err.Error(), quality)
							continue
						}
						quality.AutoPublished = true
						autoPublishedCount++
					}
				}
				_ = s.recordItem(ctx, articleID, normalized, kw.Keyword, model.CollectionItemStatusInserted, articleBodyFetchItemMessage(candidate), quality)
			} else {
				task.DuplicateCount++
				_ = s.recordItem(ctx, articleID, normalized, kw.Keyword, model.CollectionItemStatusDuplicate, articleBodyFetchItemMessage(candidate), quality)
			}
		}
	}
	if task.FailedCount > 0 && firstErr != nil {
		task.Status = model.CollectionTaskStatusFailed
	}
	now := time.Now()
	task.FinishedAt = &now
	return task, firstErr
}

func (s *ArticleCollectionService) recordItem(ctx context.Context, articleID, originalURL, keyword string, status model.CollectionItemStatus, errorMsg string, quality ArticleQualityResult) error {
	if qualityStore, ok := s.store.(ArticleCollectionQualityItemStore); ok {
		return qualityStore.RecordItemWithQuality(ctx, articleID, originalURL, keyword, status, errorMsg, quality)
	}
	return s.store.RecordItem(ctx, articleID, originalURL, keyword, status, errorMsg)
}

func (s *ArticleCollectionService) shouldAutoPublish(candidate ArticleCollectionCandidate, quality ArticleQualityResult, cfg model.ArticleCollectionConfig, autoPublishedCount int) bool {
	if !cfg.AutoPublishEnabled || cfg.AutoPublishMaxPerRun <= 0 {
		return false
	}
	if autoPublishedCount >= cfg.AutoPublishMaxPerRun {
		return false
	}
	if quality.Score < cfg.AutoPublishMinQualityScore {
		return false
	}
	if cfg.AutoPublishRequiresBody && (!candidate.FullTextAuthorized || strings.TrimSpace(candidate.BodyContent) == "") {
		return false
	}
	return true
}

func normalizeCollectionAutoPublishConfig(cfg *model.ArticleCollectionConfig) {
	if cfg == nil {
		return
	}
	if cfg.AutoPublishMinQualityScore < 0 {
		cfg.AutoPublishMinQualityScore = 0
	}
	if cfg.AutoPublishMinQualityScore > 100 {
		cfg.AutoPublishMinQualityScore = 100
	}
	if cfg.AutoPublishMaxPerRun < 0 {
		cfg.AutoPublishMaxPerRun = 0
	}
	if cfg.AutoPublishMaxPerRun > 20 {
		cfg.AutoPublishMaxPerRun = 20
	}
}

func autoPublishReason(score, threshold int) string {
	return "定时采集自动发布：质量分 " + intToString(score) + " >= 阈值 " + intToString(threshold)
}

func normalizeArticleBodyFetchStatus(status model.ArticleBodyFetchStatus, bodyContent string) model.ArticleBodyFetchStatus {
	if strings.TrimSpace(bodyContent) != "" {
		return model.ArticleBodyFetchStatusSucceeded
	}
	if status == "" {
		return model.ArticleBodyFetchStatusPending
	}
	return status
}

func articleBodyFetchItemMessage(candidate ArticleCollectionCandidate) string {
	if candidate.FullTextAuthorized && strings.TrimSpace(candidate.BodyContent) != "" {
		return ""
	}
	switch candidate.BodyFetchStatus {
	case "", model.ArticleBodyFetchStatusPending, model.ArticleBodyFetchStatusSucceeded, model.ArticleBodyFetchStatusManual:
		return ""
	}
	if strings.TrimSpace(candidate.BodyFetchError) != "" {
		return candidate.BodyFetchError
	}
	return string(candidate.BodyFetchStatus)
}
