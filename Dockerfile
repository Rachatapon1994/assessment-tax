FROM golang:1.22.2-alpine as build-base

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go test -v

RUN go build -o ./out/go-assessment-tax .

# ====================


FROM alpine:3.16.2
COPY --from=build-base /app/out/go-assessment-tax /app/go-assessment-tax

CMD ["/app/go-assessment-tax"]
