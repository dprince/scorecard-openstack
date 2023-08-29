# Build the scorecard-openstack binary
FROM --platform=$BUILDPLATFORM golang:1.19 as builder
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o scorecard-openstack /workspace/main.go

# Final image.
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.8
#FROM gcr.io/distroless/static:nonroot

ENV HOME=/opt/scorecard-openstack \
    USER_NAME=scorecard-openstack \
    USER_UID=1001

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd

WORKDIR ${HOME}

COPY --from=builder /workspace/scorecard-openstack /usr/local/bin/scorecard-openstack

ENTRYPOINT ["/usr/local/bin/scorecard-openstack"]

USER ${USER_UID}
