# Multi-stage production image for the Sumeru engine (single-node pilot).
FROM golang:1.26.6-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/sumeru ./cmd/sumeru

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/sumeru /app/sumeru
COPY --from=build /src/addons /app/addons
COPY --from=build /src/core/engine/assets /app/core/engine/assets
COPY --from=build /src/core/engine/templates /app/core/engine/templates
COPY --from=build /src/sumeru.conf.example /app/sumeru.conf
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/sumeru", "-c", "/app/sumeru.conf"]
