# corsanywhere

Zero dependency go version of https://github.com/Rob--W/cors-anywhere.

`corsanywhere` adds CORS headers to proxied requests on the fly, allowing you to bypass [existing CORS configuration](https://developer.mozilla.org/en-US/docs/Web/Security/Practical_implementation_guides/CORS).

Distributed in GitHub container registry for convenience.

## Usage

The binary requires the `Origin` header to be set and removes `Set-Cookie` headers.

The destination url is passed after `/` in the path. The protocol (`http`|`https`) is optional and defaults to `https`.

```shell
curl -H Origin:https://example.com localhost:8080/https://example.com
```

## Configuration

Basic configuration is available through environment variables.

- `PORT` (default `8080`)

## Contributing

Feel free to open a PR.
