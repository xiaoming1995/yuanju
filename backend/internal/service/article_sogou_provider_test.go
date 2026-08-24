package service

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
	"yuanju/internal/model"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSogouWeChatProviderMapsSearchResults(t *testing.T) {
	results := parseSogouWeChatResults(`
<html>
<body>
	<li>
		<h3><a target="_blank" href="https://mp.weixin.qq.com/s?__biz=abc&amp;mid=123&amp;idx=1&amp;sn=def">八字格局入门</a></h3>
		<p class="txt-info">从月令和十神理解格局。</p>
		<a class="account">命理参考</a>
	</li>
</body>
</html>`, 10)
	if len(results) != 1 {
		t.Fatalf("len(results)=%d, want 1", len(results))
	}
	got := results[0]
	if got.Title != "八字格局入门" || got.SourceName != "命理参考" || got.SearchSnippet != "从月令和十神理解格局。" {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.OriginalURL != "https://mp.weixin.qq.com/s?__biz=abc&mid=123&idx=1&sn=def" {
		t.Fatalf("original_url=%q", got.OriginalURL)
	}
}

func TestSogouWeChatProviderUsesConfiguredSearchPage(t *testing.T) {
	seenPage := ""
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "weixin.sogou.com" {
			seenPage = req.URL.Query().Get("page")
		}
		return stringResponse(req, http.StatusOK, make(http.Header), `
<html><body>
	<li id="sogou_vr_11002601_box_0">
		<h3><a target="_blank" href="https://mp.weixin.qq.com/s?__biz=abc&amp;mid=123&amp;idx=1&amp;sn=def">八字分页测试</a></h3>
		<p class="txt-info">分页摘要。</p>
		<a class="account">命理参考</a>
	</li>
</body></html>`)
	})}

	provider := NewSogouWeChatProvider("https://weixin.sogou.com/weixin", client).WithSearchPageRange(3, 3)
	results, err := provider.Search(context.Background(), "八字", 1)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if seenPage != "3" {
		t.Fatalf("page query=%q, want 3", seenPage)
	}
	if len(results) != 1 || results[0].Title != "八字分页测试" {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestRandomSampleArticleResultsLimitsCandidates(t *testing.T) {
	results := make([]ArticleSearchResult, 10)
	for i := range results {
		results[i] = ArticleSearchResult{Title: "title-" + strconv.Itoa(i)}
	}
	got := randomSampleArticleResults(results, 3)
	if len(got) != 3 {
		t.Fatalf("len(got)=%d, want 3", len(got))
	}
	seen := map[string]bool{}
	for _, item := range got {
		if seen[item.Title] {
			t.Fatalf("sample contains duplicate title %q", item.Title)
		}
		seen[item.Title] = true
	}
}

func TestSogouWeChatProviderMapsCurrentSogouRedirectResults(t *testing.T) {
	results := parseSogouWeChatResults(`
<html>
<body>
	<li id="sogou_vr_11002601_box_0">
		<div class="img-box">
			<a target="_blank" href="/link?url=abc&amp;type=2&amp;query=%E5%85%AB%E5%AD%97&amp;token=token">
				<img src="//img01.sogoucdn.com/v2/thumb?appid=201147&amp;url=https%3A%2F%2Fmmbiz.qpic.cn%2Fcover.jpg&amp;sign=sign">
			</a>
		</div>
		<div class="txt-box">
			<h3>
				<a target="_blank" href="/link?url=abc&amp;type=2&amp;query=%E5%85%AB%E5%AD%97&amp;token=token">比&ldquo;苍山负雪&rdquo;更美的<em><!--red_beg-->八字<!--red_end--></em>短句</a>
			</h3>
			<p class="txt-info">奉上一组<em><!--red_beg-->八字<!--red_end--></em>短句。</p>
			<div class="s-p">
				<span class="all-time-y2">央视新闻</span><span class="s2"><script>document.write(timeConvert('1693143923'))</script></span>
			</div>
		</div>
	</li>
</body>
</html>`, 10)
	if len(results) != 1 {
		t.Fatalf("len(results)=%d, want 1", len(results))
	}
	got := results[0]
	if got.OriginalURL != "https://weixin.sogou.com/link?url=abc&type=2&query=%E5%85%AB%E5%AD%97&token=token" {
		t.Fatalf("original_url=%q", got.OriginalURL)
	}
	if got.CoverURL != "https://img01.sogoucdn.com/v2/thumb?appid=201147&url=https%3A%2F%2Fmmbiz.qpic.cn%2Fcover.jpg&sign=sign" {
		t.Fatalf("cover_url=%q", got.CoverURL)
	}
	if got.Title != "比“苍山负雪”更美的八字短句" || got.SourceName != "央视新闻" || got.SearchSnippet != "奉上一组八字短句。" {
		t.Fatalf("unexpected result: %+v", got)
	}
	wantTime := time.Unix(1693143923, 0).UTC()
	if got.PublishedAtSource == nil || !got.PublishedAtSource.Equal(wantTime) {
		t.Fatalf("published_at_source=%v, want %v", got.PublishedAtSource, wantTime)
	}
}

func TestSogouWeChatProviderPrefersStructuredArticleItems(t *testing.T) {
	results := parseSogouWeChatResults(`
<html>
<body>
	<li class="nav-item">
		<h3><a href="https://example.com/not-an-article">导航项</a></h3>
		<p class="txt-info">这不是微信文章搜索结果。</p>
	</li>
	<li id="sogou_vr_11002601_box_0">
		<div class="txt-box">
			<h3><a target="_blank" href="/link?url=abc&amp;type=2&amp;query=%E5%85%AB%E5%AD%97&amp;token=token">八字文章</a></h3>
			<p class="txt-info">这是搜索摘要。</p>
			<div class="s-p"><span class="all-time-y2">命理号</span></div>
		</div>
	</li>
</body>
</html>`, 10)
	if len(results) != 1 {
		t.Fatalf("len(results)=%d, want 1", len(results))
	}
	if results[0].Title != "八字文章" || results[0].SourceName != "命理号" {
		t.Fatalf("unexpected result: %+v", results[0])
	}
}

func TestSogouWeChatProviderApprovesSearchSessionBeforeFetchingBody(t *testing.T) {
	approved := false
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		header := make(http.Header)
		switch req.URL.Path {
		case "/weixin":
			return stringResponse(req, http.StatusOK, header, `
<html>
<script>
	var uuid = "session-uuid";
	var ssToken = "session-token";
</script>
<body>
	<ul class="news-list">
		<li id="sogou_vr_11002601_box_0">
			<div class="txt-box">
				<h3><a target="_blank" href="https://weixin.sogou.com/link?url=abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz&amp;type=2">八字正文测试</a></h3>
				<p class="txt-info">搜索摘要。</p>
				<div class="s-p"><span class="all-time-y2">命理号</span></div>
			</div>
		</li>
	</ul>
</body>
</html>`)
		case "/approve":
			if req.URL.Query().Get("uuid") == "session-uuid" && req.URL.Query().Get("token") == "session-token" && req.URL.Query().Get("from") == "search" {
				approved = true
				header.Add("Set-Cookie", "SNUID=ok; Path=/")
			}
			return stringResponse(req, http.StatusNoContent, header, "")
		case "/link":
			cookie, _ := req.Cookie("SNUID")
			if !approved || cookie == nil || cookie.Value != "ok" {
				return stringResponse(req, http.StatusOK, header, `<html><a href="/antispider/">验证</a></html>`)
			}
			return stringResponse(req, http.StatusOK, header, `<html><body><div id="js_content"><p>会话正文第一段</p><p>会话正文第二段</p></div></body></html>`)
		default:
			return stringResponse(req, http.StatusNotFound, header, "")
		}
	})}

	provider := NewSogouWeChatProvider("https://weixin.sogou.com/weixin", client)
	results, err := provider.Search(context.Background(), "八字", 1)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results)=%d, want 1", len(results))
	}
	if !results[0].FullTextAuthorized || !strings.Contains(results[0].BodyContent, "会话正文第一段") {
		t.Fatalf("body was not fetched through approved session: %+v", results[0])
	}
	if results[0].BodyFetchStatus != model.ArticleBodyFetchStatusSucceeded || results[0].BodyFetchError != "" {
		t.Fatalf("body fetch diagnostics=%s/%q, want succeeded", results[0].BodyFetchStatus, results[0].BodyFetchError)
	}
}

