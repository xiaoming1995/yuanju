package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
	"yuanju/internal/model"
)

func TestNormalizeArticleOriginalURLDropsTrackingParamsAndFragment(t *testing.T) {
	raw := "https://mp.weixin.qq.com/s?__biz=abc&mid=123&idx=1&sn=def&chksm=track&utm_source=x&from=timeline#rd"

	normalized, err := NormalizeArticleOriginalURL(raw)
	if err != nil {
		t.Fatalf("NormalizeArticleOriginalURL unexpected err: %v", err)
	}

	want := "https://mp.weixin.qq.com/s?__biz=abc&idx=1&mid=123&sn=def"
	if normalized != want {
		t.Fatalf("normalized URL=%q, want %q", normalized, want)
	}
	if ArticleOriginalURLHash(normalized) != ArticleOriginalURLHash(raw) {
		t.Fatalf("hash should be stable across normalized and raw original URL")
	}
}

func TestArticleStatusTransitions(t *testing.T) {
	tests := []struct {
		name string
		from model.ArticleStatus
		to   model.ArticleStatus
		ok   bool
	}{
		{name: "candidate can publish", from: model.ArticleStatusCandidate, to: model.ArticleStatusPublished, ok: true},
		{name: "published can takedown", from: model.ArticleStatusPublished, to: model.ArticleStatusTakenDown, ok: true},
		{name: "taken down can republish", from: model.ArticleStatusTakenDown, to: model.ArticleStatusPublished, ok: true},
		{name: "deleted cannot publish", from: model.ArticleStatusDeleted, to: model.ArticleStatusPublished, ok: false},
		{name: "candidate cannot takedown", from: model.ArticleStatusCandidate, to: model.ArticleStatusTakenDown, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArticleStatusTransition(tt.from, tt.to)
			if tt.ok && err != nil {
				t.Fatalf("transition %s -> %s should be allowed: %v", tt.from, tt.to, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("transition %s -> %s should be rejected", tt.from, tt.to)
			}
		})
	}
}

func TestBuildArticleAnalysisPromptIncludesOnlyAuthorizedBodyContent(t *testing.T) {
	publishedAt := time.Date(2026, 7, 1, 8, 30, 0, 0, time.UTC)
	article := model.Article{
		Title:             "八字格局怎么入门",
		SourceName:        "命理参考",
		OriginalURL:       "https://mp.weixin.qq.com/s?__biz=abc&mid=123&idx=1&sn=def",
		PublishedAtSource: &publishedAt,
		SearchSnippet:     "公开搜索摘要：从十神和月令理解格局。",
		Summary:           "平台摘要：适合初学者建立结构化拆解思路。",
		BodyContent:       "授权采集的原文正文",
	}
	input := ArticleAnalysisPromptInput{
		Article:      article,
		CategoryName: "八字入门",
		TagNames:     []string{"格局", "十神"},
	}

	prompt := BuildArticleAnalysisPrompt(input)
	for _, want := range []string{"八字格局怎么入门", "命理参考", "公开搜索摘要", "平台摘要", "八字入门", "格局", "十神"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt should contain %q, got:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, article.BodyContent) {
		t.Fatalf("prompt must not include unauthorized article body content, got:\n%s", prompt)
	}

	input.Article.FullTextAuthorized = true
	prompt = BuildArticleAnalysisPrompt(input)
	if !strings.Contains(prompt, article.BodyContent) {
		t.Fatalf("prompt should include authorized body content, got:\n%s", prompt)
	}
}

func TestArticleAIUsesArticleSpecificProviderLookup(t *testing.T) {
	src, err := os.ReadFile("article_admin_service.go")
	if err != nil {
		t.Fatalf("read article admin service: %v", err)
	}
	body := string(src)
	if !strings.Contains(body, "repository.GetActiveArticleAIProvider") {
		t.Fatalf("article AI generation must use article-specific provider lookup")
	}
	if strings.Contains(body, "repository.GetActiveLLMProvider") {
		t.Fatalf("article AI generation must not implicitly use active Bazi report provider")
	}
}

