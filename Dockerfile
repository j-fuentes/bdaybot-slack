FROM golang:1.12 AS build-env
ADD . /src
RUN cd /src && go build -o bdaybot

FROM gcr.io/distroless/base
COPY --from=build-env /src/bdaybot /bdaybot
ENTRYPOINT ["/bdaybot", "-logtostderr"]
