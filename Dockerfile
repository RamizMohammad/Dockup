# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
ARG VERSION=dev
ARG TARGETOS=linux
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o /out/dockup ./cmd/dockup

# docker:cli ships the docker CLI + compose plugin, which is what deployments run through.
FROM docker:cli
RUN apk add --no-cache git openssh-client ca-certificates tzdata
COPY --from=build /out/dockup /usr/local/bin/dockup
ENV DOCKUP_DATA=/var/lib/dockup \
    DOCKUP_ADDR=:8100
VOLUME ["/var/lib/dockup"]
EXPOSE 8100
HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD wget -qO- http://127.0.0.1:8100/healthz >/dev/null || exit 1
ENTRYPOINT ["dockup"]
CMD ["serve"]
