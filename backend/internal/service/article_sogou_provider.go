package service

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"yuanju/internal/model"
)

type SogouWeChatProvider struct {
	baseURL       string
	client        *http.Client
	searchPageMin int
	searchPageMax int
}

func NewSogouWeChatProvider(baseURL string, client *http.Client) *SogouWeChatProvider {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://weixin.sogou.com/weixin"
	}
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	return &SogouWeChatProvider{baseURL: baseURL, client: client, searchPageMin: 1, searchPageMax: 1}
}

func (p *SogouWeChatProvider) WithSearchPageRange(minPage, maxPage int) *SogouWeChatProvider {
	if p == nil {
		return NewSogouWeChatProvider("", nil).WithSearchPageRange(minPage, maxPage)
	}
	if minPage <= 0 {
		minPage = 1
	}
	if maxPage <= 0 {
		maxPage = 1
	}
	if minPage > 20 {
		minPage = 20
	}
	if maxPage > 20 {
		maxPage = 20
	}
	if minPage > maxPage {
		minPage, maxPage = maxPage, minPage
	}
	copyProvider := *p
	copyProvider.searchPageMin = minPage
	copyProvider.searchPageMax = maxPage
	return &copyProvider
}

func (p *SogouWeChatProvider) Search(ctx context.Context, keyword string, limit int) ([]ArticleSearchResult, error) {
	session := p.withSession()
	u, err := url.Parse(p.baseURL)
	if err != nil {
		return nil, fmt.Errorf("sogou base url 无效: %w", err)
	}
	q := u.Query()
	q.Set("type", "2")
	q.Set("query", keyword)
	page := session.randomSearchPage()
	if page > 1 {
		q.Set("page", strconv.Itoa(page))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/126 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	resp, err := session.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sogou search 请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("sogou search 状态异常: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	bodyText := string(body)
	session.approveSearch(ctx, u, bodyText)
	parseLimit := limit
	if parseLimit > 0 {
		parseLimit *= 3
		if parseLimit < 20 {
			parseLimit = 20
		}
	}
	results := parseSogouWeChatResults(bodyText, parseLimit)
	results = randomSampleArticleResults(results, limit)
	session.enrichArticleBodies(ctx, results)
	return results, nil
}

var (
	sogouStructuredItemRe = regexp.MustCompile(`(?is)<li\b[^>]+id=["']sogou_vr_11002601_box_\d+["'][^>]*>(.*?)</li>`)
	sogouItemRe           = regexp.MustCompile(`(?is)<li\b[^>]*>(.*?)</li>`)
	sogouLinkRe           = regexp.MustCompile(`(?is)<h3[^>]*>\s*<a[^>]+href=["']([^"']+)["'][^>]*>(.*?)</a>\s*</h3>`)
	sogouCoverRe          = regexp.MustCompile(`(?is)<img[^>]+src=["']([^"']+)["']`)
	sogouSnippetRe        = regexp.MustCompile(`(?is)<p[^>]+class=["'][^"']*txt-info[^"']*["'][^>]*>(.*?)</p>`)
	sogouAccountRe        = regexp.MustCompile(`(?is)<(?:a|span)[^>]+class=["'][^"']*(?:account|all-time-y2)[^"']*["'][^>]*>(.*?)</(?:a|span)>`)
	sogouTimeRe           = regexp.MustCompile(`(?is)timeConvert\('([0-9]+)'\)`)
	sogouUUIDRe           = regexp.MustCompile(`(?is)var\s+uuid\s*=\s*["']([^"']+)["']`)
	sogouSSTokenRe        = regexp.MustCompile(`(?is)var\s+ssToken\s*=\s*["']([^"']+)["']`)
	sogouURLAppendRe      = regexp.MustCompile(`(?is)url\s*\+=\s*['"]([^'"]*)['"]`)
	sogouInnerApproveRe   = regexp.MustCompile(`(?is)approve\?uuid='\s*\+\s*['"]([^'"]+)['"]\s*\+\s*'&token='\s*\+\s*['"]([^'"]+)['"]\s*\+\s*'&from=inner`)
	wechatTitleRe         = regexp.MustCompile(`(?is)<h1[^>]+id=["']activity-name["'][^>]*>(.*?)</h1>`)
	wechatVideoMarkerRe   = regexp.MustCompile(`(?is)(?:js_page_video|video_iframe|mp_video|wxv_[0-9A-Za-z]+|腾讯视频|视频号)`)
	htmlTagRe             = regexp.MustCompile(`(?is)<[^>]+>`)
	blockBreakRe          = regexp.MustCompile(`(?i)</?(?:p|section|div|br|li|h[1-6])[^>]*>`)
	spaceRe               = regexp.MustCompile(`\s+`)
	lineSpaceRe           = regexp.MustCompile(`[ \t\r\f]+`)
	blankLineRe           = regexp.MustCompile(`\n{3,}`)
)

func (p *SogouWeChatProvider) withSession() *SogouWeChatProvider {
	if p == nil || p.client == nil {
		return NewSogouWeChatProvider("", nil)
	}
	clientCopy := *p.client
	if clientCopy.Jar == nil {
		if jar, err := cookiejar.New(nil); err == nil {
			clientCopy.Jar = jar
		}
	}
	return &SogouWeChatProvider{baseURL: p.baseURL, client: &clientCopy, searchPageMin: p.searchPageMin, searchPageMax: p.searchPageMax}
}

func (p *SogouWeChatProvider) randomSearchPage() int {
	if p == nil {
		return 1
	}
	minPage := p.searchPageMin
	maxPage := p.searchPageMax
	if minPage <= 0 {
		minPage = 1
	}
	if maxPage <= 0 {
		maxPage = minPage
	}
	if minPage > maxPage {
		minPage, maxPage = maxPage, minPage
	}
	if minPage == maxPage {
		return minPage
	}
	n, err := crand.Int(crand.Reader, big.NewInt(int64(maxPage-minPage+1)))
	if err != nil {
		return minPage + int(time.Now().UnixNano()%int64(maxPage-minPage+1))
	}
	return minPage + int(n.Int64())
}

func (p *SogouWeChatProvider) approveSearch(ctx context.Context, searchURL *url.URL, body string) {
	uuid, token := extractSogouSearchSession(body)
	if uuid == "" || token == "" || searchURL == nil {
		return
	}
	approveURL := &url.URL{Scheme: searchURL.Scheme, Host: searchURL.Host, Path: "/approve"}
	q := approveURL.Query()
	q.Set("uuid", uuid)
	q.Set("token", token)
	q.Set("from", "search")
	approveURL.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, approveURL.String(), nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/126 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Referer", searchURL.String())
	resp, err := p.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
}

func extractSogouSearchSession(body string) (uuid, token string) {
	if m := sogouUUIDRe.FindStringSubmatch(body); len(m) > 1 {
		uuid = strings.TrimSpace(m[1])
	}
	if m := sogouSSTokenRe.FindStringSubmatch(body); len(m) > 1 {
		token = strings.TrimSpace(m[1])
	}
	return uuid, token
}

func parseSogouWeChatResults(body string, limit int) []ArticleSearchResult {
	items := sogouStructuredItemRe.FindAllStringSubmatch(body, -1)
	if len(items) == 0 {
		items = sogouItemRe.FindAllStringSubmatch(body, -1)
	}
	results := make([]ArticleSearchResult, 0, len(items))
	for _, item := range items {
		block := item[1]
		link := sogouLinkRe.FindStringSubmatch(block)
		if len(link) < 3 {
			continue
		}
		originalURL := normalizeSogouURL(link[1])
		result := ArticleSearchResult{
			Rank:        len(results) + 1,
			OriginalURL: originalURL,
			Title:       cleanHTMLText(link[2]),
		}
		if m := sogouCoverRe.FindStringSubmatch(block); len(m) > 1 {
			result.CoverURL = normalizeSogouURL(m[1])
		}
		if m := sogouSnippetRe.FindStringSubmatch(block); len(m) > 1 {
			result.SearchSnippet = cleanHTMLText(m[1])
		}
		if m := sogouAccountRe.FindStringSubmatch(block); len(m) > 1 {
			result.SourceName = cleanHTMLText(m[1])
		}
		if m := sogouTimeRe.FindStringSubmatch(block); len(m) > 1 {
			if unixSeconds, err := strconv.ParseInt(m[1], 10, 64); err == nil {
				t := time.Unix(unixSeconds, 0).UTC()
				result.PublishedAtSource = &t
			}
		}
		if result.Title == "" || result.OriginalURL == "" {
			continue
		}
		results = append(results, result)
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results
}

func randomSampleArticleResults(results []ArticleSearchResult, limit int) []ArticleSearchResult {
	if limit <= 0 || len(results) <= limit {
		return results
	}
	out := append([]ArticleSearchResult(nil), results...)
	for i := len(out) - 1; i > 0; i-- {
		j := randomInt(i + 1)
		out[i], out[j] = out[j], out[i]
	}
	return out[:limit]
}

func randomInt(maxExclusive int) int {
	if maxExclusive <= 1 {
		return 0
	}
	n, err := crand.Int(crand.Reader, big.NewInt(int64(maxExclusive)))
	if err != nil {
		return int(time.Now().UnixNano() % int64(maxExclusive))
	}
	return int(n.Int64())
}

func (p *SogouWeChatProvider) enrichArticleBodies(ctx context.Context, results []ArticleSearchResult) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 1)
	for i := range results {
		if strings.TrimSpace(results[i].OriginalURL) == "" {
			continue
		}
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}
			bodyCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			content, finalURL, status, reason, err := p.fetchArticleBody(bodyCtx, results[index].OriginalURL)
			results[index].BodyFetchStatus = status
			results[index].BodyFetchError = reason
			if err != nil || strings.TrimSpace(content) == "" {
				return
			}
			results[index].BodyContent = content
			results[index].FullTextAuthorized = true
			results[index].BodyFetchStatus = model.ArticleBodyFetchStatusSucceeded
			results[index].BodyFetchError = ""
			if isWeChatArticleURL(finalURL) {
				results[index].OriginalURL = finalURL
			}
		}(i)
	}
	wg.Wait()
}

