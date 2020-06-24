package widevine

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

func TodayDRM(token string, requestBody io.Reader) ([]byte, error) {
	if token == "" {
		return nil, errors.New("no drm token available")
	}

	req, err := http.NewRequest("POST", "https://lic.drmtoday.com/license-proxy-widevine/cenc/", requestBody)

	if err != nil {
		return nil, errors.Wrap(err, "create POST request")
	}

	req.Header.Set("x-dt-auth-token", token)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, errors.Wrap(err, "post drm payload")
	}

	defer func() {
		_ = res.Body.Close()
	}()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, errors.Wrap(err, "read body response")
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, errors.Errorf("bad http code: %s: %s", res.Status, body)
	}

	var data struct {
		License string `json:"license"`
	}

	if err = json.Unmarshal(body, &data); err != nil {
		return nil, errors.Wrap(err, "json parse response")
	}

	if data.License == "" {
		return nil, errors.Errorf("no license payload received: %s: %s", res.Status, body)
	}

	payload, err := base64.StdEncoding.DecodeString(data.License)

	if err != nil {
		return nil, errors.Wrap(err, "base64 decode payload")
	}

	return payload, nil
}
