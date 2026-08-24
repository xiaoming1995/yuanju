package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"yuanju/internal/model"
	"yuanju/pkg/database"

	"github.com/lib/pq"
)

type ArticleCandidateInput struct {
	Title              string
	SourceName         string
	OriginalURL        string
	CanonicalURLHash   string
	CoverURL           string
	PublishedAtSource  *time.Time
	SearchSnippet      string
	Summary            string
	CategoryID         *string
	QualityScore       int
	QualityReasons     []model.ArticleQualityReason
	FullTextAuthorized bool
	BodyContent        string
	BodyFetchStatus    model.ArticleBodyFetchStatus
	BodyFetchError     string
}

type ArticleListFilter struct {
	Status          string
	Query           string
	CategoryID      string
	TagID           string
	Sort            string
	MinQualityScore int
	Limit           int
	Offset          int
}

type ArticleBatchResult struct {
	Updated int      `json:"updated"`
	Skipped []string `json:"skipped"`
}

type ArticleCollectionCounts struct {
	KeywordCount   int
	FoundCount     int
	InsertedCount  int
	DuplicateCount int
	FailedCount    int
}

type ArticleCollectionTaskItemInput struct {
	TaskID            string
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

const articleSelectColumns = `
	a.id, a.title, a.source_name, a.original_url, a.canonical_url_hash, a.cover_url,
	a.published_at_source, a.search_snippet, a.summary, a.ai_analysis, a.ai_status,
	a.ai_error_msg, a.category_id, a.status, a.view_count, a.original_click_count,
	a.quality_score, a.quality_reasons, a.full_text_authorized, COALESCE(a.body_content, ''), a.body_fetch_status, COALESCE(a.body_fetch_error, ''), a.created_at, a.updated_at,
	a.published_at, a.taken_down_at, a.deleted_at`

const articleReturningColumns = `
	id, title, source_name, original_url, canonical_url_hash, cover_url,
	published_at_source, search_snippet, summary, ai_analysis, ai_status,
	ai_error_msg, category_id, status, view_count, original_click_count,
	quality_score, quality_reasons, full_text_authorized, COALESCE(body_content, ''), body_fetch_status, COALESCE(body_fetch_error, ''), created_at, updated_at,
	published_at, taken_down_at, deleted_at`

type articleScanner interface {
	Scan(dest ...any) error
}

func CreateArticleCategory(name, slug string, sortOrder int, active bool) (*model.ArticleCategory, error) {
	c := &model.ArticleCategory{}
	err := database.DB.QueryRow(`
		INSERT INTO article_categories (name, slug, sort_order, active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, slug, sort_order, active, created_at, updated_at`,
		name, slug, sortOrder, active,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.SortOrder, &c.Active, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func UpdateArticleCategory(id, name, slug string, sortOrder int, active bool) (*model.ArticleCategory, error) {
	c := &model.ArticleCategory{}
	err := database.DB.QueryRow(`
		UPDATE article_categories
		SET name=$2, slug=$3, sort_order=$4, active=$5, updated_at=NOW()
		WHERE id=$1
		RETURNING id, name, slug, sort_order, active, created_at, updated_at`,
		id, name, slug, sortOrder, active,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.SortOrder, &c.Active, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func ListArticleCategories(activeOnly bool) ([]model.ArticleCategory, error) {
	query := `SELECT id, name, slug, sort_order, active, created_at, updated_at FROM article_categories`
	args := []any{}
	if activeOnly {
		query += ` WHERE active=true`
	}
	query += ` ORDER BY sort_order ASC, created_at DESC`
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ArticleCategory
	for rows.Next() {
		var c model.ArticleCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.SortOrder, &c.Active, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func CreateArticleTag(name, slug string, active bool) (*model.ArticleTag, error) {
	t := &model.ArticleTag{}
	err := database.DB.QueryRow(`
		INSERT INTO article_tags (name, slug, active)
		VALUES ($1, $2, $3)
		RETURNING id, name, slug, active, created_at, updated_at`,
		name, slug, active,
	).Scan(&t.ID, &t.Name, &t.Slug, &t.Active, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func UpdateArticleTag(id, name, slug string, active bool) (*model.ArticleTag, error) {
	t := &model.ArticleTag{}
	err := database.DB.QueryRow(`
		UPDATE article_tags
		SET name=$2, slug=$3, active=$4, updated_at=NOW()
		WHERE id=$1
		RETURNING id, name, slug, active, created_at, updated_at`,
		id, name, slug, active,
	).Scan(&t.ID, &t.Name, &t.Slug, &t.Active, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func ListArticleTags(activeOnly bool) ([]model.ArticleTag, error) {
	query := `SELECT id, name, slug, active, created_at, updated_at FROM article_tags`
	if activeOnly {
		query += ` WHERE active=true`
	}
	query += ` ORDER BY created_at DESC`
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ArticleTag
	for rows.Next() {
		var tag model.ArticleTag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.Active, &tag.CreatedAt, &tag.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, tag)
	}
	return out, rows.Err()
}

func CreateArticleKeyword(keyword string, active bool) (*model.ArticleKeyword, error) {
	k := &model.ArticleKeyword{}
	err := database.DB.QueryRow(`
		INSERT INTO article_keywords (keyword, active)
		VALUES ($1, $2)
		RETURNING id, keyword, active, created_at, updated_at`,
		keyword, active,
	).Scan(&k.ID, &k.Keyword, &k.Active, &k.CreatedAt, &k.UpdatedAt)
	return k, err
}

func UpdateArticleKeyword(id, keyword string, active bool) (*model.ArticleKeyword, error) {
	k := &model.ArticleKeyword{}
	err := database.DB.QueryRow(`
		UPDATE article_keywords
		SET keyword=$2, active=$3, updated_at=NOW()
		WHERE id=$1
		RETURNING id, keyword, active, created_at, updated_at`,
		id, keyword, active,
	).Scan(&k.ID, &k.Keyword, &k.Active, &k.CreatedAt, &k.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return k, err
}

func ListArticleKeywords(activeOnly bool) ([]model.ArticleKeyword, error) {
	query := `SELECT id, keyword, active, created_at, updated_at FROM article_keywords`
	if activeOnly {
		query += ` WHERE active=true`
	}
	query += ` ORDER BY created_at DESC`
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ArticleKeyword
	for rows.Next() {
		var k model.ArticleKeyword
		if err := rows.Scan(&k.ID, &k.Keyword, &k.Active, &k.CreatedAt, &k.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func InsertArticleCandidate(input ArticleCandidateInput) (*model.Article, bool, error) {
	if input.CanonicalURLHash == "" {
		return nil, false, fmt.Errorf("canonical_url_hash 不能为空")
	}
	input.BodyContent = strings.TrimSpace(input.BodyContent)
	input.BodyFetchError = strings.TrimSpace(input.BodyFetchError)
	input.BodyFetchStatus = normalizeRepoArticleBodyFetchStatus(input.BodyFetchStatus, input.BodyContent)
	qualityReasons, err := marshalArticleQualityReasons(input.QualityReasons)
	if err != nil {
		return nil, false, err
	}
	a := &model.Article{}
	err = scanArticle(database.DB.QueryRow(`
		INSERT INTO articles
			(title, source_name, original_url, canonical_url_hash, cover_url, published_at_source, search_snippet, summary, category_id, quality_score, quality_reasons, full_text_authorized, body_content, body_fetch_status, body_fetch_error)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (canonical_url_hash) DO NOTHING
		RETURNING `+articleReturningColumns,
		input.Title, input.SourceName, input.OriginalURL, input.CanonicalURLHash, input.CoverURL,
		input.PublishedAtSource, input.SearchSnippet, input.Summary, input.CategoryID, clampInt(input.QualityScore, 0, 100), qualityReasons, input.FullTextAuthorized, input.BodyContent,
		input.BodyFetchStatus, input.BodyFetchError,
	), a)
	if err == sql.ErrNoRows {
		existing, getErr := GetArticleByURLHash(input.CanonicalURLHash)
		if getErr == nil && existing != nil && input.FullTextAuthorized && strings.TrimSpace(input.BodyContent) != "" && strings.TrimSpace(existing.BodyContent) == "" {
			updated, updateErr := UpdateArticleBodyIfEmpty(existing.ID, input.BodyContent)
			if updateErr != nil {
				return existing, false, updateErr
			}
			if updated != nil {
				existing = updated
			}
		} else if getErr == nil && existing != nil && strings.TrimSpace(existing.BodyContent) == "" && input.BodyFetchStatus != "" && input.BodyFetchStatus != model.ArticleBodyFetchStatusPending {
			updated, updateErr := UpdateArticleBodyFetchStatus(existing.ID, input.BodyFetchStatus, input.BodyFetchError)
			if updateErr != nil {
				return existing, false, updateErr
			}
			if updated != nil {
				existing = updated
			}
		}
		return existing, false, getErr
	}
	if err != nil {
		return nil, false, err
	}
	return a, true, nil
}

func UpdateArticleBodyIfEmpty(articleID, bodyContent string) (*model.Article, error) {
	bodyContent = strings.TrimSpace(bodyContent)
	if bodyContent == "" {
		return GetArticleByID(articleID)
	}
	a := &model.Article{}
	err := scanArticle(database.DB.QueryRow(`
		UPDATE articles
		SET body_content=$2, full_text_authorized=true, body_fetch_status='succeeded', body_fetch_error='', updated_at=NOW()
		WHERE id=$1 AND COALESCE(body_content, '') = ''
		RETURNING `+articleReturningColumns,
		articleID, bodyContent,
	), a)
	if err == sql.ErrNoRows {
		return GetArticleByID(articleID)
	}
	return a, err
}

func UpdateArticleBody(articleID, bodyContent, originalURL string) (*model.Article, error) {
	bodyContent = strings.TrimSpace(bodyContent)
	if bodyContent == "" {
		return GetArticleByID(articleID)
	}
	a := &model.Article{}
	query := `
		UPDATE articles
		SET body_content=$2, full_text_authorized=true, body_fetch_status='manual', body_fetch_error='', updated_at=NOW()
		WHERE id=$1
		RETURNING ` + articleReturningColumns
	args := []any{articleID, bodyContent}
	if strings.TrimSpace(originalURL) != "" {
		query = `
			UPDATE articles
			SET body_content=$2, full_text_authorized=true, original_url=$3, body_fetch_status='succeeded', body_fetch_error='', updated_at=NOW()
			WHERE id=$1
			RETURNING ` + articleReturningColumns
		args = append(args, strings.TrimSpace(originalURL))
	}
	err := scanArticle(database.DB.QueryRow(query, args...), a)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return a, err
}

func UpdateArticleBodyFetchStatus(articleID string, status model.ArticleBodyFetchStatus, errorMsg string) (*model.Article, error) {
	status = normalizeRepoArticleBodyFetchStatus(status, "")
	if status == model.ArticleBodyFetchStatusSucceeded || status == model.ArticleBodyFetchStatusManual {
		status = model.ArticleBodyFetchStatusPending
	}
	a := &model.Article{}
	err := scanArticle(database.DB.QueryRow(`
		UPDATE articles
		SET body_fetch_status=$2, body_fetch_error=$3, updated_at=NOW()
		WHERE id=$1 AND COALESCE(body_content, '') = ''
		RETURNING `+articleReturningColumns,
		articleID, status, strings.TrimSpace(errorMsg),
	), a)
	if err == sql.ErrNoRows {
		return GetArticleByID(articleID)
	}
	return a, err
}

func normalizeRepoArticleBodyFetchStatus(status model.ArticleBodyFetchStatus, bodyContent string) model.ArticleBodyFetchStatus {
	if strings.TrimSpace(bodyContent) != "" {
		return model.ArticleBodyFetchStatusSucceeded
	}
	if status == "" {
		return model.ArticleBodyFetchStatusPending
	}
	return status
}

func GetArticleByURLHash(hash string) (*model.Article, error) {
	return getArticleByQuery(`SELECT `+articleSelectColumns+` FROM articles a WHERE a.canonical_url_hash=$1`, hash)
}

func GetArticleByID(id string) (*model.Article, error) {
	return getArticleByQuery(`SELECT `+articleSelectColumns+` FROM articles a WHERE a.id=$1`, id)
}

func GetPublishedArticleByID(id string) (*model.Article, error) {
	return getArticleByQuery(`SELECT `+articleSelectColumns+` FROM articles a WHERE a.id=$1 AND a.status='published'`, id)
}

func ListArticles(filter ArticleListFilter) ([]model.Article, int, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	where, args := buildArticleListWhere(filter)
	countQuery := `SELECT COUNT(*) FROM articles a ` + where
	var total int
	if err := database.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, filter.Limit, filter.Offset)
	limitParam := len(args) - 1
	offsetParam := len(args)
	query := `SELECT ` + articleSelectColumns + ` FROM articles a ` + where + articleOrderBy(filter.Sort) +
		fmt.Sprintf(` LIMIT $%d OFFSET $%d`, limitParam, offsetParam)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var articles []model.Article
	for rows.Next() {
		var a model.Article
		if err := scanArticle(rows, &a); err != nil {
			return nil, 0, err
		}
		articles = append(articles, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if err := loadTagsForArticles(articles); err != nil {
		return nil, 0, err
	}
	return articles, total, nil
}

func UpdateArticleTags(articleID string, tagIDs []string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM article_tag_links WHERE article_id=$1`, articleID); err != nil {
		return err
	}
	for _, tagID := range tagIDs {
		if strings.TrimSpace(tagID) == "" {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO article_tag_links (article_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, articleID, tagID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func UpdateArticleStatusWithAudit(articleID, adminID string, to model.ArticleStatus, note string) (*model.Article, error) {
	return updateArticleStatusWithAudit(articleID, adminID, to, articleAuditAction(to), note)
}

func AutoPublishArticleWithAudit(articleID string, qualityScore, threshold int) (*model.Article, error) {
	note := fmt.Sprintf("定时采集自动发布：质量分 %d >= 阈值 %d", qualityScore, threshold)
	return updateArticleStatusWithAudit(articleID, "", model.ArticleStatusPublished, "auto_publish", note)
}

func updateArticleStatusWithAudit(articleID, adminID string, to model.ArticleStatus, action, note string) (*model.Article, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var from model.ArticleStatus
	if err := tx.QueryRow(`SELECT status FROM articles WHERE id=$1 FOR UPDATE`, articleID).Scan(&from); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if err := validateArticleRepoTransition(from, to); err != nil {
		return nil, err
	}

	setTimeSQL := `published_at = published_at, taken_down_at = taken_down_at, deleted_at = deleted_at`
	switch to {
	case model.ArticleStatusPublished:
		setTimeSQL = `published_at = COALESCE(published_at, NOW()), taken_down_at = NULL`
	case model.ArticleStatusTakenDown:
		setTimeSQL = `taken_down_at = NOW()`
	case model.ArticleStatusDeleted:
		setTimeSQL = `deleted_at = NOW()`
	}

	a := &model.Article{}
	updateSQL := `UPDATE articles SET status=$2, updated_at=NOW(), ` + setTimeSQL + ` WHERE id=$1 RETURNING ` + articleReturningColumns
	if err := scanArticle(tx.QueryRow(updateSQL, articleID, to), a); err != nil {
		return nil, err
	}
	var adminIDArg any
	if strings.TrimSpace(adminID) != "" {
		adminIDArg = adminID
	}
	if _, err := tx.Exec(`
		INSERT INTO article_audit_events (article_id, admin_id, action, from_status, to_status, note)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		articleID, adminIDArg, action, from, to, note,
	); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return a, nil
}

func IncrementArticleViewCount(articleID string) error {
	res, err := database.DB.Exec(`UPDATE articles SET view_count=view_count+1, updated_at=NOW() WHERE id=$1 AND status='published'`, articleID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func RecordArticleOriginalClick(articleID, userID string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`UPDATE articles SET original_click_count=original_click_count+1, updated_at=NOW() WHERE id=$1 AND status='published'`, articleID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	if _, err := tx.Exec(`INSERT INTO article_original_clicks (article_id, user_id) VALUES ($1, $2)`, articleID, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func CreateArticleCollectionTask(triggerType string, keywordCount int) (*model.ArticleCollectionTask, error) {
	t := &model.ArticleCollectionTask{}
	err := database.DB.QueryRow(`
		INSERT INTO article_collection_tasks (trigger_type, status, keyword_count)
		VALUES ($1, 'running', $2)
		RETURNING id, trigger_type, status, started_at, finished_at, keyword_count, found_count, inserted_count, duplicate_count, failed_count, error_msg, created_at, updated_at`,
		triggerType, keywordCount,
	).Scan(&t.ID, &t.TriggerType, &t.Status, &t.StartedAt, &t.FinishedAt, &t.KeywordCount, &t.FoundCount, &t.InsertedCount, &t.DuplicateCount, &t.FailedCount, &t.ErrorMsg, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func AddArticleCollectionTaskItem(taskID, articleID, originalURL, keyword string, status model.CollectionItemStatus, errorMsg string) error {
	return AddArticleCollectionTaskItemWithQuality(ArticleCollectionTaskItemInput{
		TaskID:      taskID,
		ArticleID:   articleID,
		OriginalURL: originalURL,
		Keyword:     keyword,
		Status:      status,
		ErrorMsg:    errorMsg,
	})
}

func AddArticleCollectionTaskItemWithQuality(input ArticleCollectionTaskItemInput) error {
	var articleIDArg any
	if input.ArticleID != "" {
		articleIDArg = input.ArticleID
	}
	reasons, err := marshalArticleQualityReasons(input.QualityReasons)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec(`
		INSERT INTO article_collection_task_items (task_id, article_id, original_url, keyword, status, error_msg, quality_score, quality_reasons, skip_reason, auto_published, auto_publish_reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		input.TaskID, articleIDArg, input.OriginalURL, input.Keyword, input.Status, input.ErrorMsg, clampInt(input.QualityScore, 0, 100), reasons, strings.TrimSpace(input.SkipReason), input.AutoPublished, strings.TrimSpace(input.AutoPublishReason),
	)
	return err
}

func FinishArticleCollectionTask(taskID string, status model.CollectionTaskStatus, counts ArticleCollectionCounts, errorMsg string) error {
	_, err := database.DB.Exec(`
		UPDATE article_collection_tasks
		SET status=$2, finished_at=NOW(), keyword_count=$3, found_count=$4, inserted_count=$5,
		    duplicate_count=$6, failed_count=$7, error_msg=$8, updated_at=NOW()
		WHERE id=$1`,
		taskID, status, counts.KeywordCount, counts.FoundCount, counts.InsertedCount, counts.DuplicateCount, counts.FailedCount, errorMsg,
	)
	return err
}

func ListArticleCollectionTasks(limit, offset int) ([]model.ArticleCollectionTask, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := database.DB.Query(`
		SELECT id, trigger_type, status, started_at, finished_at, keyword_count, found_count, inserted_count, duplicate_count, failed_count, error_msg, created_at, updated_at
		FROM article_collection_tasks
		ORDER BY started_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []model.ArticleCollectionTask
	for rows.Next() {
		var task model.ArticleCollectionTask
		if err := rows.Scan(&task.ID, &task.TriggerType, &task.Status, &task.StartedAt, &task.FinishedAt, &task.KeywordCount, &task.FoundCount, &task.InsertedCount, &task.DuplicateCount, &task.FailedCount, &task.ErrorMsg, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func ListArticleCollectionTaskItems(taskID string, limit, offset int) ([]model.ArticleCollectionTaskItem, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := database.DB.Query(`
		SELECT
			i.id, i.task_id, i.article_id, i.original_url, i.keyword, i.status, i.error_msg,
			COALESCE(a.title, ''), COALESCE(a.source_name, ''), COALESCE(a.cover_url, ''),
			COALESCE(a.body_fetch_status, ''), COALESCE(a.ai_status, ''),
			COALESCE(NULLIF(i.quality_score, 0), a.quality_score, 0), COALESCE(NULLIF(i.quality_reasons, '[]'::jsonb), a.quality_reasons, '[]'::jsonb), COALESCE(i.skip_reason, ''),
			COALESCE(i.auto_published, false), COALESCE(i.auto_publish_reason, ''),
			i.created_at, i.updated_at
		FROM article_collection_task_items i
		LEFT JOIN articles a ON a.id = i.article_id
		WHERE i.task_id=$1
		ORDER BY
			CASE i.status
				WHEN 'inserted' THEN 1
				WHEN 'duplicate' THEN 2
				WHEN 'failed' THEN 3
				ELSE 4
			END,
			i.created_at ASC
		LIMIT $2 OFFSET $3`, taskID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.ArticleCollectionTaskItem
	for rows.Next() {
		var item model.ArticleCollectionTaskItem
		var qualityRaw []byte
		if err := rows.Scan(
			&item.ID, &item.TaskID, &item.ArticleID, &item.OriginalURL, &item.Keyword, &item.Status, &item.ErrorMsg,
			&item.ArticleTitle, &item.SourceName, &item.CoverURL, &item.BodyFetchStatus, &item.AIStatus,
			&item.QualityScore, &qualityRaw, &item.SkipReason, &item.AutoPublished, &item.AutoPublishReason,
			&item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.QualityReasons = parseArticleQualityReasons(qualityRaw)
		items = append(items, item)
	}
	return items, rows.Err()
}

func ListFailedArticleCollectionTaskKeywords(taskID string) ([]string, error) {
	rows, err := database.DB.Query(`
		SELECT DISTINCT keyword
		FROM article_collection_task_items
		WHERE task_id=$1 AND status='failed' AND keyword <> ''
		ORDER BY keyword ASC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keywords []string
	for rows.Next() {
		var keyword string
		if err := rows.Scan(&keyword); err != nil {
			return nil, err
		}
		keywords = append(keywords, keyword)
	}
	return keywords, rows.Err()
}

func GetArticleCollectionConfig() (*model.ArticleCollectionConfig, error) {
	cfg := &model.ArticleCollectionConfig{}
	err := database.DB.QueryRow(`
		SELECT id, enabled, frequency, schedule_interval_minutes, max_results_per_run, search_page_min, search_page_max,
		       auto_publish_enabled, auto_publish_min_quality_score, auto_publish_requires_body, auto_publish_max_per_run,
		       last_run_at, updated_at
		FROM article_collection_config
		ORDER BY updated_at DESC LIMIT 1`,
	).Scan(&cfg.ID, &cfg.Enabled, &cfg.Frequency, &cfg.ScheduleInterval, &cfg.MaxResultsPerRun, &cfg.SearchPageMin, &cfg.SearchPageMax, &cfg.AutoPublishEnabled, &cfg.AutoPublishMinQualityScore, &cfg.AutoPublishRequiresBody, &cfg.AutoPublishMaxPerRun, &cfg.LastRunAt, &cfg.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	normalizeArticleCollectionConfig(cfg)
	return cfg, err
}

func UpdateArticleCollectionConfig(input model.ArticleCollectionConfig) (*model.ArticleCollectionConfig, error) {
	cfgInput := &model.ArticleCollectionConfig{
		Enabled:                    input.Enabled,
		Frequency:                  input.Frequency,
		ScheduleInterval:           input.ScheduleInterval,
		MaxResultsPerRun:           input.MaxResultsPerRun,
		SearchPageMin:              input.SearchPageMin,
		SearchPageMax:              input.SearchPageMax,
		AutoPublishEnabled:         input.AutoPublishEnabled,
		AutoPublishMinQualityScore: input.AutoPublishMinQualityScore,
		AutoPublishRequiresBody:    input.AutoPublishRequiresBody,
		AutoPublishMaxPerRun:       input.AutoPublishMaxPerRun,
	}
	normalizeArticleCollectionConfig(cfgInput)
	cfg, err := GetArticleCollectionConfig()
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &model.ArticleCollectionConfig{}
		err = database.DB.QueryRow(`
			INSERT INTO article_collection_config (enabled, frequency, schedule_interval_minutes, max_results_per_run, search_page_min, search_page_max, auto_publish_enabled, auto_publish_min_quality_score, auto_publish_requires_body, auto_publish_max_per_run)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id, enabled, frequency, schedule_interval_minutes, max_results_per_run, search_page_min, search_page_max, auto_publish_enabled, auto_publish_min_quality_score, auto_publish_requires_body, auto_publish_max_per_run, last_run_at, updated_at`,
			cfgInput.Enabled, cfgInput.Frequency, cfgInput.ScheduleInterval, cfgInput.MaxResultsPerRun, cfgInput.SearchPageMin, cfgInput.SearchPageMax, cfgInput.AutoPublishEnabled, cfgInput.AutoPublishMinQualityScore, cfgInput.AutoPublishRequiresBody, cfgInput.AutoPublishMaxPerRun,
		).Scan(&cfg.ID, &cfg.Enabled, &cfg.Frequency, &cfg.ScheduleInterval, &cfg.MaxResultsPerRun, &cfg.SearchPageMin, &cfg.SearchPageMax, &cfg.AutoPublishEnabled, &cfg.AutoPublishMinQualityScore, &cfg.AutoPublishRequiresBody, &cfg.AutoPublishMaxPerRun, &cfg.LastRunAt, &cfg.UpdatedAt)
		normalizeArticleCollectionConfig(cfg)
		return cfg, err
	}
	err = database.DB.QueryRow(`
		UPDATE article_collection_config
		SET enabled=$2, frequency=$3, schedule_interval_minutes=$4, max_results_per_run=$5, search_page_min=$6, search_page_max=$7,
		    auto_publish_enabled=$8, auto_publish_min_quality_score=$9, auto_publish_requires_body=$10, auto_publish_max_per_run=$11, updated_at=NOW()
		WHERE id=$1
		RETURNING id, enabled, frequency, schedule_interval_minutes, max_results_per_run, search_page_min, search_page_max, auto_publish_enabled, auto_publish_min_quality_score, auto_publish_requires_body, auto_publish_max_per_run, last_run_at, updated_at`,
		cfg.ID, cfgInput.Enabled, cfgInput.Frequency, cfgInput.ScheduleInterval, cfgInput.MaxResultsPerRun, cfgInput.SearchPageMin, cfgInput.SearchPageMax, cfgInput.AutoPublishEnabled, cfgInput.AutoPublishMinQualityScore, cfgInput.AutoPublishRequiresBody, cfgInput.AutoPublishMaxPerRun,
	).Scan(&cfg.ID, &cfg.Enabled, &cfg.Frequency, &cfg.ScheduleInterval, &cfg.MaxResultsPerRun, &cfg.SearchPageMin, &cfg.SearchPageMax, &cfg.AutoPublishEnabled, &cfg.AutoPublishMinQualityScore, &cfg.AutoPublishRequiresBody, &cfg.AutoPublishMaxPerRun, &cfg.LastRunAt, &cfg.UpdatedAt)
	normalizeArticleCollectionConfig(cfg)
	return cfg, err
}

func normalizeArticleCollectionConfig(cfg *model.ArticleCollectionConfig) {
	if cfg == nil {
		return
	}
	if cfg.Frequency == "" {
		cfg.Frequency = "daily"
	}
	if cfg.ScheduleInterval <= 0 {
		if cfg.Frequency == "weekly" {
			cfg.ScheduleInterval = 10080
		} else {
			cfg.ScheduleInterval = 1440
		}
	}
	if cfg.ScheduleInterval < 1 {
		cfg.ScheduleInterval = 1
	}
	if cfg.ScheduleInterval > 10080 {
		cfg.ScheduleInterval = 10080
	}
	if cfg.MaxResultsPerRun <= 0 {
		cfg.MaxResultsPerRun = 20
	}
	if cfg.SearchPageMin <= 0 {
		cfg.SearchPageMin = 1
	}
	if cfg.SearchPageMax <= 0 {
		cfg.SearchPageMax = 5
	}
	if cfg.SearchPageMin > 20 {
		cfg.SearchPageMin = 20
	}
	if cfg.SearchPageMax > 20 {
		cfg.SearchPageMax = 20
	}
	if cfg.SearchPageMin > cfg.SearchPageMax {
		cfg.SearchPageMin, cfg.SearchPageMax = cfg.SearchPageMax, cfg.SearchPageMin
	}
	cfg.AutoPublishMinQualityScore = clampInt(cfg.AutoPublishMinQualityScore, 0, 100)
	if cfg.AutoPublishMaxPerRun < 0 {
		cfg.AutoPublishMaxPerRun = 0
	}
	if cfg.AutoPublishMaxPerRun > 20 {
		cfg.AutoPublishMaxPerRun = 20
	}
}

func GetArticleQualityConfig() (*model.ArticleQualityConfig, error) {
	cfg := &model.ArticleQualityConfig{}
	err := database.DB.QueryRow(`
		SELECT id, quality_filter_enabled, min_quality_score, allow_without_body, bonus_keywords, source_blacklist, preferred_sources, ai_quality_check_enabled, updated_at
		FROM article_quality_config
		ORDER BY updated_at DESC LIMIT 1`,
	).Scan(
		&cfg.ID, &cfg.Enabled, &cfg.MinQualityScore, &cfg.AllowWithoutBody,
		pq.Array(&cfg.BonusKeywords), pq.Array(&cfg.SourceBlacklist), pq.Array(&cfg.PreferredSources),
		&cfg.AIQualityCheckEnabled, &cfg.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	normalizeArticleQualityConfig(cfg)
	return cfg, err
}

func UpdateArticleQualityConfig(input model.ArticleQualityConfig) (*model.ArticleQualityConfig, error) {
	normalizeArticleQualityConfig(&input)
	cfg, err := GetArticleQualityConfig()
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &model.ArticleQualityConfig{}
		err = database.DB.QueryRow(`
			INSERT INTO article_quality_config (quality_filter_enabled, min_quality_score, allow_without_body, bonus_keywords, source_blacklist, preferred_sources, ai_quality_check_enabled)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, quality_filter_enabled, min_quality_score, allow_without_body, bonus_keywords, source_blacklist, preferred_sources, ai_quality_check_enabled, updated_at`,
			input.Enabled, input.MinQualityScore, input.AllowWithoutBody, pq.Array(input.BonusKeywords), pq.Array(input.SourceBlacklist), pq.Array(input.PreferredSources), input.AIQualityCheckEnabled,
		).Scan(
			&cfg.ID, &cfg.Enabled, &cfg.MinQualityScore, &cfg.AllowWithoutBody,
			pq.Array(&cfg.BonusKeywords), pq.Array(&cfg.SourceBlacklist), pq.Array(&cfg.PreferredSources),
			&cfg.AIQualityCheckEnabled, &cfg.UpdatedAt,
		)
		normalizeArticleQualityConfig(cfg)
		return cfg, err
	}
	err = database.DB.QueryRow(`
		UPDATE article_quality_config
		SET quality_filter_enabled=$2, min_quality_score=$3, allow_without_body=$4, bonus_keywords=$5, source_blacklist=$6, preferred_sources=$7, ai_quality_check_enabled=$8, updated_at=NOW()
		WHERE id=$1
		RETURNING id, quality_filter_enabled, min_quality_score, allow_without_body, bonus_keywords, source_blacklist, preferred_sources, ai_quality_check_enabled, updated_at`,
		cfg.ID, input.Enabled, input.MinQualityScore, input.AllowWithoutBody, pq.Array(input.BonusKeywords), pq.Array(input.SourceBlacklist), pq.Array(input.PreferredSources), input.AIQualityCheckEnabled,
	).Scan(
		&cfg.ID, &cfg.Enabled, &cfg.MinQualityScore, &cfg.AllowWithoutBody,
		pq.Array(&cfg.BonusKeywords), pq.Array(&cfg.SourceBlacklist), pq.Array(&cfg.PreferredSources),
		&cfg.AIQualityCheckEnabled, &cfg.UpdatedAt,
	)
	normalizeArticleQualityConfig(cfg)
	return cfg, err
}

func normalizeArticleQualityConfig(cfg *model.ArticleQualityConfig) {
	if cfg == nil {
		return
	}
	cfg.MinQualityScore = clampInt(cfg.MinQualityScore, 0, 100)
	cfg.BonusKeywords = normalizeStringList(cfg.BonusKeywords)
	cfg.SourceBlacklist = normalizeStringList(cfg.SourceBlacklist)
	cfg.PreferredSources = normalizeStringList(cfg.PreferredSources)
}

func MarkArticleCollectionConfigLastRun() error {
	_, err := database.DB.Exec(`
		UPDATE article_collection_config
		SET last_run_at=NOW(), updated_at=NOW()
		WHERE id = (SELECT id FROM article_collection_config ORDER BY updated_at DESC LIMIT 1)`)
	return err
}

func GetActiveArticleAIProvider() (*model.ArticleAIProvider, error) {
	p := &model.ArticleAIProvider{}
	err := database.DB.QueryRow(`
		SELECT id, name, type, base_url, model, api_key_encrypted, api_key_preview, active, created_at, updated_at
		FROM article_ai_providers WHERE active=true LIMIT 1`,
	).Scan(&p.ID, &p.Name, &p.Type, &p.BaseURL, &p.Model, &p.APIKeyEncrypted, &p.APIKeyPreview, &p.Active, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func CreateArticleAIProvider(name, providerType, baseURL, modelName, encryptedKey, preview string, active bool) (*model.ArticleAIProvider, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if active {
		if _, err := tx.Exec(`UPDATE article_ai_providers SET active=false, updated_at=NOW()`); err != nil {
			return nil, err
		}
	}
	p := &model.ArticleAIProvider{}
	err = tx.QueryRow(`
		INSERT INTO article_ai_providers (name, type, base_url, model, api_key_encrypted, api_key_preview, active)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, name, type, base_url, model, api_key_encrypted, api_key_preview, active, created_at, updated_at`,
		name, providerType, baseURL, modelName, encryptedKey, preview, active,
	).Scan(&p.ID, &p.Name, &p.Type, &p.BaseURL, &p.Model, &p.APIKeyEncrypted, &p.APIKeyPreview, &p.Active, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, tx.Commit()
}

func GetArticleAIPrompt() (*model.ArticleAIPrompt, error) {
	p := &model.ArticleAIPrompt{}
	err := database.DB.QueryRow(`
		SELECT id, content, description, updated_at
		FROM article_ai_prompts
		ORDER BY updated_at DESC LIMIT 1`,
	).Scan(&p.ID, &p.Content, &p.Description, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func UpdateArticleAIPrompt(content, description string) (*model.ArticleAIPrompt, error) {
	p, err := GetArticleAIPrompt()
	if err != nil {
		return nil, err
	}
	if p == nil {
		p = &model.ArticleAIPrompt{}
		err = database.DB.QueryRow(`
			INSERT INTO article_ai_prompts (content, description)
			VALUES ($1, $2)
			RETURNING id, content, description, updated_at`,
			content, description,
		).Scan(&p.ID, &p.Content, &p.Description, &p.UpdatedAt)
		return p, err
	}
	err = database.DB.QueryRow(`
		UPDATE article_ai_prompts SET content=$2, description=$3, updated_at=NOW()
		WHERE id=$1
		RETURNING id, content, description, updated_at`,
		p.ID, content, description,
	).Scan(&p.ID, &p.Content, &p.Description, &p.UpdatedAt)
	return p, err
}

func SaveArticleAIAnalysis(articleID string, analysis model.ArticleAIAnalysis) error {
	raw, err := json.Marshal(analysis)
	if err != nil {
		return err
	}
	res, err := database.DB.Exec(`
		UPDATE articles
		SET ai_analysis=$2, ai_status='succeeded', ai_error_msg='', summary=COALESCE(NULLIF($3, ''), summary), updated_at=NOW()
		WHERE id=$1`,
		articleID, raw, analysis.OneSentenceSummary,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func RecordArticleAIAnalysisFailure(articleID, errorMsg string) error {
	res, err := database.DB.Exec(`UPDATE articles SET ai_status='failed', ai_error_msg=$2, updated_at=NOW() WHERE id=$1`, articleID, errorMsg)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func getArticleByQuery(query string, args ...any) (*model.Article, error) {
	a := &model.Article{}
	err := scanArticle(database.DB.QueryRow(query, args...), a)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = loadTagsForArticle(a)
	return a, nil
}

func scanArticle(scanner articleScanner, a *model.Article) error {
	var categoryID sql.NullString
	var publishedAtSource, publishedAt, takenDownAt, deletedAt sql.NullTime
	var aiRaw []byte
	var qualityRaw []byte
	if err := scanner.Scan(
		&a.ID, &a.Title, &a.SourceName, &a.OriginalURL, &a.CanonicalURLHash, &a.CoverURL,
		&publishedAtSource, &a.SearchSnippet, &a.Summary, &aiRaw, &a.AIStatus,
		&a.AIErrorMsg, &categoryID, &a.Status, &a.ViewCount, &a.OriginalClickCount,
		&a.QualityScore, &qualityRaw, &a.FullTextAuthorized, &a.BodyContent, &a.BodyFetchStatus, &a.BodyFetchError, &a.CreatedAt, &a.UpdatedAt,
		&publishedAt, &takenDownAt, &deletedAt,
	); err != nil {
		return err
	}
	if categoryID.Valid {
		a.CategoryID = &categoryID.String
	}
	if publishedAtSource.Valid {
		a.PublishedAtSource = &publishedAtSource.Time
	}
	if publishedAt.Valid {
		a.PublishedAt = &publishedAt.Time
	}
	if takenDownAt.Valid {
		a.TakenDownAt = &takenDownAt.Time
	}
	if deletedAt.Valid {
		a.DeletedAt = &deletedAt.Time
	}
	a.AIAnalysisRaw = aiRaw
	a.QualityReasonsRaw = qualityRaw
	parseArticleAI(a)
	parseArticleQuality(a)
	return nil
}

func parseArticleAI(a *model.Article) {
	if len(a.AIAnalysisRaw) == 0 || string(a.AIAnalysisRaw) == "null" {
		return
	}
	var analysis model.ArticleAIAnalysis
	if err := json.Unmarshal(a.AIAnalysisRaw, &analysis); err == nil {
		a.AIAnalysis = &analysis
	}
}

func parseArticleQuality(a *model.Article) {
	a.QualityReasons = parseArticleQualityReasons(a.QualityReasonsRaw)
}

func parseArticleQualityReasons(raw []byte) []model.ArticleQualityReason {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var reasons []model.ArticleQualityReason
	if err := json.Unmarshal(raw, &reasons); err != nil {
		return nil
	}
	return reasons
}

func marshalArticleQualityReasons(reasons []model.ArticleQualityReason) ([]byte, error) {
	if reasons == nil {
		reasons = []model.ArticleQualityReason{}
	}
	return json.Marshal(reasons)
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func normalizeStringList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	return out
}

func buildArticleListWhere(filter ArticleListFilter) (string, []any) {
	var clauses []string
	var args []any
	if strings.TrimSpace(filter.Status) != "" {
		args = append(args, filter.Status)
		clauses = append(clauses, fmt.Sprintf("a.status=$%d", len(args)))
	}
	if strings.TrimSpace(filter.CategoryID) != "" {
		args = append(args, filter.CategoryID)
		clauses = append(clauses, fmt.Sprintf("a.category_id=$%d", len(args)))
	}
	if strings.TrimSpace(filter.TagID) != "" {
		args = append(args, filter.TagID)
		clauses = append(clauses, fmt.Sprintf("EXISTS (SELECT 1 FROM article_tag_links atl WHERE atl.article_id=a.id AND atl.tag_id=$%d)", len(args)))
	}
	if q := strings.TrimSpace(filter.Query); q != "" {
		args = append(args, "%"+q+"%")
		placeholder := fmt.Sprintf("$%d", len(args))
		clauses = append(clauses, "(a.title ILIKE "+placeholder+" OR a.source_name ILIKE "+placeholder+" OR a.summary ILIKE "+placeholder+" OR a.search_snippet ILIKE "+placeholder+")")
	}
	if filter.MinQualityScore > 0 {
		args = append(args, filter.MinQualityScore)
		clauses = append(clauses, fmt.Sprintf("a.quality_score >= $%d", len(args)))
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func articleOrderBy(sort string) string {
	if sort == "quality" {
		return " ORDER BY a.quality_score DESC, COALESCE(a.published_at_source, a.published_at, a.created_at) DESC"
	}
	if sort == "hot" {
		return " ORDER BY a.view_count DESC, COALESCE(a.published_at_source, a.published_at, a.created_at) DESC"
	}
	return " ORDER BY COALESCE(a.published_at_source, a.published_at, a.created_at) DESC"
}

func loadTagsForArticle(article *model.Article) error {
	if article == nil || article.ID == "" {
		return nil
	}
	articles := []model.Article{*article}
	if err := loadTagsForArticles(articles); err != nil {
		return err
	}
	article.Tags = articles[0].Tags
	return nil
}

func loadTagsForArticles(articles []model.Article) error {
	if len(articles) == 0 {
		return nil
	}
	ids := make([]string, 0, len(articles))
	index := make(map[string]int, len(articles))
	for i := range articles {
		ids = append(ids, articles[i].ID)
		index[articles[i].ID] = i
	}
	rows, err := database.DB.Query(`
		SELECT atl.article_id, t.id, t.name, t.slug, t.active, t.created_at, t.updated_at
		FROM article_tag_links atl
		JOIN article_tags t ON t.id=atl.tag_id
		WHERE atl.article_id::text = ANY($1)
		ORDER BY t.created_at DESC`, pq.Array(ids))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var articleID string
		var tag model.ArticleTag
		if err := rows.Scan(&articleID, &tag.ID, &tag.Name, &tag.Slug, &tag.Active, &tag.CreatedAt, &tag.UpdatedAt); err != nil {
			return err
		}
		if i, ok := index[articleID]; ok {
			articles[i].Tags = append(articles[i].Tags, tag)
		}
	}
	return rows.Err()
}

func validateArticleRepoTransition(from, to model.ArticleStatus) error {
	if from == to {
		return nil
	}
	allowed := map[model.ArticleStatus]map[model.ArticleStatus]bool{
		model.ArticleStatusCandidate: {
			model.ArticleStatusPublished: true,
			model.ArticleStatusRejected:  true,
			model.ArticleStatusDeleted:   true,
		},
		model.ArticleStatusPublished: {
			model.ArticleStatusTakenDown: true,
			model.ArticleStatusDeleted:   true,
		},
		model.ArticleStatusRejected: {
			model.ArticleStatusPublished: true,
			model.ArticleStatusDeleted:   true,
		},
		model.ArticleStatusTakenDown: {
			model.ArticleStatusPublished: true,
			model.ArticleStatusDeleted:   true,
		},
	}
	if allowed[from][to] {
		return nil
	}
	return fmt.Errorf("非法文章状态流转: %s -> %s", from, to)
}

func articleAuditAction(to model.ArticleStatus) string {
	switch to {
	case model.ArticleStatusPublished:
		return "publish"
	case model.ArticleStatusRejected:
		return "reject"
	case model.ArticleStatusTakenDown:
		return "take_down"
	case model.ArticleStatusDeleted:
		return "delete"
	default:
		return "update"
	}
}
