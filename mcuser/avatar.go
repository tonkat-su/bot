package mcuser

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
)

var ErrAvatarServiceDown = errors.New("minotar.net is down")

func GetFace(name string) ([]byte, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}

	u := &url.URL{
		Scheme: "https",
		Host:   "minotar.net",
		Path:   path.Join("help", name, "128.png"),
	}

	response, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, ErrAvatarServiceDown
	}
	return io.ReadAll(response.Body)
}
