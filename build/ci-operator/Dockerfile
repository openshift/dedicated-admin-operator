FROM openshift/origin-release:golang-1.10 AS build
ENV OPERATOR_PATH=/go/src/github.com/openshift/dedicated-admin-operator
COPY . ${OPERATOR_PATH}
WORKDIR ${OPERATOR_PATH}
RUN yum update -y --exclude=golang && yum clean all && make build

FROM openshift/origin-release:golang-1.10
ENV OPERATOR_PATH=/go/src/github.com/openshift/dedicated-admin-operator \
    OPERATOR_BIN=dedicated-admin-operator

RUN yum install ca-certificates && yum clean all
WORKDIR /root/
COPY --from=build ${OPERATOR_PATH}/build/_output/bin/${OPERATOR_BIN} /usr/local/bin/${OPERATOR_BIN}
LABEL io.openshift.managed.name="dedicated-admin-operator" \
      io.openshift.managed.description="Operator to manage permissions for Dedicated admins"
      