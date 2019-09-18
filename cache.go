package quranc

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"

	"go.etcd.io/bbolt"
)

type boltCacheMiddleware struct {
	db   *bbolt.DB
	next QuranAPI
}

const (
	bucketChapters     = "chapters"
	bucketChapter      = "chapter"
	bucketChapterInfo  = "chapterinfo"
	bucketJuzzah       = "juzzah"
	bucketLanguages    = "languages"
	bucketRecitations  = "recitations"
	bucketTafsiraat    = "tafsiraat"
	bucketTranslations = "translations"
	bucketVerse        = "verse"
	bucketVerseTafsir  = "verse_tafsir"
	bucketVerses       = "verses"
)

func BoltCache(client QuranAPI, db *bbolt.DB) (QuranAPI, error) {
	buckets := map[string][]string{
		bucketChapters:     {bucketChapter, bucketChapterInfo},
		bucketJuzzah:       nil,
		bucketLanguages:    nil,
		bucketRecitations:  nil,
		bucketTafsiraat:    nil,
		bucketTranslations: nil,
		bucketVerses:       {bucketVerse, bucketVerseTafsir},
	}
	for bucket, nestedBuckets := range buckets {
		err := db.Update(func(tx *bbolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return fmt.Errorf("create bucket %q: %s", bucket, err)
			}

			for _, nestedBucket := range nestedBuckets {
				_, err := b.CreateBucketIfNotExists([]byte(nestedBucket))
				if err != nil {
					return fmt.Errorf("create nested bucket %q: %s", nestedBucket, err)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return &boltCacheMiddleware{
		db:   db,
		next: client,
	}, nil
}

func (bc *boltCacheMiddleware) Recitations(ctx context.Context, reqOpts ...ReqOptFn) ([]Recitation, error) {
	var opt reqOpt
	for _, o := range reqOpts {
		opt = o(opt)
	}

	bucket := []byte(bucketRecitations)
	cacheID := []byte(itoa(opt.languageID))

	var out []Recitation
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		return valueDecode(b.Get(cacheID), &out)
	})
	if err == nil {
		return out, nil
	}

	clientOut, err := bc.next.Recitations(ctx, reqOpts...)
	if err != nil {
		return nil, err
	}

	// safely ignore error here, if we have an error we swallow it since it is not in the critical path.
	bc.db.Update(func(tx *bbolt.Tx) error {
		buf, err := valueEncoder(clientOut)
		if err != nil {
			return err
		}
		b := tx.Bucket(bucket)
		return b.Put(cacheID, buf.Bytes())
	})

	return clientOut, nil
}

func (bc *boltCacheMiddleware) Translations(ctx context.Context, reqOpts ...ReqOptFn) ([]Translation, error) {
	var opt reqOpt
	for _, o := range reqOpts {
		opt = o(opt)
	}

	bucket := []byte(bucketTranslations)
	cacheID := []byte(itoa(opt.languageID))

	var out []Translation
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		return valueDecode(b.Get(cacheID), &out)
	})
	if err == nil {
		return out, nil
	}

	clientOut, err := bc.next.Translations(ctx, reqOpts...)
	if err != nil {
		return nil, err
	}

	// safely ignore error here, if we have an error we swallow it since it is not in the critical path.
	bc.db.Update(func(tx *bbolt.Tx) error {
		buf, err := valueEncoder(clientOut)
		if err != nil {
			return err
		}
		b := tx.Bucket(bucket)
		return b.Put(cacheID, buf.Bytes())
	})

	return clientOut, nil
}

func (bc *boltCacheMiddleware) Languages(ctx context.Context, reqOpts ...ReqOptFn) ([]Language, error) {
	var opt reqOpt
	for _, o := range reqOpts {
		opt = o(opt)
	}

	bucket := []byte(bucketLanguages)
	cacheID := []byte(itoa(opt.languageID))

	var out []Language
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		return valueDecode(b.Get(cacheID), &out)
	})
	if err == nil {
		return out, nil
	}

	clientOut, err := bc.next.Languages(ctx, reqOpts...)
	if err != nil {
		return nil, err
	}

	// safely ignore error here, if we have an error we swallow it since it is not in the critical path.
	bc.db.Update(func(tx *bbolt.Tx) error {
		buf, err := valueEncoder(clientOut)
		if err != nil {
			return err
		}
		b := tx.Bucket(bucket)
		return b.Put(cacheID, buf.Bytes())
	})

	return clientOut, nil
}

