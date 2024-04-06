#    Copyright 2021 AERIS-Consulting e.U.
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.


# Builder: go-builder
ARG GO_VERSION=1.21
FROM --platform=${TARGETPLATFORM:-linux/amd64} golang:${GO_VERSION} as go-builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ENV GOPROXY=https://proxy.golang.org

WORKDIR /app/
ADD . .
RUN go get -v -t -d ./...
RUN GO_ENABLED=0 go build -ldflags="-w -s" -o /bin -v ./...

# Final image
FROM --platform=${TARGETPLATFORM:-linux/amd64} debian

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /usr/local/bin/tini
RUN chmod +x /usr/local/bin/tini

WORKDIR /http-to-kafka
COPY --from=go-builder /bin/http-to-kafka .
ADD http-server.* .

RUN useradd -ms /bin/bash http-to-kafka
RUN chown -R http-to-kafka:root /http-to-kafka && \
    chmod -R 770 /http-to-kafka

COPY entrypoint.sh /usr/local/bin
RUN chmod +x /usr/local/bin/entrypoint.sh

HEALTHCHECK --start-period=2s --interval=5s --timeout=2s --retries=5 CMD ["nc", "-z", "localhost", "8080"]
ENTRYPOINT ["/usr/local/bin/tini", "--", "/usr/local/bin/entrypoint.sh"]

EXPOSE 8080
EXPOSE 8433