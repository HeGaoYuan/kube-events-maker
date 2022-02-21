FROM golang:1.16 AS builder

ADD . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GO11MODULE=on go build -a -o /main .

FROM gcr.io/distroless/base
COPY --from=builder --chown=nonroot:nonroot /main /kube-events-maker

USER nonroot

ENTRYPOINT ["/kube-events-maker"]
