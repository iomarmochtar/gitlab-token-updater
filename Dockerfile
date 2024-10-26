FROM goreleaser/goreleaser:v2.3.2 AS builder
WORKDIR /build
COPY . .
RUN goreleaser build --clean --single-target

FROM alpine:3.20.3
RUN apk --no-cache add ca-certificates
COPY --from=builder /build/dist/gitlab-token-updater_linux_amd64_v1/gitlab-token-updater /usr/bin/gitlab-token-updater
ENTRYPOINT ["gitlab-token-updater"]
CMD ["--help"]