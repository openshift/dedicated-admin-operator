FROM openshift/origin-release:golang-1.10 AS build
ENV OPERATOR_PATH=/go/src/github.com/openshift/dedicated-admin-operator
COPY . ${OPERATOR_PATH}
WORKDIR ${OPERATOR_PATH}
RUN yum update -y --exclude=golang && yum clean all && make build

FROM alpine:3.8
ENV OPERATOR_PATH=/go/src/github.com/openshift/dedicated-admin-operator \
    OPERATOR_BIN=dedicated-admin-operator \
    USER_UID=1001 \
    USER_NAME=dedicated-admin-operator

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build ${OPERATOR_PATH}/build/_output/bin/${OPERATOR_BIN} /usr/local/bin/${OPERATOR_BIN}
USER ${USER_UID}
LABEL io.openshift.managed.name="dedicated-admin-operator" \
      io.openshift.managed.description="Operator to manage permissions for Dedicated admins"
      