package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"
	"yuanju/internal/model"
	"yuanju/pkg/database"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func withArticleRepositoryDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()
	pg, err := tcpg.Run(ctx,
		"postgres:16-alpine",
		tcpg.WithDatabase("yuanju_article_test"),
		tcpg.WithUsername("test"),
		tcpg.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	orig := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = orig })

	if _, err := database.Migrate(database.ModeApply); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestInsertArticleCandidateDedupesByCanonicalURLHash(t *testing.T) {
	withArticleRepositoryDB(t)

	input := ArticleCandidateInput{
		Title:            "八字入门参考",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=123&idx=1&sn=def&utm_source=x#rd",
		CanonicalURLHash: "dedupe-hash",
		SearchSnippet:    "公开搜索摘要",
	}
	first, inserted, err := InsertArticleCandidate(input)
	if err != nil {
		t.Fatalf("first insert: %v", err)
	}
	if !inserted {
		t.Fatalf("first insert should create row")
	}
	second, inserted, err := InsertArticleCandidate(input)
	if err != nil {
		t.Fatalf("second insert: %v", err)
	}
	if inserted {
		t.Fatalf("second insert should be deduped")
	}
	if second.ID != first.ID {
		t.Fatalf("deduped article id=%s, want %s", second.ID, first.ID)
	}
}

func TestInsertArticleCandidateStoresAndBackfillsBodyContent(t *testing.T) {
	withArticleRepositoryDB(t)

	withBody, inserted, err := InsertArticleCandidate(ArticleCandidateInput{
		Title:              "带正文文章",
		SourceName:         "命理参考",
		OriginalURL:        "https://mp.weixin.qq.com/s?__biz=abc&mid=223&idx=1&sn=def",
		CanonicalURLHash:   "body-hash",
		FullTextAuthorized: true,
		BodyContent:        "完整正文",
	})
	if err != nil {
		t.Fatalf("insert with body: %v", err)
	}
	if !inserted || !withBody.FullTextAuthorized || withBody.BodyContent != "完整正文" {
		t.Fatalf("article body not stored: inserted=%v article=%+v", inserted, withBody)
	}
	if withBody.BodyFetchStatus != model.ArticleBodyFetchStatusSucceeded || withBody.BodyFetchError != "" {
		t.Fatalf("unexpected body fetch diagnostics: %+v", withBody)
	}

	withoutBody, _, err := InsertArticleCandidate(ArticleCandidateInput{
		Title:            "待补正文文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=224&idx=1&sn=def",
		CanonicalURLHash: "body-backfill-hash",
	})
	if err != nil {
		t.Fatalf("insert without body: %v", err)
	}
	backfilled, inserted, err := InsertArticleCandidate(ArticleCandidateInput{
		Title:              "待补正文文章",
		SourceName:         "命理参考",
		OriginalURL:        withoutBody.OriginalURL,
		CanonicalURLHash:   "body-backfill-hash",
		FullTextAuthorized: true,
		BodyContent:        "后续采集正文",
	})
	if err != nil {
		t.Fatalf("backfill body: %v", err)
	}
	if inserted || backfilled.ID != withoutBody.ID || backfilled.BodyContent != "后续采集正文" || !backfilled.FullTextAuthorized {
		t.Fatalf("unexpected backfilled article inserted=%v article=%+v", inserted, backfilled)
	}
	if backfilled.BodyFetchStatus != model.ArticleBodyFetchStatusSucceeded || backfilled.BodyFetchError != "" {
		t.Fatalf("unexpected backfilled body fetch diagnostics: %+v", backfilled)
	}
}

func TestUpdateArticleStatusWithAuditWritesAuditEvent(t *testing.T) {
	db := withArticleRepositoryDB(t)
	var adminID string
	if err := db.QueryRow(`INSERT INTO admins (email, password_hash, name) VALUES ('article-admin@example.com', 'hash', 'Article Admin') RETURNING id`).Scan(&adminID); err != nil {
		t.Fatalf("insert admin: %v", err)
	}
	article, _, err := InsertArticleCandidate(ArticleCandidateInput{
		Title:            "候选文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=456&idx=1&sn=def",
		CanonicalURLHash: "status-hash",
	})
	if err != nil {
		t.Fatalf("insert article: %v", err)
	}

	updated, err := UpdateArticleStatusWithAudit(article.ID, adminID, model.ArticleStatusPublished, "通过审核")
	if err != nil {
		t.Fatalf("publish article: %v", err)
	}
	if updated.Status != model.ArticleStatusPublished || updated.PublishedAt == nil {
		t.Fatalf("updated status=%s published_at=%v", updated.Status, updated.PublishedAt)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM article_audit_events WHERE article_id=$1 AND admin_id=$2 AND action='publish'`, article.ID, adminID).Scan(&count); err != nil {
		t.Fatalf("count audit: %v", err)
	}
	if count != 1 {
		t.Fatalf("audit count=%d, want 1", count)
	}
}

func TestAutoPublishArticleWithAuditWritesSystemAuditEvent(t *testing.T) {
	db := withArticleRepositoryDB(t)
	article, _, err := InsertArticleCandidate(ArticleCandidateInput{
		Title:            "自动发布文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=457&idx=1&sn=def",
		CanonicalURLHash: "auto-publish-hash",
	})
	if err != nil {
		t.Fatalf("insert article: %v", err)
	}

	updated, err := AutoPublishArticleWithAudit(article.ID, 92, 90)
	if err != nil {
		t.Fatalf("auto publish article: %v", err)
	}
	if updated.Status != model.ArticleStatusPublished || updated.PublishedAt == nil {
		t.Fatalf("updated status=%s published_at=%v", updated.Status, updated.PublishedAt)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM article_audit_events WHERE article_id=$1 AND admin_id IS NULL AND action='auto_publish' AND note LIKE '%质量分 92%'`, article.ID).Scan(&count); err != nil {
		t.Fatalf("count auto publish audit: %v", err)
	}
	if count != 1 {
		t.Fatalf("auto publish audit count=%d, want 1", count)
	}
}

