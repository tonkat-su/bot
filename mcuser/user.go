package mcuser

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"
)

type playerDBResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Player struct {
			Meta struct {
				NameHistory []struct {
					Name string `json:"name"`
				} `json:"name_history"`
			} `json:"meta"`
			Username string `json:"username"`
			ID       string `json:"id"`
			RawID    string `json:"raw_id"`
			Avatar   string `json:"avatar"`
		} `json:"player"`
	} `json:"data"`
	Success bool `json:"success"`
}

func GetUuid(name string) (string, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   "playerdb.co",
		Path:   path.Join("api", "player", "minecraft", name),
	}
	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data playerDBResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	if data.Code != "player.found" || data.Data.Player.RawID == "" {
		return "", nil
	}

	return data.Data.Player.RawID, nil
}
