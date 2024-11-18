# corsanywhere

Zero dependency go version of https://github.com/Rob--W/cors-anywhere, with some extra utility.

`corsanywhere` adds CORS headers to proxied requests on the fly, allowing you to bypass [existing CORS configuration](https://developer.mozilla.org/en-US/docs/Web/Security/Practical_implementation_guides/CORS). See [usage](#usage) below.

Distributed in GitHub container registry [as a docker image](https://github.com/fourfs/corsanywhere/pkgs/container/corsanywhere) for convenience.

## Usage

`corsanywhere` removes `Set-Cookie` headers by default (See [configuration](#configuration)).
`X-Forwarded-*` headers are not sent to host.

Destination url is passed after `/` in the path. Protocol (`http`|`https`) is optional and defaults to `https`.

A special `X-Set-Origin` header can be included to spoof `Origin` and `Referer` headers.
The header value should follow the example of `http(s)://domain.tld` with no trailing slashes.
An extra trailing slash is automatically added for the `Referer` header.

### Run it with docker:

```shell
docker run -p 8080:8080 ghcr.io/fourfs/corsanywhere:latest
```

### Run it locally:

```shell
go run main.go
```

### Request from anywhere:

```shell
curl localhost:8080/https://example.com
```

## Configuration

Basic configuration is available through environment variables.

- `PORT` (default `8080`)
- `REQUIRE_HEADERS` (default `""`) - Comma-separated list of headers to require. Case-insensitive.
- `REMOVE_HEADERS` (default `"Set-Cookie,Set-Cookie2"`) - Comma-separated list of headers to remove. Case-insensitive.

## Contributing

Feel free to open a PR.
