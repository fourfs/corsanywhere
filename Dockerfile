ARG go_version
ARG tag

FROM golang:${go_version}-alpine AS build

WORKDIR /
COPY go.mod main.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-X main.version=${tag}" -o corsanywhere

FROM gcr.io/distroless/static

COPY --from=build /corsanywhere /
ENTRYPOINT [ "/corsanywhere" ]