func TestSogouWeChatProviderFollowsSogouInnerScriptRedirect(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		header := make(http.Header)
		switch req.URL.Host + req.URL.Path {
		case "weixin.sogou.com/weixin":
			return stringResponse(req, http.StatusOK, header, `
<html>
<script>var uuid = "session-uuid"; var ssToken = "session-token";</script>
<body>
	<li id="sogou_vr_11002601_box_0">
		<h3><a target="_blank" href="https://weixin.sogou.com/link?url=abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz&amp;type=2">八字正文测试</a></h3>
		<p class="txt-info">搜索摘要。</p>
		<span class="all-time-y2">命理号</span>
	</li>
</body>
</html>`)
		case "weixin.sogou.com/approve":
			return stringResponse(req, http.StatusNoContent, header, "")
		case "weixin.sogou.com/link":
			return stringResponse(req, http.StatusOK, header, `
<meta content="always" name="referrer">
<script>
    (new Image()).src = 'https://weixin.sogou.com/approve?uuid=' + 'inner-uuid' + '&token=' + 'inner-token' + '&from=inner';
    setTimeout(function () {
        var url = '';
        url += 'https://mp.';
        url += 'weixin.qq.c';
        url += 'om/s?src=11';
        url += '&timestamp=1783412174&';
        url += 'signature=abc&new=1';
        window.location.replace(url)
    },100);
</script>`)
		case "mp.weixin.qq.com/s":
			return stringResponse(req, http.StatusOK, header, `<html><body><div id="js_content"><p>脚本跳转后的正文</p></div></body></html>`)
		default:
			return stringResponse(req, http.StatusNotFound, header, "")
		}
	})}

	provider := NewSogouWeChatProvider("https://weixin.sogou.com/weixin", client)
	results, err := provider.Search(context.Background(), "八字", 1)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results)=%d, want 1", len(results))
	}
	if !results[0].FullTextAuthorized || results[0].BodyContent != "脚本跳转后的正文" {
		t.Fatalf("body was not fetched through script redirect: %+v", results[0])
	}
	if results[0].BodyFetchStatus != model.ArticleBodyFetchStatusSucceeded || results[0].BodyFetchError != "" {
		t.Fatalf("body fetch diagnostics=%s/%q, want succeeded", results[0].BodyFetchStatus, results[0].BodyFetchError)
	}
	if !strings.HasPrefix(results[0].OriginalURL, "https://mp.weixin.qq.com/s?") {
		t.Fatalf("original_url=%q, want mp.weixin redirect", results[0].OriginalURL)
	}
	if !strings.Contains(results[0].OriginalURL, "&timestamp=") {
		t.Fatalf("original_url=%q, want timestamp query to be preserved", results[0].OriginalURL)
	}
}

