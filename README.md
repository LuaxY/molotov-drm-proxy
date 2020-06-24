# Molotov.tv DRM Proxy

This server offer ability to proxify [Molotov.tv](https://www.molotov.tv/) and its DRM ([DRM Today](https://castlabs.com/drmtoday/)).  
Valid account is required.

## :warning: Disclaimer

This project aims to demonstrate the feasibility (PoC), sharing copyrighted content without the necessary authorization is prohibited, use at your own risk.

## Build & Run

```shell script
docker build -t motolov-proxy .
docker run --name motolov-proxy -p 80:80 -e PORT=80 -e USER={MOLOTOV_USER} -e PASS={MOLOTOV_PASS} motolov-proxy 
```

:warning: Some geo-restriction are in place, the server IP must be in correct country to watch content.

## Profit

You need episode ID in order to request the correct video, by example Rick & Morty S04E01 is `3997154`. 

Manifest: http://localhost/cdn/3997154.mpd  
DRM: http://localhost/drm/3997154

You can test here: https://www.jwplayer.com/developers/stream-tester/  
(Use `Widevine` as DRM)

### Infos

1. I have removed subtitles from manifest in order to be compatible with JWPlayer.
2. Your content MUST be distribued with HTTPS, you can use Let's Encrypt or CloudFlare for this. 