func TestArticleCollectionSkipsInactiveKeywordsAndDedupesByURL(t *testing.T) {
	store := newFakeArticleCollectionStore([]model.ArticleKeyword{
		{ID: "kw-active", Keyword: "八字格局", Active: true},
		{ID: "kw-inactive", Keyword: "紫微", Active: false},
	})
	provider := &fakeArticleSearchProvider{
		results: map[string][]ArticleSearchResult{
			"八字格局": {
				{Title: "格局入门", SourceName: "命理参考", OriginalURL: "https://mp.weixin.qq.com/s?__biz=abc&mid=1&idx=1&sn=a&utm_source=x"},
				{Title: "格局入门重复", SourceName: "命理参考", OriginalURL: "https://mp.weixin.qq.com/s?__biz=abc&mid=1&idx=1&sn=a#from"},
			},
		},
	}
	svc := NewArticleCollectionService(provider, store)

	task, err := svc.RunManual(context.Background(), 10)
	if err != nil {
		t.Fatalf("RunManual unexpected err: %v", err)
	}
	if task.KeywordCount != 1 || task.FoundCount != 2 || task.InsertedCount != 1 || task.DuplicateCount != 1 || task.FailedCount != 0 {
		t.Fatalf("unexpected task counts: %+v", task)
	}
	if len(provider.queries) != 1 || provider.queries[0] != "八字格局" {
		t.Fatalf("provider queries=%v, want only active keyword", provider.queries)
	}
	if len(store.items) != 2 {
		t.Fatalf("task items=%d, want 2", len(store.items))
	}
}

func TestArticleCollectionRecordsProviderFailure(t *testing.T) {
	store := newFakeArticleCollectionStore([]model.ArticleKeyword{{ID: "kw", Keyword: "八字", Active: true}})
	provider := &fakeArticleSearchProvider{err: errors.New("provider blocked")}
	svc := NewArticleCollectionService(provider, store)

	task, err := svc.RunManual(context.Background(), 5)
	if err == nil {
		t.Fatalf("RunManual should return provider failure")
	}
	if task.Status != model.CollectionTaskStatusFailed || task.FailedCount != 1 || !strings.Contains(task.ErrorMsg, "provider blocked") {
		t.Fatalf("unexpected failed task: %+v", task)
	}
	if len(store.items) != 1 || store.items[0].Status != model.CollectionItemStatusFailed {
		t.Fatalf("failure item not recorded: %+v", store.items)
	}
}

func TestArticleCollectionMaxResultsAppliesAcrossKeywords(t *testing.T) {
	store := newFakeArticleCollectionStore([]model.ArticleKeyword{
		{ID: "kw-1", Keyword: "八字", Active: true},
		{ID: "kw-2", Keyword: "桃花运", Active: true},
		{ID: "kw-3", Keyword: "生肖运势", Active: true},
	})
	provider := &fakeArticleSearchProvider{results: map[string][]ArticleSearchResult{}}
	for _, keyword := range []string{"八字", "桃花运", "生肖运势"} {
		for i := 0; i < 10; i++ {
			provider.results[keyword] = append(provider.results[keyword], ArticleSearchResult{
				Title:       keyword + "文章" + string(rune('A'+i)),
				SourceName:  "命理参考",
				OriginalURL: "https://mp.weixin.qq.com/s?__biz=abc&mid=" + string(rune('1'+i)) + "&idx=1&sn=" + keyword,
			})
		}
	}
	svc := NewArticleCollectionService(provider, store)

	task, err := svc.RunManual(context.Background(), 5)
	if err != nil {
		t.Fatalf("RunManual unexpected err: %v", err)
	}
	if task.FoundCount != 5 || task.InsertedCount != 5 || len(store.items) != 5 {
		t.Fatalf("task should cap total collected results at 5: task=%+v items=%d", task, len(store.items))
	}
	if strings.Join(provider.queries, ",") != "八字,桃花运,生肖运势" {
		t.Fatalf("provider queries=%v, want all active keywords", provider.queries)
	}
	if got := strings.Join(provider.limits, ","); got != "2,2,1" {
		t.Fatalf("provider limits=%s, want fair total budget 2,2,1", got)
	}
}