func (p *SogouWeChatProvider) fetchArticleBody(ctx context.Context, rawURL string) (content string, finalURL string, status model.ArticleBodyFetchStatus, reason string, err error) {
	target := prepareSogouClickURL(rawURL)
	body, finalURL, err := p.fetchHTML(ctx, target, "https://weixin.sogou.com/weixin")
	if err != nil {
		status, reason = classifyArticleBodyFetchError(err)
		return "", finalURL, status, reason, err
	}
	content = parseWeChatArticleBody(body)
	if strings.TrimSpace(content) != "" {
		return content, finalURL, model.ArticleBodyFetchStatusSucceeded, "", nil
	}
	redirectURL := extractSogouScriptRedirectURL(body)
	if redirectURL == "" {
		status, reason = classifyArticleBodyFetchHTML(body, finalURL)
		return "", finalURL, status, reason, nil
	}
	p.approveInnerRedirect(ctx, body, finalURL)
	body, finalURL, err = p.fetchHTML(ctx, redirectURL, finalURL)
	if err != nil {
		status, reason = classifyArticleBodyFetchError(err)
		return "", finalURL, status, reason, err
	}
	content = parseWeChatArticleBody(body)
	if strings.TrimSpace(content) != "" {
		return content, finalURL, model.ArticleBodyFetchStatusSucceeded, "", nil
	}
	status, reason = classifyArticleBodyFetchHTML(body, finalURL)
	return "", finalURL, status, reason, nil
}

