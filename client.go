package quranc

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/jsteenb2/httpc"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type clientOpt struct {
	host string
}

type ClientOptFn func(opt clientOpt) clientOpt

func ClientHost(host string) ClientOptFn {
	return func(opt clientOpt) clientOpt {
		opt.host = host
		return opt
	}
}

type Client struct {
	c *httpc.Client
}

func New(doer Doer, opts ...ClientOptFn) *Client {
	opt := clientOpt{
		host: "http://staging.quran.com:3000",
	}
	for _, o := range opts {
		opt = o(opt)
	}

	baseURL := opt.host + "/api/v3"
	return &Client{
		c: httpc.New(doer, httpc.WithBaseURL(baseURL)),
	}
}

type Recitation struct {
	ID                    int    `json:"id"`
	Style                 string `json:"style"`
	ReciterNameEng        string `json:"reciter_name_eng"`
	ReciterNameTranslated string `json:"reciter_name_translated"`
}

type TranslatedName struct {
	LanguageName string `json:"language_name"`
	Name         string `json:"name"`
}

func (c *Client) Recitations(ctx context.Context, reqOpts ...ReqOptFn) ([]Recitation, error) {
	var opt reqOpt
	for _, optFn := range reqOpts {
		opt = optFn(opt)
	}

	var resp struct {
		Recitations []Recitation `json:"recitations"`
	}
	req := c.c.Get("/options/recitations")
	err := opt.applyQueryParams(req).
		Success(httpc.StatusOK()).
		DecodeJSON(&resp).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Recitations, nil
}

type Translation struct {
	ID           int    `json:"id"`
	AuthorName   string `json:"author_name"`
	LanguageName string `json:"language_name"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
}

func (c *Client) Translations(ctx context.Context, reqOpts ...ReqOptFn) ([]Translation, error) {
	var opt reqOpt
	for _, optFn := range reqOpts {
		opt = optFn(opt)
	}

	var resp struct {
		Translations []Translation `json:"translations"`
	}
	req := c.c.Get("/options/translations")
	err := opt.applyQueryParams(req).
		Success(httpc.StatusOK()).
		DecodeJSON(&resp).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(resp.Translations, func(i, j int) bool {
		return resp.Translations[i].ID < resp.Translations[j].ID
	})

	return resp.Translations, nil
}

type Language struct {
	ID              int              `json:"id"`
	Name            string           `json:"name"`
	IsoCode         string           `json:"iso_code"`
	NativeName      string           `json:"native_name"`
	Direction       string           `json:"direction"`
	TranslatedNames []TranslatedName `json:"translated_names"`
}

func (c *Client) Languages(ctx context.Context, reqOpts ...ReqOptFn) ([]Language, error) {
	var opt reqOpt
	for _, optFn := range reqOpts {
		opt = optFn(opt)
	}

	var resp struct {
		Languages []Language `json:"languages"`
	}
	req := c.c.Get("/options/languages")
	err := opt.applyQueryParams(req).
		Success(httpc.StatusOK()).
		DecodeJSON(&resp).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(resp.Languages, func(i, j int) bool {
		return resp.Languages[i].ID < resp.Languages[j].ID
	})

	return resp.Languages, nil
}

type Tafsir struct {
	ID           int    `json:"id"`
	AuthorName   string `json:"author_name"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	LanguageName string `json:"language_name"`
}

func (c *Client) Tafsiraat(ctx context.Context, reqOpts ...ReqOptFn) ([]Tafsir, error) {
	var opt reqOpt
	for _, optFn := range reqOpts {
		opt = optFn(opt)
	}

	var resp struct {
		Tafsirs []Tafsir `json:"tafsirs"`
	}
	req := c.c.Get("/options/tafsirs")
	err := opt.applyQueryParams(req).
		Success(httpc.StatusOK()).
		DecodeJSON(&resp).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(resp.Tafsirs, func(i, j int) bool {
		return resp.Tafsirs[i].ID < resp.Tafsirs[j].ID
	})

	return resp.Tafsirs, nil
}

