# syntax=docker/dockerfile:1
# check=error=true

FROM docker.io/library/node:lts-slim AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM docker.io/library/golang:1.27 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
COPY web/embed.go web/
COPY --from=web /src/web/dist web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ssdvb ./cmd/ssdvb

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/ssdvb /ssdvb
EXPOSE 3000
ENTRYPOINT ["/ssdvb"]
CMD ["serve"]