func TestArticleCollectionRetryUsesFailedKeywords(t *testing.T) {
	store := newFakeArticleCollectionStore([]model.ArticleKeyword{
		{ID: "kw-active", Keyword: "八字", Active: true},
		{ID: "kw-other", Keyword: "紫微", Active: true},
	})
	store.retryKeywords = []string{"八字"}
	provider := &fakeArticleSearchProvider{
		results: map[string][]ArticleSearchResult{
			"八字": {{Title: "八字重试文章", SourceName: "命理参考", OriginalURL: "https://mp.weixin.qq.com/s?__biz=abc&mid=2&idx=1&sn=b"}},
			"紫微": {{Title: "不应采集", SourceName: "命理参考", OriginalURL: "https://mp.weixin.qq.com/s?__biz=abc&mid=3&idx=1&sn=c"}},
		},
	}
	svc := NewArticleCollectionService(provider, store)

	task, err := svc.RetryFailedKeywords(context.Background(), "old-task", 10)
	if err != nil {
		t.Fatalf("RetryFailedKeywords unexpected err: %v", err)
	}
	if task.KeywordCount != 1 || task.InsertedCount != 1 {
		t.Fatalf("unexpected retry task: %+v", task)
	}
	if len(provider.queries) != 1 || provider.queries[0] != "八字" {
		t.Fatalf("provider queries=%v, want only failed keyword", provider.queries)
	}
}

func TestArticleCollectionQualityFilterSkipsLowScoreItems(t *testing.T) {
	store := newFakeArticleCollectionStore([]model.ArticleKeyword{{ID: "kw", Keyword: "八字", Active: true}})
	store.qualityConfig = &model.ArticleQualityConfig{Enabled: true, MinQualityScore: 80, AllowWithoutBody: false}
	provider := &fakeArticleSearchProvider{
		results: map[string][]ArticleSearchResult{
			"八字": {
				{Rank: 1, Title: "八字短讯", SourceName: "命理参考", OriginalURL: "https://mp.weixin.qq.com/s?__biz=abc&mid=8&idx=1&sn=short", SearchSnippet: "八字"},
				{Rank: 2, Title: "八字完整解析", SourceName: "命理参考", OriginalURL: "https://mp.weixin.qq.com/s?__biz=abc&mid=9&idx=1&sn=full", SearchSnippet: "八字", BodyContent: strings.Repeat("八字内容", 320), FullTextAuthorized: true},
			},
		},
	}
	svc := NewArticleCollectionService(provider, store)

	task, err := svc.RunManual(context.Background(), 10)
	if err != nil {
		t.Fatalf("RunManual unexpected err: %v", err)
	}
	if task.FoundCount != 2 || task.InsertedCount != 1 {
		t.Fatalf("quality filter should only insert one item: %+v", task)
	}
	if len(store.items) != 2 || store.items[0].Status != model.CollectionItemStatusSkipped || store.items[1].Status != model.CollectionItemStatusInserted {
		t.Fatalf("unexpected quality task items: %+v", store.items)
	}
	if store.items[0].SkipReason == "" || store.items[1].QualityScore < 80 {
		t.Fatalf("quality metadata was not recorded: %+v", store.items)
	}
}

func TestArticleQualityScoringAppliesSourceRulesAndDoesNotRequireMetrics(t *testing.T) {
	now := time.Date(2026, 7, 15, 8, 0, 0, 0, time.UTC)
	published := now.Add(-24 * time.Hour)
	input := ArticleQualityInput{
		Candidate: ArticleCollectionCandidate{
			Rank:              2,
			Title:             "八字财运完整解析",
			SourceName:        "命理参考",
			OriginalURL:       "https://mp.weixin.qq.com/s?__biz=abc&mid=10&idx=1&sn=ok",
			PublishedAtSource: &published,
			SearchSnippet:     "八字财运",
			BodyContent:       strings.Repeat("八字财运内容", 260),
			BodyFetchStatus:   model.ArticleBodyFetchStatusSucceeded,
		},
		Keyword: "八字",
		Rank:    2,
		Config: model.ArticleQualityConfig{
			Enabled:          true,
			MinQualityScore:  70,
			AllowWithoutBody: true,
			BonusKeywords:    []string{"财运"},
			PreferredSources: []string{"命理参考"},
		},
		Now: now,
	}
	got := ScoreArticleCandidate(input)
	if got.Skip || got.Score < 80 {
		t.Fatalf("preferred source should produce a usable high score without external metrics: %+v", got)
	}

	input.Config.SourceBlacklist = []string{"命理参考"}
	got = ScoreArticleCandidate(input)
	if !got.Skip || got.SkipReason == "" || got.Score != 0 {
		t.Fatalf("blacklisted source should be hard-skipped: %+v", got)
	}
}

