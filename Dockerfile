FROM --platform=$BUILDPLATFORM golang:1.27.4 AS build
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -trimpath -o /out/google-search-mcp ./cmd/google-search-mcp

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build --chown=65532:65532 /out/google-search-mcp /app/google-search-mcp
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/app/google-search-mcp"]