func (p *SogouWeChatProvider) fetchHTML(ctx context.Context, target, referer string) (body string, finalURL string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/126 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	if strings.TrimSpace(referer) != "" {
		req.Header.Set("Referer", referer)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", resp.Request.URL.String(), fmt.Errorf("article body 状态异常: %d", resp.StatusCode)
	}
	rawBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", resp.Request.URL.String(), err
	}
	return string(rawBody), resp.Request.URL.String(), nil
}

func classifyArticleBodyFetchError(err error) (model.ArticleBodyFetchStatus, string) {
	if err == nil {
		return model.ArticleBodyFetchStatusFailed, ""
	}
	msg := strings.TrimSpace(err.Error())
	if strings.Contains(msg, "context deadline exceeded") || strings.Contains(msg, "Client.Timeout exceeded") {
		return model.ArticleBodyFetchStatusTimeout, msg
	}
	if strings.Contains(msg, "状态异常") {
		return model.ArticleBodyFetchStatusHTTPError, msg
	}
	return model.ArticleBodyFetchStatusFailed, msg
}

func classifyArticleBodyFetchHTML(body, finalURL string) (model.ArticleBodyFetchStatus, string) {
	lowerBody := strings.ToLower(body)
	lowerURL := strings.ToLower(finalURL)
	switch {
	case strings.Contains(lowerBody, "/antispider/") || strings.Contains(lowerBody, "antispider"):
		if strings.Contains(lowerURL, "weixin.sogou.com") || strings.Contains(body, "验证码") || strings.Contains(body, "VerifyCode") {
			return model.ArticleBodyFetchStatusSogouVerifyRequired, "搜狗返回验证码或反爬验证页"
		}
		return model.ArticleBodyFetchStatusWechatAntispider, "微信返回反爬验证页"
	case strings.Contains(body, "验证码") || strings.Contains(body, "VerifyCode"):
		return model.ArticleBodyFetchStatusSogouVerifyRequired, "源站返回验证码验证页"
	case isWeChatArticleURL(finalURL) && wechatVideoMarkerRe.MatchString(body):
		return model.ArticleBodyFetchStatusWechatVideoPage, "页面更像视频或富媒体内容，没有可解析的图文正文"
	case isWeChatArticleURL(finalURL):
		return model.ArticleBodyFetchStatusWechatNoJSContent, "微信公众号页面缺少 js_content 正文节点"
	case strings.Contains(lowerURL, "weixin.sogou.com"):
		return model.ArticleBodyFetchStatusSogouRedirectMissing, "搜狗跳转页没有返回可用的微信原文地址"
	default:
		return model.ArticleBodyFetchStatusFailed, "未能从页面解析正文"
	}
}

