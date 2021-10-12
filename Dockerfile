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
FROM golang:stretch as go-builder

ENV GOPROXY=https://proxy.golang.org

WORKDIR /go/src/http-to-kafka
COPY . .
RUN go get -v -t -d ./...
RUN go build -v ./...
RUN go install -v ./...

# Final image
FROM debian:stretch
RUN apt -y update && apt -y upgrade && \
    apt install -y net-tools bash

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /usr/local/bin/tini
RUN chmod +x /usr/local/bin/tini

WORKDIR /http-to-kafka
COPY --from=go-builder /go/bin/http-to-kafka .
COPY http-server.* .

COPY entrypoint.sh /usr/local/bin
RUN chmod +x /usr/local/bin/entrypoint.sh

HEALTHCHECK --start-period=2s --interval=5s --timeout=2s --retries=5 CMD ["nc", "-z", "localhost", "8080"]
ENTRYPOINT ["/usr/local/bin/tini", "--", "/usr/local/bin/entrypoint.sh"]

EXPOSE 8080
EXPOSE 8433