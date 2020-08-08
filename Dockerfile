# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.
FROM golang:1.12 as ks-apiserver-builder

COPY / /go/src/kubesphere.io/kubesphere

WORKDIR /go/src/kubesphere.io/kubesphere
RUN CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 GOFLAGS=-mod=vendor go build -i -ldflags '-w -s' -o ks-upgrade cmd/upgrade/upgrade.go

FROM alpine:3.9
RUN apk update && apk add bash bash-completion busybox-extras net-tools vim curl wget tcpdump ca-certificates && update-ca-certificates && rm -rf /var/cache/apk/* && curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.15.3/bin/linux/amd64/kubectl && chmod +x ./kubectl && mv ./kubectl /usr/local/bin/kubectl && echo -e 'source /usr/share/bash-completion/bash_completion\nsource <(kubectl completion bash)' >>~/.bashrc
COPY --from=ks-apiserver-builder /go/src/kubesphere.io/kubesphere/ks-upgrade /usr/local/bin/
CMD ["sh"]
