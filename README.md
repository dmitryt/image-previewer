# image-previewer 🚀
![Build Status](https://github.com/dmitryt/image-previewer/workflows/Lint,%20Test%20and%20Deploy/badge.svg)
![PR Status](https://github.com/dmitryt/image-previewer/workflows/Lint%20and%20Test/badge.svg)
![Go Report Card](https://goreportcard.com/badge/github.com/dmitryt/image-previewer)

Web service for resizing images. Supported extensions: jpg, png, gif.

## Usage

1. As a separate service

```console
docker run -p 8083:8082  greml1n/image-previewer
```
2. Run locally

```console
make run
```

## Example

Launch the service and open the following URL in browser

```console
http://localhost:8082/fill/300/200/www.audubon.org/sites/default/files/a1_1902_16_barred-owl_sandra_rothenberg_kk.jpg
```

## Environment Variables

```console
# defaults to "8082"
PORT=3000

# defaults to "5242880" - 5mb
MAX_FILE_SIZE=10485760

# defaults to ".cache"
CACHE_DIR=/path/to-dir

# defaults to "10"
CACHE_SIZE=50
```
