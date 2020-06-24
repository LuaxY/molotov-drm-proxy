package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zencoder/go-dash/mpd"

	"widevine-proxy/internal/molotov"
	"widevine-proxy/internal/widevine"
)

var (
	molotovClient *molotov.Client
)

func main() {
	var err error

	port, _ := os.LookupEnv("PORT")
	user, _ := os.LookupEnv("USER")
	pass, _ := os.LookupEnv("PASS")

	ctx := context.Background()

	molotovClient, err = molotov.New(ctx, user, pass)

	if err != nil {
		log.Fatal(err)
	}

	connected, err := molotovClient.Auth()

	if err != nil || !connected {
		log.Fatal(err, connected)
	}

	log.Println("molotov: logged with user/pass")

	// Renew auth token each hours
	go func() {
		ticker := time.NewTicker(1 * time.Hour)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				connected, err := molotovClient.Auth()

				if err != nil || !connected {
					log.Fatal(err, connected)
				}

				log.Println("token refreshed", molotovClient.AccessToken)
			}
		}
	}()

	router := mux.NewRouter()

	router.HandleFunc("/", home)
	router.HandleFunc("/drm/{id:[0-9]+}", proxy)
	router.HandleFunc("/cdn/{id:[0-9]+}.mpd", cdn)

	http.Handle("/", router)

	log.Printf("listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func home(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("DRM Proxy"))
}

func proxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var err error
	var asset *molotov.Asset

	asset, err = molotovClient.GetAsset(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		_ = r.Body.Close()
	}()

	payload, err := widevine.TodayDRM(asset.DRM.Token, r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(payload)
}

func cdn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var err error
	var asset *molotov.Asset

	asset, err = molotovClient.GetAsset(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := http.Get(asset.Stream.URL)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		_ = res.Body.Close()
	}()

	manifest, err := mpd.Read(res.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add CDN base URL
	// FIXME we need url without last part, I use path.Dir for this, but this 'break' url, so I fix url with strings.Replace
	manifest.BaseURL = path.Dir(asset.Stream.URL) + "/"
	manifest.BaseURL = strings.Replace(manifest.BaseURL, "https:/", "https://", 1)

	for _, period := range manifest.Periods {
		for id, adaptionSet := range period.AdaptationSets {
			if *adaptionSet.ContentType == "text" {
				// Remove subtitles AdaptationSet (not working /w JWPlayer)
				// But subtitles are not encrypted, so they can me downloaded easily
				period.AdaptationSets[id] = nil
			}
		}
	}

	_ = manifest.Write(w)
}