func TestRecordArticleOriginalClickRequiresPublishedArticle(t *testing.T) {
	db := withArticleRepositoryDB(t)
	var userID string
	if err := db.QueryRow(`INSERT INTO users (email, password_hash) VALUES ('article-user@example.com', 'hash') RETURNING id`).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	hidden, _, err := InsertArticleCandidate(ArticleCandidateInput{
		Title:            "隐藏文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=789&idx=1&sn=def",
		CanonicalURLHash: "hidden-click-hash",
	})
	if err != nil {
		t.Fatalf("insert hidden article: %v", err)
	}
	if err := RecordArticleOriginalClick(hidden.ID, userID); err == nil {
		t.Fatalf("clicking non-published article should fail")
	}

	published, _, err := InsertArticleCandidate(ArticleCandidateInput{
		Title:            "发布文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=790&idx=1&sn=def",
		CanonicalURLHash: "published-click-hash",
	})
	if err != nil {
		t.Fatalf("insert published article: %v", err)
	}
	if _, err := db.Exec(`UPDATE articles SET status='published' WHERE id=$1`, published.ID); err != nil {
		t.Fatalf("publish fixture: %v", err)
	}
	if err := RecordArticleOriginalClick(published.ID, userID); err != nil {
		t.Fatalf("click published article: %v", err)
	}
	var clicks int
	if err := db.QueryRow(`SELECT original_click_count FROM articles WHERE id=$1`, published.ID).Scan(&clicks); err != nil {
		t.Fatalf("read click count: %v", err)
	}
	if clicks != 1 {
		t.Fatalf("click count=%d, want 1", clicks)
	}
}

func TestListArticlesFiltersByTagAndSortsHot(t *testing.T) {
	db := withArticleRepositoryDB(t)
	tag, err := CreateArticleTag("格局", "geju", true)
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}
	hot, _, err := InsertArticleCandidate(ArticleCandidateInput{
		Title:            "热门格局文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=800&idx=1&sn=def",
		CanonicalURLHash: "hot-filter-hash",
		SearchSnippet:    "格局摘要",
	})
	if err != nil {
		t.Fatalf("insert hot: %v", err)
	}
	cold, _, err := InsertArticleCandidate(ArticleCandidateInput{
		Title:            "冷门格局文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=801&idx=1&sn=def",
		CanonicalURLHash: "cold-filter-hash",
		SearchSnippet:    "格局摘要",
	})
	if err != nil {
		t.Fatalf("insert cold: %v", err)
	}
	other, _, err := InsertArticleCandidate(ArticleCandidateInput{
		Title:            "无关文章",
		SourceName:       "命理参考",
		OriginalURL:      "https://mp.weixin.qq.com/s?__biz=abc&mid=802&idx=1&sn=def",
		CanonicalURLHash: "other-filter-hash",
		SearchSnippet:    "无关摘要",
	})
	if err != nil {
		t.Fatalf("insert other: %v", err)
	}
	if err := UpdateArticleTags(hot.ID, []string{tag.ID}); err != nil {
		t.Fatalf("tag hot: %v", err)
	}
	if err := UpdateArticleTags(cold.ID, []string{tag.ID}); err != nil {
		t.Fatalf("tag cold: %v", err)
	}
	if _, err := db.Exec(`UPDATE articles SET status='published', view_count=9 WHERE id=$1`, hot.ID); err != nil {
		t.Fatalf("publish hot: %v", err)
	}
	if _, err := db.Exec(`UPDATE articles SET status='published', view_count=2 WHERE id=$1`, cold.ID); err != nil {
		t.Fatalf("publish cold: %v", err)
	}
	if _, err := db.Exec(`UPDATE articles SET status='published', view_count=99 WHERE id=$1`, other.ID); err != nil {
		t.Fatalf("publish other: %v", err)
	}

	articles, total, err := ListArticles(ArticleListFilter{
		Status: "published",
		TagID:  tag.ID,
		Sort:   "hot",
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("list articles: %v", err)
	}
	if total != 2 || len(articles) != 2 {
		t.Fatalf("total=%d len=%d, want 2", total, len(articles))
	}
	if articles[0].ID != hot.ID || articles[1].ID != cold.ID {
		t.Fatalf("articles not sorted by hot within tag: got %s then %s", articles[0].Title, articles[1].Title)
	}
}

func TestListFailedArticleCollectionTaskKeywords(t *testing.T) {
	withArticleRepositoryDB(t)
	task, err := CreateArticleCollectionTask("manual", 2)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if err := AddArticleCollectionTaskItem(task.ID, "", "", "八字", model.CollectionItemStatusFailed, "blocked"); err != nil {
		t.Fatalf("insert failed item: %v", err)
	}
	if err := AddArticleCollectionTaskItem(task.ID, "", "", "八字", model.CollectionItemStatusFailed, "blocked again"); err != nil {
		t.Fatalf("insert duplicate failed item: %v", err)
	}
	if err := AddArticleCollectionTaskItem(task.ID, "", "", "紫微", model.CollectionItemStatusInserted, ""); err != nil {
		t.Fatalf("insert non-failed item: %v", err)
	}

	keywords, err := ListFailedArticleCollectionTaskKeywords(task.ID)
	if err != nil {
		t.Fatalf("list failed keywords: %v", err)
	}
	if len(keywords) != 1 || keywords[0] != "八字" {
		t.Fatalf("keywords=%v, want [八字]", keywords)
	}
}
