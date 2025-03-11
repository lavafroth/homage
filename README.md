# Homage

Simple homepage for home lab battery stats and services.

## Deploy

Run the following commands to build the project:

```
tailwindcss -i input.css -o assets/style.css
go build .
```

Optionally, modify the server address to that of your home lab.

The above command builds a single binary with the project name. Drop it on the server and execute it.