func TestArticleCollectionAutoPublishDisabledLeavesCandidate(t *testing.T) {
	store := newFakeArticleCollectionStore([]model.ArticleKeyword{{ID: "kw", Keyword: "八字", Active: true}})
	store.collectionConfig = &model.ArticleCollectionConfig{AutoPublishEnabled: false, AutoPublishMinQualityScore: 70, AutoPublishRequiresBody: true, AutoPublishMaxPerRun: 3}
	provider := &fakeArticleSearchProvider{results: map[string][]ArticleSearchResult{
		"八字": {highQualityArticleResult("1")},
	}}
	svc := NewArticleCollectionService(provider, store)

	task, err := svc.RunManual(context.Background(), 10)
	if err != nil {
		t.Fatalf("RunManual unexpected err: %v", err)
	}
	if task.InsertedCount != 1 || len(store.autoPublished) != 0 || store.items[0].AutoPublished {
		t.Fatalf("auto publish should stay disabled: task=%+v items=%+v auto=%+v", task, store.items, store.autoPublished)
	}
}

func TestArticleCollectionAutoPublishesHighScoreWithBody(t *testing.T) {
	store := newFakeArticleCollectionStore([]model.ArticleKeyword{{ID: "kw", Keyword: "八字", Active: true}})
	store.collectionConfig = &model.ArticleCollectionConfig{AutoPublishEnabled: true, AutoPublishMinQualityScore: 70, AutoPublishRequiresBody: true, AutoPublishMaxPerRun: 3}
	provider := &fakeArticleSearchProvider{results: map[string][]ArticleSearchResult{
		"八字": {highQualityArticleResult("1")},
	}}
	svc := NewArticleCollectionService(provider, store)

	task, err := svc.RunManual(context.Background(), 10)
	if err != nil {
		t.Fatalf("RunManual unexpected err: %v", err)
	}
	if task.InsertedCount != 1 || len(store.autoPublished) != 1 || !store.items[0].AutoPublished || store.items[0].AutoPublishReason == "" {
		t.Fatalf("high score with body should auto publish: task=%+v items=%+v auto=%+v", task, store.items, store.autoPublished)
	}
}

func TestArticleCollectionAutoPublishRequiresBody(t *testing.T) {
	store := newFakeArticleCollectionStore([]model.ArticleKeyword{{ID: "kw", Keyword: "八字", Active: true}})
	store.collectionConfig = &model.ArticleCollectionConfig{AutoPublishEnabled: true, AutoPublishMinQualityScore: 40, AutoPublishRequiresBody: true, AutoPublishMaxPerRun: 3}
	result := highQualityArticleResult("1")
	result.BodyContent = ""
	result.FullTextAuthorized = false
	result.BodyFetchStatus = model.ArticleBodyFetchStatusWechatAntispider
	provider := &fakeArticleSearchProvider{results: map[string][]ArticleSearchResult{
		"八字": {result},
	}}
	svc := NewArticleCollectionService(provider, store)

	task, err := svc.RunManual(context.Background(), 10)
	if err != nil {
		t.Fatalf("RunManual unexpected err: %v", err)
	}
	if task.InsertedCount != 1 || len(store.autoPublished) != 0 || store.items[0].AutoPublished {
		t.Fatalf("missing body should block auto publish: task=%+v items=%+v auto=%+v", task, store.items, store.autoPublished)
	}
}

func TestArticleCollectionAutoPublishPerRunCapAndDuplicate(t *testing.T) {
	store := newFakeArticleCollectionStore([]model.ArticleKeyword{{ID: "kw", Keyword: "八字", Active: true}})
	store.collectionConfig = &model.ArticleCollectionConfig{AutoPublishEnabled: true, AutoPublishMinQualityScore: 70, AutoPublishRequiresBody: true, AutoPublishMaxPerRun: 1}
	duplicate := highQualityArticleResult("dup")
	provider := &fakeArticleSearchProvider{results: map[string][]ArticleSearchResult{
		"八字": {
			highQualityArticleResult("1"),
			highQualityArticleResult("2"),
			duplicate,
			duplicate,
		},
	}}
	svc := NewArticleCollectionService(provider, store)

	task, err := svc.RunManual(context.Background(), 10)
	if err != nil {
		t.Fatalf("RunManual unexpected err: %v", err)
	}
	if task.InsertedCount != 3 || task.DuplicateCount != 1 || len(store.autoPublished) != 1 {
		t.Fatalf("auto publish cap/duplicate behavior mismatch: task=%+v items=%+v auto=%+v", task, store.items, store.autoPublished)
	}
	if !store.items[0].AutoPublished || store.items[1].AutoPublished || store.items[3].AutoPublished {
		t.Fatalf("only first inserted item should auto publish: %+v", store.items)
	}
}

