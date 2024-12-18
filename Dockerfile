FROM golang:1.21.1 AS build
WORKDIR /go/src/github.com/otherguy/k8s-controller-sidecars
RUN apt-get update && apt-get install -y upx-ucl
ADD . .
RUN go get -v \
  && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix cgo -o main . \
  && upx main \
  && mkdir -p /empty

FROM scratch
LABEL maintainer="Alexander Graf <hi@otherguy.io"
COPY --from=build /go/src/github.com/otherguy/k8s-controller-sidecars/main /
COPY --from=build /empty /tmp
ARG VCS_REF=main
ARG BUILD_DATE=""
ARG VERSION="${VCS_REF}"
LABEL org.label-schema.schema-version '1.0'
LABEL org.label-schema.name           'k8s-sidecars-cleanup-operator'
LABEL org.label-schema.description    'A Kubernetes Operator to clean up sidecar containers in finished jobs'
LABEL org.label-schema.vcs-url        'https://github.com/otherguy/k8s-controller-sidecars'
LABEL org.label-schema.version        '${VERSION}'
LABEL org.label-schema.build-date     '${BUILD_DATE}'
LABEL org.label-schema.vcs-ref        '${VCS_REF}'
ENV VCS_REF="${VCS_REF}"
ENV BUILD_DATE="${BUILD_DATE}"
ENV VERSION="${VERSION}"
CMD ["/main"]