func (p *SogouWeChatProvider) approveInnerRedirect(ctx context.Context, body, referer string) {
	m := sogouInnerApproveRe.FindStringSubmatch(body)
	if len(m) < 3 {
		return
	}
	approveURL := "https://weixin.sogou.com/approve?uuid=" + url.QueryEscape(m[1]) + "&token=" + url.QueryEscape(m[2]) + "&from=inner"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, approveURL, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/126 Safari/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	if strings.TrimSpace(referer) != "" {
		req.Header.Set("Referer", referer)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
}

func extractSogouScriptRedirectURL(body string) string {
	matches := sogouURLAppendRe.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return ""
	}
	var b strings.Builder
	for _, m := range matches {
		if len(m) > 1 {
			b.WriteString(decodeSogouScriptURLFragment(m[1]))
		}
	}
	raw := strings.TrimSpace(b.String())
	if raw == "" {
		return ""
	}
	raw = strings.ReplaceAll(raw, "@", "")
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	if parsed.Host != "mp.weixin.qq.com" {
		return ""
	}
	return parsed.String()
}

func decodeSogouScriptURLFragment(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, `\/`, "/")
	value = strings.ReplaceAll(value, `\u0026`, "&")
	value = strings.ReplaceAll(value, `\U0026`, "&")
	value = strings.ReplaceAll(value, `\x26`, "&")
	value = strings.ReplaceAll(value, `\X26`, "&")
	value = strings.ReplaceAll(value, "&amp;", "&")
	value = strings.ReplaceAll(value, "&#38;", "&")
	value = strings.ReplaceAll(value, "&#x26;", "&")
	value = strings.ReplaceAll(value, "&#X26;", "&")
	return value
}

func prepareSogouClickURL(raw string) string {
	u := decodeHTMLText(raw)
	if !strings.Contains(u, "weixin.sogou.com/link") || strings.Contains(u, "&k=") || strings.Contains(u, "?k=") {
		return u
	}
	return prepareSogouClickURLWithK(u, randomSogouClickK())
}

