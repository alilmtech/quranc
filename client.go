package quranc

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jsteenb2/httpc"
)

type QuranAPI interface {
	Recitations(ctx context.Context, reqOpts ...ReqOptFn) ([]Recitation, error)
	Translations(ctx context.Context, reqOpts ...ReqOptFn) ([]Translation, error)
	Languages(ctx context.Context, reqOpts ...ReqOptFn) ([]Language, error)
	Tafsiraat(ctx context.Context, reqOpts ...ReqOptFn) ([]Tafsir, error)
	Chapters(ctx context.Context, reqOpts ...ReqOptFn) ([]Chapter, error)
	Chapter(ctx context.Context, id int, reqOpts ...ReqOptFn) (Chapter, error)
	ChapterInfo(ctx context.Context, id int, reqOpts ...ReqOptFn) (ChapterInfo, error)
	Verses(ctx context.Context, chapterID int, reqOpts ...VersesReqOptFn) ([]Verse, error)
	Verse(ctx context.Context, chapterID, verseID int) (Verse, error)
	Juzzah(ctx context.Context) ([]Juz, error)
	VerseTafsir(ctx context.Context, chapterID, verseID int, reqOpts ...VerseTafsirReqOptFn) ([]VerseTafsir, error)
	Search(ctx context.Context, query SearchRequest) (SearchResponse, error)
}

// Doer is an interface to abstract the http client out to its basic functionality.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type clientOpt struct {
	host string
	doer Doer
}

// ClientOptFn is an option to set the options of the client constructor.
type ClientOptFn func(opt clientOpt) clientOpt

// WithHost sets the host for the client url.
func WithHost(host string) ClientOptFn {
	return func(opt clientOpt) clientOpt {
		opt.host = host
		return opt
	}
}

// WithHTTPClient sets the http client on the quran api client.
func WithHTTPClient(doer Doer) ClientOptFn {
	return func(opt clientOpt) clientOpt {
		opt.doer = doer
		return opt
	}
}

// Client is the API client  that translates the quran.com api into familiar go types.
type Client struct {
	c *httpc.Client
}

// New Constructs a new Client. All default options will be  used if no options are
// provided to overwrite them. The defaults are:
//	host: https://quran.com/api
func New(opts ...ClientOptFn) *Client {
	opt := clientOpt{
		doer: &http.Client{Timeout: 15 * time.Second},
		host: "https://quran.com/api",
	}
	for _, o := range opts {
		opt = o(opt)
	}

	baseURL := opt.host + "/api/v3"
	return &Client{
		c: httpc.New(opt.doer, httpc.WithBaseURL(baseURL)),
	}
}

// Recitation is a recitation provided from quran.com.
type Recitation struct {
	ID                    int    `json:"id"`
	Style                 string `json:"style"`
	ReciterNameEng        string `json:"reciter_name_eng"`
	ReciterNameTranslated string `json:"reciter_name_translated"`
}

// Recitations returns all the available quran.com recitations.
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

// Translation is a translation available via the quran.com api. The translation's ID  maybe used in
// other api calls to add translations to the response.
type Translation struct {
	ID           int    `json:"id"`
	AuthorName   string `json:"author_name"`
	LanguageName string `json:"language_name"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
}

// Translations returns all the available quran.com translations.
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

// TranslatedName is a name and the language is is translated from.
type TranslatedName struct {
	LanguageName string `json:"language_name"`
	Name         string `json:"name"`
}

// Language is the a language with associated quran.com identifiers. The language ID is
// useful in filtering other api calls to the language provided. The iso code is useful
// in other contexts like search and verses calls.
type Language struct {
	ID              int              `json:"id"`
	Name            string           `json:"name"`
	IsoCode         string           `json:"iso_code"`
	NativeName      string           `json:"native_name"`
	Direction       string           `json:"direction"`
	TranslatedNames []TranslatedName `json:"translated_names"`
}

// Languages returns all the available quran.com languages.
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

// Tafsir is a tafsir overview available from quran.com. Food for thought, the slug
// is never populated but is "supported" through the docs, but not in reality.
type Tafsir struct {
	ID           int    `json:"id"`
	AuthorName   string `json:"author_name"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	LanguageName string `json:"language_name"`
}

// Tafsiraat returns all the available quran.com tafsiraat.
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

// Chapter or surah along with its relevant metadata combine to detail the summary of the
// chatper as a whole.
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

// Chapters returns the available chapters from quran.com.
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

// Chapters returns the the given chapter by id from quran.com.
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
		Language   string
		Recitation int
		TextType   string

		Page   int
		Limit  int
		Offset int

		Media        []int
		Translations []int
	}
)

func (v versesReqOpt) queryParams(r *httpc.Request) *httpc.Request {
	if v.Language != "" {
		r = r.QueryParam("language", v.Language)
	}

	if v.Recitation > 0 {
		r = r.QueryParam("recitation", strconv.Itoa(v.Recitation))
	}

	if v.TextType != "" {
		r = r.QueryParam("text_type", v.TextType)
	}

	if v.Page > 0 {
		r = r.QueryParam("page", strconv.Itoa(v.Page))
	}

	if v.Limit > 0 {
		r = r.QueryParam("limit", strconv.Itoa(v.Limit))
	}

	for _, media := range v.Media {
		r = r.QueryParam("media[]", strconv.Itoa(media))
	}

	for _, translation := range v.Translations {
		r = r.QueryParam("translations[]", strconv.Itoa(translation))
	}

	return r
}

func (v versesReqOpt) key(chapterID int) ([]byte, error) {
	sort.Ints(v.Media)
	sort.Ints(v.Translations)

	input := struct {
		VerseReqOpts versesReqOpt
		ChapterID    int
	}{
		VerseReqOpts: v,
		ChapterID:    chapterID,
	}

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(input)
	return buf.Bytes(), err
}

func VersesLanguage(isoCode string) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.Language = isoCode
		return opts
	}
}

func VersesRecitation(recitation int) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.Recitation = recitation
		return opts
	}
}

func VersesTextType(textType string) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.TextType = textType
		return opts
	}
}

func VersesLimit(i int) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.Limit = i
		return opts
	}
}

func VersesOffset(i int) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.Offset = i
		return opts
	}
}

func VersesPage(i int) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.Page = i
		return opts
	}
}

func VersesMedia(media []int) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.Media = media
		return opts
	}
}

func VersesTranslations(translations []int) VersesReqOptFn {
	return func(opts versesReqOpt) versesReqOpt {
		opts.Translations = translations
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
