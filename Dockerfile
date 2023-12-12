FROM golang:1.18.0-alpine3.15 AS builder
LABEL stage=builder
RUN apk add --no-cache bash git gcc libc-dev make cmake
ADD . /workspace/terraform-provider-ipm
RUN cd /workspace/terraform-provider-ipm \
    && make build install

FROM alpine:latest AS final
RUN apk add --no-cache bash terraform git
WORKDIR /ipm-services
COPY --from=builder /root/.terraform.d /root/.terraform.d
COPY --from=builder /workspace/terraform-provider-ipm/terraform-ipm-modules /ipm-services
RUN chmod +x /ipm-services/setup.sh 
RUN export PATH=$PATH:/ipm-services:/ipm-services/use-cases/commands
ENTRYPOINT [ "/ipm-services/setup.sh"]
