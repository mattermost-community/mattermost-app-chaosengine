# Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
# See LICENSE.txt for license information.

# Build the mattermost chaos-engine app
ARG DOCKER_BUILD_IMAGE=golang:1.16
ARG DOCKER_BASE_IMAGE=alpine:3.14
ARG BUILD_VERSION

FROM ${DOCKER_BUILD_IMAGE} AS build
WORKDIR /opt/chaosengine
COPY . /opt/chaosengine
RUN make build-linux

# Final Image
FROM ${DOCKER_BASE_IMAGE}

LABEL name="ChaosEngine" \
  maintainer="cloud-team@mattermost.com" \
  vendor="Mattermost" \
  distribution-scope="public" \
  architecture="x86_64" \
  url="https://mattermost.com" \
  io.k8s.description="ChaosEngine is a Mattermost App to run Chaos GameDays as part of your ChatOps." \
  io.k8s.display-name="Mattermost ChaosEngine"

ENV CHAOS=/chaos/mattermost-app-chaosengine \
    USER_UID=10001 \
    USER_NAME=chaos \
    CHAOS_DIR=/chaos

RUN apk update && apk add libc6-compat && apk add ca-certificates
COPY --from=build /opt/chaosengine/dist/mattermost-app-chaosengine /chaos/mattermost-app-chaosengine
COPY --from=build /opt/chaosengine/build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

WORKDIR /chaos

USER ${USER_UID}

EXPOSE  3000

ENTRYPOINT ["/usr/local/bin/entrypoint"]