func TestSogouWeChatProviderRecordsVerifyPageBodyFailure(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		header := make(http.Header)
		switch req.URL.Host + req.URL.Path {
		case "weixin.sogou.com/weixin":
			return stringResponse(req, http.StatusOK, header, `
<html><body>
	<li id="sogou_vr_11002601_box_0">
		<h3><a target="_blank" href="https://weixin.sogou.com/link?url=abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz&amp;type=2">八字正文测试</a></h3>
		<p class="txt-info">搜索摘要。</p>
		<span class="all-time-y2">命理号</span>
	</li>
</body></html>`)
		case "weixin.sogou.com/link":
			return stringResponse(req, http.StatusOK, header, `<html><body><form id="VerifyCode">请输入验证码</form></body></html>`)
		default:
			return stringResponse(req, http.StatusNotFound, header, "")
		}
	})}

	provider := NewSogouWeChatProvider("https://weixin.sogou.com/weixin", client)
	results, err := provider.Search(context.Background(), "八字", 1)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results)=%d, want 1", len(results))
	}
	if results[0].FullTextAuthorized || results[0].BodyContent != "" {
		t.Fatalf("body should not be authorized: %+v", results[0])
	}
	if results[0].BodyFetchStatus != model.ArticleBodyFetchStatusSogouVerifyRequired || !strings.Contains(results[0].BodyFetchError, "验证") {
		t.Fatalf("body fetch diagnostics=%s/%q, want sogou verify", results[0].BodyFetchStatus, results[0].BodyFetchError)
	}
}

