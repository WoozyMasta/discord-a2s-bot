FROM docker.io/golang:1.23 AS build

ARG ARCH=amd64
ARG VERSION=0.0.0
ARG COMMIT=nil
ARG UPX_VERSION=4.2.4

# hadolint ignore=DL3008
RUN set -xeu; \
    apt-get update; \
    apt-get install -y --no-install-recommends xz-utils curl; \
    curl -#Lo upx.tar.xz \
        "https://github.com/upx/upx/releases/download/v$UPX_VERSION/upx-$UPX_VERSION-${ARCH}_linux.tar.xz"; \
    tar -xvf upx.tar.xz --strip-components=1 "upx-$UPX_VERSION-${ARCH}_linux/upx"; \
    chmod +x upx; \
    mv upx /usr/local/bin/upx; \
    rm -f upx.tar.xz

WORKDIR /src/discord-a2s-bot
COPY go.mod go.sum ./
RUN go mod download

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOFLAGS="-buildvcs=false -trimpath"
ENV GOARCH=$ARCH

COPY ./cli ./cli
COPY ./internal ./internal

RUN set -eux;\
    go mod tidy; \
    go generate ./...; \
    go build -ldflags="-s -w \
        -X 'internal/vars.Version=$VERSION' \
        -X 'internal/vars.Commit=$COMMIT' \
        -X 'internal/vars.BuildTime=$(date -uIs)' \
        -X 'internal/vars.URL=https://$(grep -Po 'module \K.*$' go.mod)'" \
      -o "./discord-a2s-bot" "cli/"*.go; \
    upx --lzma --best ./discord-a2s-bot; \
    upx -t ./discord-a2s-bot

FROM scratch
COPY --from=build /src/discord-a2s-bot/discord-a2s-bot /discord-a2s-bot
ENTRYPOINT ["/discord-a2s-bot"]
