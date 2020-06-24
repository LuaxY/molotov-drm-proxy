package molotov

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

const (
	xMolotovAgent = "{\"app_id\":\"electron_app\",\"app_build\":3,\"app_version_name\":\"4.2.1\",\"type\":\"desktop\",\"os_version\":\"macOs new version\",\"electron_version\":\"4.1.5\",\"os\":\"macOS\",\"manufacturer\":\"Apple\",\"serial\":\"7B819232-2DCB-5BD4-8D4F-A27CDB4F90FA\",\"model\":\"MacBook Pro\",\"hasTouchbar\":false,\"brand\":\"Apple\",\"api_version\":8,\"features_supported\":[\"social\",\"download_to_go\",\"new_button_conversion\",\"paywall\",\"channel_separator\",\"download_to_go_lot_2\",\"empty_view_v2\"],\"inner_app_version_name\":\"3.55.0\",\"qa\":false}"
	userAgent     = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Molotov/4.2.1 Chrome/69.0.3497.128 Electron/4.1.5 Safari/537.36"
)

type Client struct {
	ctx         context.Context
	user        string
	pass        string
	AccessToken string
}

func New(ctx context.Context, user, pass string) (*Client, error) {
	return &Client{
		ctx:  ctx,
		user: user,
		pass: pass,
	}, nil
}

func (c *Client) Auth() (bool, error) {
	if c.user == "" || c.pass == "" {
		return false, errors.New("no credentials provided")
	}

	payload := struct {
		GrantType string `json:"grant_type"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}{
		GrantType: "password",
		Email:     c.user,
		Password:  c.pass,
	}

	jPayload, err := json.Marshal(payload)

	if err != nil {
		return false, errors.Wrap(err, "json format credentials")
	}

	req, err := http.NewRequest("POST", "https://fapi.molotov.tv/v3.1/auth/login", bytes.NewBuffer(jPayload))

	if err != nil {
		return false, errors.Wrap(err, "create GET request")
	}

	req.Header.Set("X-Molotov-Agent", xMolotovAgent)
	req.Header.Set("User-Agent", userAgent)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return false, errors.Wrap(err, "post login")
	}

	defer func() {
		_ = res.Body.Close()
	}()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return false, errors.Wrap(err, "read body response")
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return false, errors.Errorf("bad http code: %s: %s", res.Status, body)
	}

	var data struct {
		Auth struct {
			AccessToken string `json:"access_token"`
		} `json:"auth"`
	}

	if err = json.Unmarshal(body, &data); err != nil {
		return false, errors.Wrap(err, "json parse response")
	}

	if data.Auth.AccessToken == "" {
		return false, errors.Errorf("no access token received: %s: %s", res.Status, body)
	}

	c.AccessToken = data.Auth.AccessToken

	return true, nil
}

type Asset struct {
	DRM struct {
		Token string `json:"token"`
	} `json:"drm"`
	Stream struct {
		URL string `json:"url"`
	} `json:"stream"`
}

func (c *Client) GetAsset(id int) (*Asset, error) {
	if c.AccessToken == "" {
		return nil, errors.New("no molotov access token available")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://fapi.molotov.tv/v2/me/assets?cwatch=true&id=%d&trkCp=season&trkCs=vod&trkOp=home&trkOs=on-tv-77&type=vod&access_token=%s", id, c.AccessToken), nil)

	if err != nil {
		return nil, errors.Wrap(err, "create GET asset request")
	}

	req.Header.Set("X-Molotov-Agent", xMolotovAgent)
	req.Header.Set("User-Agent", userAgent)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, errors.Wrap(err, "get asset")
	}

	defer func() {
		_ = res.Body.Close()
	}()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, errors.Wrap(err, "read body response")
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		// TODO re Auth()
		return nil, errors.Errorf("bad http code: %s: %s", res.Status, body)
	}

	var asset Asset

	if err = json.Unmarshal(body, &asset); err != nil {
		return nil, errors.Wrap(err, "json parse response")
	}

	if asset.DRM.Token == "" || asset.Stream.URL == "" {
		return nil, errors.Errorf("no drm token or stream url received: %s: %s", res.Status, body)
	}

	return &asset, nil
}
