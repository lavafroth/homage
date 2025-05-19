# Homage

Simple homepage for home lab battery stats and services.

## Deploy

(Optional) download orbitron font

```sh
curl -L https://github.com/theleagueof/orbitron/raw/refs/heads/master/webfonts/orbitron-light-webfont.ttf -o assets/font.ttf
```

Build the project

```sh
tailwindcss -i input.css -o assets/style.css
go build .
```

Optionally, modify the `config.txt` to fit your home lab.

The above command builds a single binary. Drop it on the server and execute it.
