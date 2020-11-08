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

func queryPlayerDb(identifier string) (response *playerDBResponse, err error) {
	u := &url.URL{
		Scheme: "https",
		Host:   "playerdb.co",
		Path:   path.Join("api", "player", "minecraft", identifier),
	}
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data playerDBResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	if data.Code != "player.found" || data.Data.Player.RawID == "" {
		return nil, nil
	}
	return &data, nil
}

func GetUuid(name string) (string, error) {
	data, err := queryPlayerDb(name)
	if err != nil {
		return "", err
	}
	return data.Data.Player.ID, nil
}

func GetUsername(id string) (string, error) {
	data, err := queryPlayerDb(id)
	if err != nil {
		return "", err
	}
	return data.Data.Player.Username, nil
}
