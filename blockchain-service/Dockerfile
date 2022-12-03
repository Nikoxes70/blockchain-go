FROM golang:1.13.3 AS contentonboarding

WORKDIR /code/blockchain/blockchain
COPY ./content-onboarding-service/go.mod .
COPY ./content-onboarding-service/go.sum .

RUN go mod download
COPY ./content-onboarding-service /code/mmservices/content-onboarding-service

RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -v ./cmd/contentonboardingweb

FROM ubuntu:18.04
RUN apt-get update && apt-get install -y curl
ENV WORK_DIR /content-onboarding-service
RUN mkdir -p $WORK_DIR
RUN mkdir -p $WORK_DIR/logs
WORKDIR $WORK_DIR

COPY --from=contentonboarding /code/mmservices/content-onboarding-service/contentonboardingweb contentonboardingweb
COPY --from=contentonboarding /code/mmservices/content-onboarding-service/config config

ENV ENV development
ENV CONFIG_FOLDER_PATH config/development
ENV SERVICE_NAME content-onboarding-service
ENV DATADOG_TRACE_AGENT_HOSTNAME '172.17.0.1:8126'

CMD ./contentonboardingweb  -configFolder=$CONFIG_FOLDER_PATH \
                            -datadogAgent=$DATADOG_TRACE_AGENT_HOSTNAME
