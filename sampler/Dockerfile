FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/centinela-sampler ./cmd/sampler

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/centinela-sampler /centinela-sampler
USER 65532:65532
EXPOSE 9500
ENTRYPOINT ["/centinela-sampler"]