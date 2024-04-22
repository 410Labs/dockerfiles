FROM golang:1.22.2 as build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian11

EXPOSE 80/tcp
COPY --from=build /go/bin/app /
CMD ["/app"]