func (bc *boltCacheMiddleware) Tafsiraat(ctx context.Context, reqOpts ...ReqOptFn) ([]Tafsir, error) {
	var opt reqOpt
	for _, o := range reqOpts {
		opt = o(opt)
	}

	bucket := []byte(bucketTafsiraat)
	cacheID := []byte(itoa(opt.languageID))

	var out []Tafsir
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		return valueDecode(b.Get(cacheID), &out)
	})
	if err == nil {
		return out, nil
	}

	clientOut, err := bc.next.Tafsiraat(ctx, reqOpts...)
	if err != nil {
		return nil, err
	}

	// safely ignore error here, if we have an error we swallow it since it is not in the critical path.
	bc.db.Update(func(tx *bbolt.Tx) error {
		buf, err := valueEncoder(clientOut)
		if err != nil {
			return err
		}

		b := tx.Bucket(bucket)
		return b.Put(cacheID, buf.Bytes())
	})

	return clientOut, nil
}

func (bc *boltCacheMiddleware) Chapters(ctx context.Context, reqOpts ...ReqOptFn) ([]Chapter, error) {
	var opt reqOpt
	for _, o := range reqOpts {
		opt = o(opt)
	}

	bucket := []byte(bucketChapters)
	cacheID := []byte(itoa(opt.languageID))

	var out []Chapter
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		return valueDecode(b.Get(cacheID), &out)
	})
	if err == nil {
		return out, nil
	}

	clientOut, err := bc.next.Chapters(ctx, reqOpts...)
	if err != nil {
		return nil, err
	}

	// safely ignore error here, if we have an error we swallow it since it is not in the critical path.
	bc.db.Update(func(tx *bbolt.Tx) error {
		buf, err := valueEncoder(clientOut)
		if err != nil {
			return err
		}
		b := tx.Bucket(bucket)
		return b.Put(cacheID, buf.Bytes())
	})

	return clientOut, nil
}

func (bc *boltCacheMiddleware) Chapter(ctx context.Context, id int, reqOpts ...ReqOptFn) (Chapter, error) {
	var opt reqOpt
	for _, o := range reqOpts {
		opt = o(opt)
	}

	bucket := []byte(bucketChapters)
	nestedBucket := []byte(bucketChapter)
	cacheID := []byte(join(itoa(opt.languageID), itoa(id)))

	var out Chapter
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket).Bucket(nestedBucket)
		return valueDecode(b.Get(cacheID), &out)
	})
	if err == nil {
		return out, nil
	}

	clientOut, err := bc.next.Chapter(ctx, id, reqOpts...)
	if err != nil {
		return Chapter{}, err
	}

	// safely ignore error here, if we have an error we swallow it since it is not in the critical path.
	bc.db.Update(func(tx *bbolt.Tx) error {
		buf, err := valueEncoder(clientOut)
		if err != nil {
			return err
		}
		b := tx.Bucket(bucket).Bucket(nestedBucket)
		return b.Put(cacheID, buf.Bytes())
	})

	return clientOut, nil
}

func (bc *boltCacheMiddleware) ChapterInfo(ctx context.Context, id int, reqOpts ...ReqOptFn) (ChapterInfo, error) {
	var opt reqOpt
	for _, o := range reqOpts {
		opt = o(opt)
	}

	bucket := []byte(bucketChapters)
	nestedBucket := []byte(bucketChapterInfo)
	cacheID := []byte(join(itoa(opt.languageID), itoa(id)))

	var out ChapterInfo
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket).Bucket(nestedBucket)
		return valueDecode(b.Get(cacheID), &out)
	})
	if err == nil {
		return out, nil
	}

	clientOut, err := bc.next.ChapterInfo(ctx, id, reqOpts...)
	if err != nil {
		return ChapterInfo{}, err
	}

	// safely ignore error here, if we have an error we swallow it since it is not in the critical path.
	bc.db.Update(func(tx *bbolt.Tx) error {
		buf, err := valueEncoder(clientOut)
		if err != nil {
			return err
		}
		b := tx.Bucket(bucket).Bucket(nestedBucket)
		return b.Put(cacheID, buf.Bytes())
	})

	return clientOut, nil
}

