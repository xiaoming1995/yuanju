package service

import (
	"context"
	"errors"
	"yuanju/internal/model"
	"yuanju/internal/repository"
)

var ErrArticleModuleDisabled = errors.New("资讯模块已关闭")

type RepositoryArticleCollectionStore struct {
	taskID string
}

func NewRepositoryArticleCollectionStore(taskID string) *RepositoryArticleCollectionStore {
	return &RepositoryArticleCollectionStore{taskID: taskID}
}

func (s *RepositoryArticleCollectionStore) ListKeywords(_ context.Context) ([]model.ArticleKeyword, error) {
	return repository.ListArticleKeywords(true)
}

func (s *RepositoryArticleCollectionStore) InsertCandidate(_ context.Context, input ArticleCollectionCandidate) (string, bool, error) {
	article, inserted, err := repository.InsertArticleCandidate(repository.ArticleCandidateInput{
		Title:              input.Title,
		SourceName:         input.SourceName,
		OriginalURL:        input.OriginalURL,
		CanonicalURLHash:   input.CanonicalURLHash,
		CoverURL:           input.CoverURL,
		PublishedAtSource:  input.PublishedAtSource,
		SearchSnippet:      input.SearchSnippet,
		QualityScore:       input.QualityScore,
		QualityReasons:     input.QualityReasons,
		FullTextAuthorized: input.FullTextAuthorized,
		BodyContent:        input.BodyContent,
		BodyFetchStatus:    input.BodyFetchStatus,
		BodyFetchError:     input.BodyFetchError,
	})
	if err != nil || article == nil {
		return "", inserted, err
	}
	return article.ID, inserted, nil
}

func (s *RepositoryArticleCollectionStore) RecordItem(_ context.Context, articleID, originalURL, keyword string, status model.CollectionItemStatus, errorMsg string) error {
	return repository.AddArticleCollectionTaskItem(s.taskID, articleID, originalURL, keyword, status, errorMsg)
}

func (s *RepositoryArticleCollectionStore) RecordItemWithQuality(_ context.Context, articleID, originalURL, keyword string, status model.CollectionItemStatus, errorMsg string, quality ArticleQualityResult) error {
	return repository.AddArticleCollectionTaskItemWithQuality(repository.ArticleCollectionTaskItemInput{
		TaskID:            s.taskID,
		ArticleID:         articleID,
		OriginalURL:       originalURL,
		Keyword:           keyword,
		Status:            status,
		ErrorMsg:          errorMsg,
		QualityScore:      quality.Score,
		QualityReasons:    quality.Reasons,
		SkipReason:        quality.SkipReason,
		AutoPublished:     quality.AutoPublished,
		AutoPublishReason: quality.AutoPublishReason,
	})
}

func (s *RepositoryArticleCollectionStore) GetQualityConfig(_ context.Context) (*model.ArticleQualityConfig, error) {
	return repository.GetArticleQualityConfig()
}

func (s *RepositoryArticleCollectionStore) GetCollectionConfig(_ context.Context) (*model.ArticleCollectionConfig, error) {
	return repository.GetArticleCollectionConfig()
}

func (s *RepositoryArticleCollectionStore) AutoPublishArticle(_ context.Context, articleID string, qualityScore, threshold int) error {
	_, err := repository.AutoPublishArticleWithAudit(articleID, qualityScore, threshold)
	return err
}

func (s *RepositoryArticleCollectionStore) ListFailedKeywords(_ context.Context, taskID string) ([]string, error) {
	return repository.ListFailedArticleCollectionTaskKeywords(taskID)
}

func RunManualArticleCollectionTask(ctx context.Context, provider ArticleSearchProvider, maxResults int) (model.ArticleCollectionTask, error) {
	if err := ensureArticleModuleEnabledForCollection(); err != nil {
		return model.ArticleCollectionTask{}, err
	}
	keywords, err := repository.ListArticleKeywords(true)
	if err != nil {
		return model.ArticleCollectionTask{}, err
	}
	task, err := repository.CreateArticleCollectionTask("manual", len(keywords))
	if err != nil {
		return model.ArticleCollectionTask{}, err
	}
	store := NewRepositoryArticleCollectionStore(task.ID)
	result, runErr := NewArticleCollectionService(provider, store).RunManual(ctx, maxResults)
	result.ID = task.ID
	result.StartedAt = task.StartedAt

	status := result.Status
	if status == "" {
		status = model.CollectionTaskStatusSucceeded
	}
	finishErr := repository.FinishArticleCollectionTask(task.ID, status, repository.ArticleCollectionCounts{
		KeywordCount:   result.KeywordCount,
		FoundCount:     result.FoundCount,
		InsertedCount:  result.InsertedCount,
		DuplicateCount: result.DuplicateCount,
		FailedCount:    result.FailedCount,
	}, result.ErrorMsg)
	if runErr != nil {
		return result, runErr
	}
	return result, finishErr
}

func RetryArticleCollectionTask(ctx context.Context, taskID string, provider ArticleSearchProvider, maxResults int) (model.ArticleCollectionTask, error) {
	if err := ensureArticleModuleEnabledForCollection(); err != nil {
		return model.ArticleCollectionTask{}, err
	}
	task, err := repository.CreateArticleCollectionTask("retry", 0)
	if err != nil {
		return model.ArticleCollectionTask{}, err
	}
	store := NewRepositoryArticleCollectionStore(task.ID)
	result, runErr := NewArticleCollectionService(provider, store).RetryFailedKeywords(ctx, taskID, maxResults)
	result.ID = task.ID
	result.StartedAt = task.StartedAt

	status := result.Status
	if status == "" {
		status = model.CollectionTaskStatusSucceeded
	}
	finishErr := repository.FinishArticleCollectionTask(task.ID, status, repository.ArticleCollectionCounts{
		KeywordCount:   result.KeywordCount,
		FoundCount:     result.FoundCount,
		InsertedCount:  result.InsertedCount,
		DuplicateCount: result.DuplicateCount,
		FailedCount:    result.FailedCount,
	}, result.ErrorMsg)
	if runErr != nil {
		return result, runErr
	}
	return result, finishErr
}

func ensureArticleModuleEnabledForCollection() error {
	enabled, err := repository.IsArticleModuleEnabled()
	if err != nil {
		return err
	}
	if !enabled {
		return ErrArticleModuleDisabled
	}
	return nil
}
