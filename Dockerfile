FROM alpine:3.7

# RUN apk add --no-cache ca-certificates

COPY main /

CMD ["/main"]