type Chapter struct {
	ID              int    `json:"id"`
	ChapterNumber   int    `json:"chapter_number"`
	BismillahPre    bool   `json:"bismillah_pre"`
	RevelationOrder int    `json:"revelation_order"`
	RevelationPlace string `json:"revelation_place"`
	NameComplex     string `json:"name_complex"`
	NameArabic      string `json:"name_arabic"`
	NameSimple      string `json:"name_simple"`
	VersesCount     int    `json:"verses_count"`
	Pages           struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"pages"`
	TranslatedName TranslatedName `json:"translated_name"`
}

type apiChapter struct {
	ID              int            `json:"id"`
	ChapterNumber   int            `json:"chapter_number"`
	BismillahPre    bool           `json:"bismillah_pre"`
	RevelationOrder int            `json:"revelation_order"`
	RevelationPlace string         `json:"revelation_place"`
	NameArabic      string         `json:"name_arabic"`
	NameComplex     string         `json:"name_complex"`
	NameSimple      string         `json:"name_simple"`
	VersesCount     int            `json:"verses_count"`
	Pages           []int          `json:"pages"`
	TranslatedName  TranslatedName `json:"translated_name"`
}

func apiChapterToChapter(ch apiChapter) Chapter {
	return Chapter{
		ID:              ch.ID,
		ChapterNumber:   ch.ChapterNumber,
		BismillahPre:    ch.BismillahPre,
		RevelationOrder: ch.RevelationOrder,
		RevelationPlace: ch.RevelationPlace,
		NameArabic:      ch.NameArabic,
		NameComplex:     ch.NameComplex,
		NameSimple:      ch.NameSimple,
		VersesCount:     ch.VersesCount,
		Pages: struct {
			Start int `json:"start"`
			End   int `json:"end"`
		}{
			Start: ch.Pages[0],
			End:   ch.Pages[1],
		},
		TranslatedName: ch.TranslatedName,
	}
}

func (c *Client) Chapters(ctx context.Context, reqOpts ...ReqOptFn) ([]Chapter, error) {
	var opt reqOpt
	for _, optFn := range reqOpts {
		opt = optFn(opt)
	}

	var resp struct {
		Chapters []apiChapter `json:"chapters"`
	}
	req := c.c.Get("/chapters")
	err := opt.applyQueryParams(req).
		Success(httpc.StatusOK()).
		DecodeJSON(&resp).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	chapters := make([]Chapter, len(resp.Chapters))
	for i, ch := range resp.Chapters {
		chapters[i] = apiChapterToChapter(ch)
	}

	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].ChapterNumber < chapters[j].ChapterNumber
	})

	return chapters, nil
}

func (c *Client) Chapter(ctx context.Context, id int, reqOpts ...ReqOptFn) (Chapter, error) {
	var opt reqOpt
	for _, optFn := range reqOpts {
		opt = optFn(opt)
	}

	var resp struct {
		Chapter apiChapter `json:"chapter"`
	}
	req := c.c.Get("/chapters/" + strconv.Itoa(id))
	err := opt.applyQueryParams(req).
		Success(httpc.StatusOK()).
		DecodeJSON(&resp).
		Do(ctx)
	if err != nil {
		return Chapter{}, err
	}

	return apiChapterToChapter(resp.Chapter), nil
}

type ChapterInfo struct {
	ChapterID    int    `json:"chapter_id"`
	Text         string `json:"text"`
	Source       string `json:"source"`
	ShortText    string `json:"short_text"`
	LanguageName string `json:"language_name"`
}

func (c *Client) ChapterInfo(ctx context.Context, id int, reqOpts ...ReqOptFn) (ChapterInfo, error) {
	var opt reqOpt
	for _, optFn := range reqOpts {
		opt = optFn(opt)
	}

	var resp struct {
		ChapterInfo ChapterInfo `json:"chapter_info"`
	}
	req := c.c.Get("/chapters/" + strconv.Itoa(id) + "/info")
	err := opt.applyQueryParams(req).
		Success(httpc.StatusOK()).
		DecodeJSON(&resp).
		Do(ctx)
	if err != nil {
		return ChapterInfo{}, err
	}

	return resp.ChapterInfo, nil
}

type (
	ReqOptFn func(opt reqOpt) reqOpt

	reqOpt struct {
		languageID int
	}
)

func (o reqOpt) applyQueryParams(r *httpc.Request) *httpc.Request {
	if o.languageID > 0 {
		r = r.QueryParam("language", strconv.Itoa(o.languageID))
	}

	return r
}

func LanguageID(id int) ReqOptFn {
	return func(opt reqOpt) reqOpt {
		opt.languageID = id
		return opt
	}
}

type Resource struct {
	ID           int    `json:"id"`
	LanguageName string `json:"language_name"`
	Text         string `json:"text"`
	ResourceName string `json:"resource_name"`
	ResourceID   int    `json:"resource_id"`
}

type Verse struct {
	ID           int    `json:"id"`
	VerseNumber  int    `json:"verse_number"`
	ChapterID    int    `json:"chapter_id"`
	VerseKey     string `json:"verse_key"`
	TextMadani   string `json:"text_madani"`
	TextIndopak  string `json:"text_indopak"`
	TextSimple   string `json:"text_simple"`
	JuzNumber    int    `json:"juz_number"`
	HizbNumber   int    `json:"hizb_number"`
	RubNumber    int    `json:"rub_number"`
	Sajdah       string `json:"sajdah"`
	SajdahNumber int    `json:"sajdah_number"`
	PageNumber   int    `json:"page_number"`
	Audio        struct {
		URL      string     `json:"url"`
		Duration int        `json:"duration"`
		Segments [][]string `json:"segments"`
		Format   string     `json:"format"`
	} `json:"audio"`
	Translations  []Resource `json:"translations"`
	MediaContents []struct {
		URL        string `json:"url"`
		EmbedText  string `json:"embed_text"`
		Provider   string `json:"provider"`
		AuthorName string `json:"author_name"`
	} `json:"media_contents"`
	Words []Word `json:"words"`
}

type Word struct {
	ID          int    `json:"id"`
	Position    int    `json:"position"`
	TextMadani  string `json:"text_madani"`
	TextIndopak string `json:"text_indopak"`
	TextSimple  string `json:"text_simple"`
	VerseKey    string `json:"verse_key"`
	ClassName   string `json:"class_name"`
	LineNumber  int    `json:"line_number"`
	PageNumber  int    `json:"page_number"`
	Code        string `json:"code"`
	CodeV3      string `json:"code_v3"`
	CharType    string `json:"char_type"`
	Audio       struct {
		URL string `json:"url"`
	} `json:"audio"`
	Translation     Resource `json:"translation"`
	Transliteration Resource `json:"transliteration"`
}

type (
	VersesReqOptFn func(opts versesReqOpt) versesReqOpt

	versesReqOpt struct {
		language   string
		recitation int
		textType   string

		page   int
		limit  int
		offset int

		media        []int
		translations []int
	}
)

func (v versesReqOpt) queryParams(r *httpc.Request) *httpc.Request {
	if v.language != "" {
		r = r.QueryParam("language", v.language)
	}

	if v.recitation > 0 {
		r = r.QueryParam("recitation", strconv.Itoa(v.recitation))
	}

	if v.textType != "" {
		r = r.QueryParam("text_type", v.textType)
	}

	if v.page > 0 {
		r = r.QueryParam("page", strconv.Itoa(v.page))
	}

	if v.limit > 0 {
		r = r.QueryParam("limit", strconv.Itoa(v.limit))
	}

	for _, media := range v.media {
		r = r.QueryParam("media[]", strconv.Itoa(media))
	}

	for _, translation := range v.translations {
		r = r.QueryParam("translations[]", strconv.Itoa(translation))
	}

	return r
}

func VersesLanguage(isoCode string) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.language = isoCode
		return opts
	}
}

func VersesRecitation(recitation int) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.recitation = recitation
		return opts
	}
}

func VersesTextType(textType string) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.textType = textType
		return opts
	}
}

func VersesLimit(i int) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.limit = i
		return opts
	}
}

func VersesOffset(i int) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.offset = i
		return opts
	}
}

func VersesPage(i int) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.page = i
		return opts
	}
}

func VersesMedia(media []int) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.media = media
		return opts
	}
}

func VersesTranslations(translations []int) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.translations = translations
		return opts
	}
}

func (c *Client) Verses(ctx context.Context, chapterID int, reqOpts ...VersesReqOptFn) ([]Verse, error) {
	var opts versesReqOpt
	for _, optFn := range reqOpts {
		opts = optFn(opts)
	}

	req := c.c.Get("/chapters/" + strconv.Itoa(chapterID) + "/verses")
	req = opts.queryParams(req)

	var resp struct {
		Verses []Verse `json:"verses"`
		Meta   struct {
			CurrentPage int         `json:"current_page"`
			NextPage    int         `json:"next_page"`
			PrevPage    interface{} `json:"prev_page"`
			TotalPages  int         `json:"total_pages"`
			TotalCount  int         `json:"total_count"`
		} `json:"meta"`
	}
	err := req.
		Success(httpc.StatusOK()).
		DecodeJSON(&resp).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	return resp.Verses, nil
}

// TODO: make github issue to fix the route in api docs for this route is routed incorrectly
func (c *Client) Verse(ctx context.Context, chapterID, verseID int) (Verse, error) {
	var resp struct {
		Verse Verse `json:"verse"`
	}

	endpoint := "/chapters/" + strconv.Itoa(chapterID) + "/verses/" + strconv.Itoa(verseID)
	err := c.c.Get(endpoint).
		Success(httpc.StatusOK()).
		DecodeJSON(&resp).
		Do(ctx)
	if err != nil {
		return Verse{}, err
	}

	return resp.Verse, nil
}

type Juz struct {
	ID           int          `json:"id"`
	JuzNumber    int          `json:"juz_number"`
	VerseMapping []JuzMapping `json:"verse_mapping"`
}

type JuzMapping struct {
	ChapterID  int `json:"chapter_id"`
	StartVerse int `json:"start_verse"`
	EndVerse   int `json:"end_verse"`
}

type apiJuz struct {
	ID           int               `json:"id"`
	JuzNumber    int               `json:"juz_number"`
	VerseMapping map[string]string `json:"verse_mapping"`
}

func convertAPIJuzToJuz(j apiJuz) Juz {
	strToInt := func(s string) int {
		i, err := strconv.Atoi(s)
		if err != nil {
			return -1
		}
		return i
	}

	juz := Juz{
		ID:        j.ID,
		JuzNumber: j.JuzNumber,
	}

	for chapterID, ayaat := range j.VerseMapping {
		startEnds := strings.Split(ayaat, "-")
		if len(startEnds) != 2 {
			continue
		}

		juz.VerseMapping = append(juz.VerseMapping, JuzMapping{
			ChapterID:  strToInt(chapterID),
			StartVerse: strToInt(startEnds[0]),
			EndVerse:   strToInt(startEnds[1]),
		})
	}

	sort.Slice(juz.VerseMapping, func(i, j int) bool {
		return juz.VerseMapping[i].ChapterID < juz.VerseMapping[j].ChapterID
	})

	return juz
}

func (c *Client) Juzzah(ctx context.Context) ([]Juz, error) {
	var resp struct {
		Juzzah []struct {
			ID           int               `json:"id"`
			JuzNumber    int               `json:"juz_number"`
			VerseMapping map[string]string `json:"verse_mapping"`
		} `json:"juzs"`
	}
	err := c.c.Get("/juzs").
		Success(httpc.StatusOK()).
		DecodeJSON(&resp).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	juzzah := make([]Juz, len(resp.Juzzah))
	for i, aj := range resp.Juzzah {
		juzzah[i] = convertAPIJuzToJuz(aj)
	}

	return juzzah, nil
}

type VerseTafsir struct {
	ID           int    `json:"id"`
	Text         string `json:"text"`
	VerseID      int    `json:"verse_id"`
	LanguageName string `json:"language_name"`
	ResourceName string `json:"resource_name"`

	// VerseKey  outlined int he api response, but there is nothing that speaks
	// to it in the api documentation... hopefully someone can fill in the gap here
	VerseKey interface{} `json:"verse_key"`
}

type (
	VerseTafsirReqOptFn func(opts verseTafsirReqOpts) verseTafsirReqOpts

	verseTafsirReqOpts struct {
		Tafsir string
	}
)

func TafsirID(id int) VerseTafsirReqOptFn {
	return func(opts verseTafsirReqOpts) verseTafsirReqOpts {
		opts.Tafsir = strconv.Itoa(id)
		return opts
	}
}

func (c *Client) VerseTafsir(ctx context.Context, chapterID, verseID int, reqOpts ...VerseTafsirReqOptFn) ([]VerseTafsir, error) {
	var opts verseTafsirReqOpts
	for _, optFn := range reqOpts {
		opts = optFn(opts)
	}

	endpoint := "/chapters/" + strconv.Itoa(chapterID) + "/verses/" + strconv.Itoa(verseID) + "/tafsirs"
	req := c.c.Get(endpoint)

	if opts.Tafsir != "" {
		req = req.QueryParam("tafsirs", opts.Tafsir)
	}

	var resp struct {
		Tafsirs []VerseTafsir `json:"tafsirs"`
	}
	err := req.
		Success(httpc.StatusOK()).
		DecodeJSON(&resp).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	return resp.Tafsirs, nil
}

type (
	SearchRequest struct {
		Query    string
		Language string
		Page     int
		Size     int
	}

	SearchResponse struct {
		Query       string        `json:"query"`
		TotalCount  int           `json:"total_count"`
		Took        int           `json:"took"`
		CurrentPage int           `json:"current_page"`
		TotalPages  int           `json:"total_pages"`
		PerPage     int           `json:"per_page"`
		Results     []SearchVerse `json:"results"`
	}

	SearchVerse struct {
		ID           int        `json:"id"`
		VerseNumber  int        `json:"verse_number"`
		ChapterID    int        `json:"chapter_id"`
		VerseKey     string     `json:"verse_key"`
		TextMadani   string     `json:"text_madani"`
		Words        []Word     `json:"words"`
		Translations []Resource `json:"translations"`
	}
)

func (c *Client) Search(ctx context.Context, query SearchRequest) (SearchResponse, error) {
	if query.Query == "" {
		return SearchResponse{}, errors.New("no query param provided")
	}

	req := c.c.Get("/search").
		QueryParam("q", query.Query)
	if query.Language != "" {
		req = req.QueryParam("language", query.Language)
	}
	if query.Page > 0 {
		req = req.QueryParam("page", strconv.Itoa(query.Page))
	}
	if query.Size > 0 {
		req = req.QueryParam("size", strconv.Itoa(query.Size))
	}

	var resp SearchResponse
	err := req.
		Success(httpc.StatusOK()).
		DecodeJSON(&resp).
		Do(ctx)
	if err != nil {
		return SearchResponse{}, err
	}

	return resp, nil
}
