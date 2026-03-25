package applib

import (
	"net/url"
	"regexp"
	"strings"
)

// URLFilter решает, допускать ли кандидата во frontier.
type URLFilter interface {
	Allow(c URLCandidate) bool
}

// ChainFilter применяет фильтры по порядку: первый отказ — skip (AND-логика).
type ChainFilter []URLFilter

func (cf ChainFilter) Allow(c URLCandidate) bool {
	for _, f := range cf {
		if !f.Allow(c) {
			return false
		}
	}
	return true
}

// FilterConfig описывает настройки стандартных фильтров.
type FilterConfig struct {
	MaxQueryParams int // максимальное число query-параметров в URL (default: 4)
}

func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		MaxQueryParams: 4,
	}
}

// NewDefaultChainFilter создаёт стандартный набор фильтров.
func NewDefaultChainFilter(cfg FilterConfig) ChainFilter {
	return ChainFilter{
		NewExtensionFilter(),
		NewTrapFilter(),
		NewQueryParamFilter(cfg.MaxQueryParams),
		NewContentFilter(),
	}
}

// --- ExtensionFilter --------------------------------------------------------

// extensionBlocklist — медиа и архивы без полезного текстового контента.
var extensionBlocklist = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".svg": true, ".webp": true, ".ico": true,
	".mp4": true, ".mp3": true, ".avi": true, ".mov": true, ".webm": true,
	".zip": true, ".tar": true, ".gz": true, ".rar": true,
	".exe": true, ".dmg": true, ".apk": true,
	".css": true, ".js": true, ".map": true,
	".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
}

// ExtensionFilter блокирует медиа-файлы и архивы.
// Документы (pdf, docx и пр.) не блокируются — они уже выделяются отдельным путём
// через enqueueDocuments и не попадают в enqueueLinks.
type ExtensionFilter struct{}

func NewExtensionFilter() *ExtensionFilter { return &ExtensionFilter{} }

func (f *ExtensionFilter) Allow(c URLCandidate) bool {
	raw := strings.ToLower(c.URL)
	// Берём только путь (убираем query/fragment)
	if idx := strings.IndexAny(raw, "?#"); idx >= 0 {
		raw = raw[:idx]
	}
	for ext := range extensionBlocklist {
		if strings.HasSuffix(raw, ext) {
			return false
		}
	}
	return true
}

// --- TrapFilter -------------------------------------------------------------

var (
	paginationTrapRe = regexp.MustCompile(`(?i)[?&/](?:page|p)[=/](\d+)`)
	calendarTrapRe   = regexp.MustCompile(`(?i)/(?:calendar|events?)/\d{4}`)
	cdnCgiRe         = regexp.MustCompile(`(?i)/cdn-cgi/`)
)

// trapQueryParams — query-параметры, сигнализирующие о spider trap.
var trapQueryParams = map[string]bool{
	"sort": true, "order": true, "dir": true,
	"asc": true, "desc": true, "view": true,
	"filter": true,
}

// TrapFilter блокирует известные spider traps: бесконечные пагинации,
// календарные пути, служебные пути CDN, параметрические ловушки.
type TrapFilter struct{}

func NewTrapFilter() *TrapFilter { return &TrapFilter{} }

func (f *TrapFilter) Allow(c URLCandidate) bool {
	// CDN служебные пути
	if cdnCgiRe.MatchString(c.URL) {
		return false
	}
	// Бесконечные календарные пути
	if calendarTrapRe.MatchString(c.URL) {
		return false
	}
	// Пагинация с номером > 50
	if m := paginationTrapRe.FindStringSubmatch(c.URL); m != nil {
		// Если номер страницы числовой и > 50 — trap
		n := 0
		for _, ch := range m[1] {
			n = n*10 + int(ch-'0')
		}
		if n > 50 {
			return false
		}
	}
	// Trap query params
	parsed, err := url.Parse(c.URL)
	if err != nil {
		return true // не можем разобрать — пропускаем
	}
	for key := range parsed.Query() {
		if trapQueryParams[strings.ToLower(key)] {
			return false
		}
	}
	return true
}

// --- QueryParamFilter -------------------------------------------------------

// QueryParamFilter блокирует URL с числом query-параметров сверх лимита.
// URL с >N параметрами часто генерируются динамически и являются ловушками.
type QueryParamFilter struct {
	max int
}

func NewQueryParamFilter(max int) *QueryParamFilter {
	return &QueryParamFilter{max: max}
}

func (f *QueryParamFilter) Allow(c URLCandidate) bool {
	parsed, err := url.Parse(c.URL)
	if err != nil {
		return true
	}
	return len(parsed.Query()) <= f.max
}

// --- ContentFilter ----------------------------------------------------------

// contentBlockPatterns — пути со служебным / малоценным контентом.
var contentBlockPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)/(login|signin|sign-in|signup|sign-up|register)(/?$|[/?#])`),
	regexp.MustCompile(`(?i)/(logout|signout)(/?$|[/?#])`),
	regexp.MustCompile(`(?i)/(cart|checkout|basket|order)(/?$|[/?#])`),
	regexp.MustCompile(`(?i)/print(/|$)`),
	regexp.MustCompile(`(?i)/pdf(/|$)`),
	regexp.MustCompile(`(?i)/feed(\.xml|/|$)`),
	regexp.MustCompile(`(?i)/rss(\.xml|/|$)`),
	regexp.MustCompile(`(?i)/sitemap(\.xml|/|$)`),
	regexp.MustCompile(`(?i)/robots\.txt`),
}

// ContentFilter блокирует служебные страницы: авторизация, корзина,
// print-версии, RSS/Atom ленты.
type ContentFilter struct{}

func NewContentFilter() *ContentFilter { return &ContentFilter{} }

func (f *ContentFilter) Allow(c URLCandidate) bool {
	parsed, err := url.Parse(c.URL)
	if err != nil {
		return true
	}
	path := parsed.Path
	for _, re := range contentBlockPatterns {
		if re.MatchString(path) {
			return false
		}
	}
	return true
}
