FROM golang:onbuild as builder

WORKDIR /app

ADD . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o goapp .

#---

FROM alpine

WORKDIR /app

COPY --from=builder /app/goapp .

EXPOSE 8000

CMD ["./goapp"]