func highQualityArticleResult(id string) ArticleSearchResult {
	published := time.Date(2026, 7, 15, 8, 0, 0, 0, time.UTC)
	return ArticleSearchResult{
		Rank:               1,
		Title:              "八字财运完整解析" + id,
		SourceName:         "命理参考",
		OriginalURL:        "https://mp.weixin.qq.com/s?__biz=abc&mid=" + id + "&idx=1&sn=ok",
		PublishedAtSource:  &published,
		SearchSnippet:      "八字财运",
		BodyContent:        strings.Repeat("八字财运内容", 260),
		FullTextAuthorized: true,
		BodyFetchStatus:    model.ArticleBodyFetchStatusSucceeded,
	}
}

type fakeArticleSearchProvider struct {
	results map[string][]ArticleSearchResult
	err     error
	queries []string
	limits  []string
}

func (p *fakeArticleSearchProvider) Search(_ context.Context, keyword string, limit int) ([]ArticleSearchResult, error) {
	p.queries = append(p.queries, keyword)
	p.limits = append(p.limits, string(rune('0'+limit)))
	if p.err != nil {
		return nil, p.err
	}
	results := append([]ArticleSearchResult(nil), p.results[keyword]...)
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

type fakeCollectionItem struct {
	ArticleID         string
	OriginalURL       string
	Keyword           string
	Status            model.CollectionItemStatus
	ErrorMsg          string
	QualityScore      int
	QualityReasons    []model.ArticleQualityReason
	SkipReason        string
	AutoPublished     bool
	AutoPublishReason string
}

type fakeArticleCollectionStore struct {
	keywords         []model.ArticleKeyword
	retryKeywords    []string
	qualityConfig    *model.ArticleQualityConfig
	collectionConfig *model.ArticleCollectionConfig
	seen             map[string]string
	items            []fakeCollectionItem
	autoPublished    map[string]bool
	nextID           int
}

func newFakeArticleCollectionStore(keywords []model.ArticleKeyword) *fakeArticleCollectionStore {
	return &fakeArticleCollectionStore{keywords: keywords, seen: map[string]string{}, autoPublished: map[string]bool{}}
}

func (s *fakeArticleCollectionStore) ListKeywords(_ context.Context) ([]model.ArticleKeyword, error) {
	return s.keywords, nil
}

func (s *fakeArticleCollectionStore) InsertCandidate(_ context.Context, input ArticleCollectionCandidate) (string, bool, error) {
	if id, ok := s.seen[input.CanonicalURLHash]; ok {
		return id, false, nil
	}
	s.nextID++
	id := "article-id"
	if s.nextID > 1 {
		id = id + "-" + string(rune('0'+s.nextID))
	}
	s.seen[input.CanonicalURLHash] = id
	return id, true, nil
}

func (s *fakeArticleCollectionStore) RecordItem(_ context.Context, articleID, originalURL, keyword string, status model.CollectionItemStatus, errorMsg string) error {
	s.items = append(s.items, fakeCollectionItem{ArticleID: articleID, OriginalURL: originalURL, Keyword: keyword, Status: status, ErrorMsg: errorMsg})
	return nil
}

func (s *fakeArticleCollectionStore) RecordItemWithQuality(_ context.Context, articleID, originalURL, keyword string, status model.CollectionItemStatus, errorMsg string, quality ArticleQualityResult) error {
	s.items = append(s.items, fakeCollectionItem{
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
	return nil
}

func (s *fakeArticleCollectionStore) GetQualityConfig(_ context.Context) (*model.ArticleQualityConfig, error) {
	return s.qualityConfig, nil
}

func (s *fakeArticleCollectionStore) GetCollectionConfig(_ context.Context) (*model.ArticleCollectionConfig, error) {
	return s.collectionConfig, nil
}

func (s *fakeArticleCollectionStore) AutoPublishArticle(_ context.Context, articleID string, _, _ int) error {
	s.autoPublished[articleID] = true
	return nil
}

func (s *fakeArticleCollectionStore) ListFailedKeywords(_ context.Context, _ string) ([]string, error) {
	return s.retryKeywords, nil
}
