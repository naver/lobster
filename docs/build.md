## Building from source

This project has separate source codes for `lobster` components.
You can build with a `Makefile` file in each directory.

### Before building

- The source code is based on `golang` and the build image is based on `docker`
- The API documentation uses `swagger` and `widdershins` is utilized for the GitHub
  - https://github.com/swaggo/swag
    - `make swag`
  - https://github.com/Mermade/widdershins?tab=readme-ov-file
    - `make widdershins`

### Lobster

- Reference `lobster/build/*/Dockerfile` from `lobster/Makefile`

```bash
make image-store
make image-query
make image-global-query
make image-syncer
make image-operator
make image-loggen
```

- If you want to build it yourself, please see below.

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o lobster ./lobster/cmd/lobster-store/main.go 
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o lobster ./lobster/cmd/lobster-query/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o lobster ./lobster/cmd/lobster-global-query/main.go 
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o lobster ./lobster/cmd/lobster-syncer/main.go 
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o lobster ./lobster/cmd/lobster-operator/main.go 
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o lobster ./lobster/cmd/loggen/main.go 
```

