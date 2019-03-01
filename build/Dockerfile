FROM alpine:3.8

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/bin/dedicated-admin-operator /usr/local/bin/dedicated-admin-operator
