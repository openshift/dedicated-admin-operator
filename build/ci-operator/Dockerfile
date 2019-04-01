FROM registry.svc.ci.openshift.org/openshift/release:golang-1.10 AS builder
ENV OPERATOR_PATH=/go/src/github.com/openshift/dedicated-admin-operator
COPY . ${OPERATOR_PATH}
WORKDIR ${OPERATOR_PATH}
RUN yum update -y --exclude=golang && yum clean all && make build

FROM registry.svc.ci.openshift.org/openshift/origin-v4.0:base
ENV OPERATOR_PATH=/go/src/github.com/openshift/dedicated-admin-operator \
    OPERATOR_BIN=dedicated-admin-operator

RUN yum install ca-certificates && yum clean all
WORKDIR /root/
COPY --from=builder ${OPERATOR_PATH}/build/_output/bin/${OPERATOR_BIN} /usr/local/bin/${OPERATOR_BIN}
LABEL io.openshift.managed.name="dedicated-admin-operator" \
      io.openshift.managed.description="Operator to manage permissions for Dedicated admins"
      