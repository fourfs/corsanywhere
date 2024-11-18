ARG GO_VERSION
ARG VERSION

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /
COPY go.mod main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-X main.version=${VERSION}" -o corsanywhere

FROM gcr.io/distroless/static

COPY --from=build /corsanywhere /
ENTRYPOINT [ "/corsanywhere" ]
