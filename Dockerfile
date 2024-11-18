ARG go_version

FROM golang:${go_version}-alpine AS build
COPY go.mod main.go ./

ARG tag
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-X main.version=${tag}" -o /corsanywhere

FROM gcr.io/distroless/static

COPY --from=build /corsanywhere /
ENTRYPOINT [ "/corsanywhere" ]