func TestSogouWeChatProviderRecordsWechatVideoPageBodyFailure(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		header := make(http.Header)
		switch req.URL.Host + req.URL.Path {
		case "weixin.sogou.com/weixin":
			return stringResponse(req, http.StatusOK, header, `
<html><body>
	<li id="sogou_vr_11002601_box_0">
		<h3><a target="_blank" href="https://mp.weixin.qq.com/s?__biz=abc&amp;mid=1&amp;idx=1&amp;sn=video">视频文章</a></h3>
		<p class="txt-info">搜索摘要。</p>
		<span class="all-time-y2">命理号</span>
	</li>
</body></html>`)
		case "mp.weixin.qq.com/s":
			return stringResponse(req, http.StatusOK, header, `<html><body><div class="js_page_video">视频号内容</div></body></html>`)
		default:
			return stringResponse(req, http.StatusNotFound, header, "")
		}
	})}

	provider := NewSogouWeChatProvider("https://weixin.sogou.com/weixin", client)
	results, err := provider.Search(context.Background(), "八字", 1)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results)=%d, want 1", len(results))
	}
	if results[0].BodyFetchStatus != model.ArticleBodyFetchStatusWechatVideoPage || !strings.Contains(results[0].BodyFetchError, "视频") {
		t.Fatalf("body fetch diagnostics=%s/%q, want wechat video page", results[0].BodyFetchStatus, results[0].BodyFetchError)
	}
}

func TestExtractSogouScriptRedirectURLPreservesTimestampQuery(t *testing.T) {
	got := extractSogouScriptRedirectURL(`
<script>
	var url = '';
	url += 'https://mp.weixin.qq.com/s?src=11';
	url += '&timestamp=1783422961&ver=6828';
	url += '&signature=abc&new=1';
</script>`)
	want := "https://mp.weixin.qq.com/s?src=11&timestamp=1783422961&ver=6828&signature=abc&new=1"
	if got != want {
		t.Fatalf("redirect url=%q, want %q", got, want)
	}
}

func stringResponse(req *http.Request, status int, header http.Header, body string) (*http.Response, error) {
	if header == nil {
		header = make(http.Header)
	}
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func TestStableSogouArticleHashIgnoresVolatileRedirectToken(t *testing.T) {
	publishedAt := time.Unix(1693143923, 0).UTC()
	first := stableSogouArticleHash("八字短句", "央视新闻", &publishedAt, "https://weixin.sogou.com/link?url=a&token=1")
	second := stableSogouArticleHash("八字短句", "央视新闻", &publishedAt, "https://weixin.sogou.com/link?url=b&token=2")
	if first != second {
		t.Fatalf("stable hash changed across sogou redirect tokens")
	}
}

func TestParseWeChatArticleBodyExtractsJsContentText(t *testing.T) {
	body := parseWeChatArticleBody(`
<html>
<body>
	<h1 id="activity-name">标题</h1>
	<div id="js_content">
		<section>
			<p>第一段<strong>重点</strong></p>
			<p>第二段<br>换行</p>
		</section>
	</div>
	<script>window.foo = true</script>
</body>
</html>`)
	want := "第一段重点\n第二段\n换行"
	if body != want {
		t.Fatalf("body=%q, want %q", body, want)
	}
}

func TestParseWeChatArticleBodySkipsAntispiderPage(t *testing.T) {
	if body := parseWeChatArticleBody(`<html><a href="/antispider/">验证</a><div id="js_content">正文</div></html>`); body != "" {
		t.Fatalf("antispider body=%q, want empty", body)
	}
}

func TestPrepareSogouClickURLAddsKHParams(t *testing.T) {
	raw := "https://weixin.sogou.com/link?url=abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz&type=2"
	got := prepareSogouClickURLWithK(raw, 50)
	if got == raw || !strings.Contains(got, "&k=50&h=") {
		t.Fatalf("prepareSogouClickURL=%q", got)
	}
}