func prepareSogouClickURLWithK(raw string, k int) string {
	u := decodeHTMLText(raw)
	if !strings.Contains(u, "weixin.sogou.com/link") || strings.Contains(u, "&k=") || strings.Contains(u, "?k=") {
		return u
	}
	if k < 1 {
		k = 1
	}
	if k > 100 {
		k = 100
	}
	idx := strings.Index(u, "url=")
	if idx < 0 {
		return u
	}
	hPos := idx + len("url=") + 21 + k
	if hPos < 0 || hPos >= len(u) {
		return u
	}
	separator := "&"
	if !strings.Contains(u, "?") {
		separator = "?"
	}
	return fmt.Sprintf("%s%sk=%d&h=%s", u, separator, k, url.QueryEscape(u[hPos:hPos+1]))
}

func randomSogouClickK() int {
	n, err := crand.Int(crand.Reader, big.NewInt(100))
	if err != nil {
		return int(time.Now().UnixNano()%100) + 1
	}
	return int(n.Int64()) + 1
}

func parseWeChatArticleBody(body string) string {
	if strings.Contains(body, "/antispider/") || strings.Contains(body, "antispider") {
		return ""
	}
	segment := extractHTMLNodeByID(body, "js_content")
	if strings.TrimSpace(segment) == "" {
		return ""
	}
	text := cleanArticleBodyText(segment)
	if len([]rune(text)) > 20000 {
		runes := []rune(text)
		text = string(runes[:20000])
	}
	return text
}

func extractHTMLNodeByID(body, id string) string {
	idIndex := strings.Index(body, `id="`+id+`"`)
	if idIndex < 0 {
		idIndex = strings.Index(body, `id='`+id+`'`)
	}
	if idIndex < 0 {
		return ""
	}
	start := strings.LastIndex(strings.ToLower(body[:idIndex]), "<div")
	if start < 0 {
		return ""
	}
	lower := strings.ToLower(body[start:])
	depth := 0
	for offset := 0; offset < len(lower); {
		nextOpen := strings.Index(lower[offset:], "<div")
		nextClose := strings.Index(lower[offset:], "</div")
		if nextOpen < 0 && nextClose < 0 {
			break
		}
		if nextOpen >= 0 && (nextClose < 0 || nextOpen < nextClose) {
			depth++
			offset += nextOpen + len("<div")
			continue
		}
		depth--
		offset += nextClose
		endTag := strings.Index(lower[offset:], ">")
		if endTag < 0 {
			return ""
		}
		offset += endTag + 1
		if depth <= 0 {
			return body[start : start+offset]
		}
	}
	return ""
}

func cleanArticleBodyText(value string) string {
	withBreaks := blockBreakRe.ReplaceAllString(value, "\n")
	text := htmlTagRe.ReplaceAllString(decodeHTMLText(withBreaks), "")
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(lineSpaceRe.ReplaceAllString(line, " "))
		if line != "" {
			out = append(out, line)
		}
	}
	return strings.TrimSpace(blankLineRe.ReplaceAllString(strings.Join(out, "\n"), "\n\n"))
}

func isWeChatArticleURL(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && strings.EqualFold(parsed.Host, "mp.weixin.qq.com")
}

func normalizeSogouURL(value string) string {
	decoded := decodeHTMLText(value)
	if strings.HasPrefix(decoded, "//") {
		return "https:" + decoded
	}
	if strings.HasPrefix(decoded, "/") {
		return "https://weixin.sogou.com" + decoded
	}
	return decoded
}

func stableSogouArticleHash(title, source string, publishedAt *time.Time, originalURL string) string {
	identity := strings.TrimSpace(title) + "|" + strings.TrimSpace(source)
	if publishedAt != nil {
		identity += "|" + publishedAt.UTC().Format(time.RFC3339)
	}
	if identity == "|" {
		return ArticleOriginalURLHash(originalURL)
	}
	sum := sha256.Sum256([]byte("sogou:" + identity))
	return hex.EncodeToString(sum[:])
}

func cleanHTMLText(value string) string {
	return strings.TrimSpace(spaceRe.ReplaceAllString(htmlTagRe.ReplaceAllString(decodeHTMLText(value), ""), " "))
}

func decodeHTMLText(value string) string {
	return html.UnescapeString(strings.TrimSpace(value))
}
