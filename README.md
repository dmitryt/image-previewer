# image-previewer 🚀
![Build Status](https://github.com/dmitryt/image-previewer/workflows/Lint,%20Test%20and%20Deploy/badge.svg)
![PR Status](https://github.com/dmitryt/image-previewer/workflows/Lint%20and%20Test/badge.svg)
![Go Report Card](https://goreportcard.com/badge/github.com/dmitryt/image-previewer)

Web service for resizing images. Supported extensions: jpg, png, gif.

## Example

TODO

## Environment Variables

```console
# defaults to "8082"
PORT=3000

# defaults to "5242880" - 5mb
MAX_FILE_SIZE=10485760

# defaults to "info"
LOG_LEVEL=debug

# defaults to ".cache"
CACHE_DIR=/path/to-dir

# defaults to "10"
CACHE_SIZE=50
```
