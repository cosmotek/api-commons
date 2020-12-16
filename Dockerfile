FROM alpine:latest as certs
RUN apk --update add ca-certificates
RUN apk add --no-cache tzdata

FROM scratch

# update local certs
COPY --from=certs /etc/ssl/ /etc/ssl

# copy tz data
COPY --from=certs /usr/share/zoneinfo /usr/share/zoneinfo

COPY bin/api /api
COPY migrations/ /migrations

# expose version number
ARG VERSION=noversion
ENV VERSION=$VERSION

# gRPC
EXPOSE 5000

ENTRYPOINT ["/api"]