func (bc *boltCacheMiddleware) Verses(ctx context.Context, chapterID int, reqOpts ...VersesReqOptFn) ([]Verse, error) {
	var opt versesReqOpt
	for _, o := range reqOpts {
		opt = o(opt)
	}

	bucket := []byte(bucketVerses)
	cacheID, err := opt.key(chapterID)
	if err != nil {
		return bc.next.Verses(ctx, chapterID, reqOpts...)
	}

	var out []Verse
	err = bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		return valueDecode(b.Get(cacheID), &out)
	})
	if err == nil {
		return out, nil
	}

	clientOut, err := bc.next.Verses(ctx, chapterID, reqOpts...)
	if err != nil {
		return nil, err
	}

	// safely ignore error here, if we have an error we swallow it since it is not in the critical path.
	bc.db.Update(func(tx *bbolt.Tx) error {
		buf, err := valueEncoder(clientOut)
		if err != nil {
			return err
		}
		b := tx.Bucket(bucket)
		return b.Put(cacheID, buf.Bytes())
	})

	return clientOut, nil
}

func (bc *boltCacheMiddleware) Verse(ctx context.Context, chapterID, verseID int) (Verse, error) {
	bucket := []byte(bucketVerses)
	nestedBucket := []byte(bucketVerse)
	cacheID := []byte(join(itoa(chapterID), itoa(verseID)))

	var out Verse
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket).Bucket(nestedBucket)
		return valueDecode(b.Get(cacheID), &out)
	})
	if err == nil {
		return out, nil
	}

	clientOut, err := bc.next.Verse(ctx, chapterID, verseID)
	if err != nil {
		return Verse{}, err
	}

	// safely ignore error here, if we have an error we swallow it since it is not in the critical path.
	bc.db.Update(func(tx *bbolt.Tx) error {
		buf, err := valueEncoder(clientOut)
		if err != nil {
			return err
		}
		b := tx.Bucket(bucket).Bucket(nestedBucket)
		return b.Put(cacheID, buf.Bytes())
	})

	return clientOut, nil
}

func (bc *boltCacheMiddleware) Juzzah(ctx context.Context) ([]Juz, error) {
	bucket := []byte(bucketJuzzah)
	cacheID := []byte("juzzah")

	var out []Juz
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		return valueDecode(b.Get(cacheID), &out)
	})
	if err == nil {
		return out, nil
	}

	clientOut, err := bc.next.Juzzah(ctx)
	if err != nil {
		return nil, err
	}

	// safely ignore error here, if we have an error we swallow it since it is not in the critical path.
	bc.db.Update(func(tx *bbolt.Tx) error {
		buf, err := valueEncoder(clientOut)
		if err != nil {
			return err
		}
		b := tx.Bucket(bucket)
		return b.Put(cacheID, buf.Bytes())
	})

	return clientOut, nil
}

func (bc *boltCacheMiddleware) VerseTafsir(ctx context.Context, chapterID, verseID int, reqOpts ...VerseTafsirReqOptFn) ([]VerseTafsir, error) {
	var opt verseTafsirReqOpts
	for _, o := range reqOpts {
		opt = o(opt)
	}

	bucket := []byte(bucketVerses)
	nestedBucket := []byte(bucketVerseTafsir)
	cacheID := []byte(join(opt.Tafsir, itoa(chapterID), itoa(verseID)))

	var out []VerseTafsir
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket).Bucket(nestedBucket)
		return valueDecode(b.Get(cacheID), &out)
	})
	if err == nil {
		return out, nil
	}

	clientOut, err := bc.next.VerseTafsir(ctx, chapterID, verseID, reqOpts...)
	if err != nil {
		return nil, err
	}

	// safely ignore error here, if we have an error we swallow it since it is not in the critical path.
	bc.db.Update(func(tx *bbolt.Tx) error {
		buf, err := valueEncoder(clientOut)
		if err != nil {
			return err
		}
		b := tx.Bucket(bucket).Bucket(nestedBucket)
		return b.Put(cacheID, buf.Bytes())
	})

	return clientOut, nil
}

func (bc *boltCacheMiddleware) Search(ctx context.Context, query SearchRequest) (SearchResponse, error) {
	return bc.next.Search(ctx, query)
}

func valueDecode(b []byte, v interface{}) error {
	buf := bytes.NewBuffer(b)

	if err := gob.NewDecoder(buf).Decode(v); err != nil {
		return err
	}

	return nil
}

func valueEncoder(v interface{}) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return nil, err
	}
	return &buf, nil
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

func join(ss ...string) string {
	return strings.Join(ss, ":")
}
