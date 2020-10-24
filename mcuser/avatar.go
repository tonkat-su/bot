package mcuser

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

var ErrAvatarServiceDown = errors.New("crafatar.com is down")

func GetFace(uuid string) ([]byte, error) {
	if uuid == "" {
		return nil, errors.New("uuid is required")
	}

	query := url.Values{}
	query.Set("size", "128")
	query.Set("overlay", "true")
	query.Set("scale", "10")

	u := &url.URL{
		Scheme:   "https",
		Host:     "crafatar.com",
		Path:     path.Join("avatar", uuid),
		RawQuery: query.Encode(),
	}

	response, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, ErrAvatarServiceDown
	}
	return ioutil.ReadAll(response.Body)
}
