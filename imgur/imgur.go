package imgur

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"hash/crc32"
	"mime/multipart"
	"net/http"
	"sync"
)

// ImageUploadRequest is an image to upload with its metadata
type ImageUploadRequest struct {
	Image []byte
	Name  string
}

type Image struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Datetime    int64         `json:"datetime"`
	Type        string        `json:"type"`
	Animated    bool          `json:"animated"`
	Width       int64         `json:"width"`
	Height      int64         `json:"height"`
	Size        int64         `json:"size"`
	Views       int64         `json:"views"`
	Bandwidth   int64         `json:"bandwidth"`
	Vote        int64         `json:"vote"`
	Favorite    bool          `json:"favorite"`
	Nsfw        bool          `json:"nsfw"`
	Section     string        `json:"section"`
	AccountURL  string        `json:"account_url"`
	AccountID   int64         `json:"account_id"`
	IsAd        bool          `json:"is_ad"`
	InMostViral bool          `json:"in_most_viral"`
	Tags        []interface{} `json:"tags"`
	AdType      int64         `json:"ad_type"`
	AdURL       string        `json:"ad_url"`
	InGallery   bool          `json:"in_gallery"`
	Deletehash  string        `json:"deletehash"`
	Name        string        `json:"name"`
	Link        string        `json:"link"`
}

func (c *Client) upload(ctx context.Context, image *ImageUploadRequest) (*Image, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("image", image.Name)
	if err != nil {
		return nil, err
	}

	_, err = part.Write(image.Image)
	if err != nil {
		return nil, err
	}

	err = writer.WriteField("name", image.Name)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequestWithContext(ctx, "POST", "https://api.imgur.com/3/upload", &buf)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Authorization", "Client-ID "+c.ClientId)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type data struct {
		Data    Image `json:"data"`
		Success bool
		Status  int64
	}

	var d data
	err = json.NewDecoder(resp.Body).Decode(&d)
	if err != nil {
		return nil, err
	}

	if !d.Success {
		return nil, errors.New("failed to upload")
	}

	return &d.Data, nil
}

func (c *Client) Upload(ctx context.Context, req *ImageUploadRequest) (*Image, error) {
	c.init.Do(func() {
		c.imgCache = &imgCache{
			m: make(map[uint32]*Image),
		}
	})

	return c.imgCache.getImageUrl(ctx, req.Image, func() (*Image, error) {
		return c.upload(ctx, req)
	})
}

type imgCache struct {
	m map[uint32]*Image
}

func (cache *imgCache) getImageUrl(ctx context.Context, data []byte, upload func() (*Image, error)) (*Image, error) {
	key := crc32.Checksum(data, crc32.MakeTable(crc32.Castagnoli))

	var (
		image *Image
		ok    bool
		err   error
	)
	image, ok = cache.m[key]
	if !ok {
		image, err = upload()
		if err != nil {
			return nil, err
		}
		cache.m[key] = image
	}
	return image, nil
}

type Client struct {
	init     sync.Once
	imgCache *imgCache
	ClientId string
}
