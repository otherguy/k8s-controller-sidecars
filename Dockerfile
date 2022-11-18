FROM golang:1.19.3 AS build
WORKDIR /go/src/github.com/otherguy/k8s-controller-sidecars
RUN apt-get update && apt-get install -y upx
ADD . .
RUN go get -v \
 && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix cgo -o main . \
 && upx main \
 && mkdir -p /empty

FROM scratch
COPY --from=build /go/src/github.com/otherguy/k8s-controller-sidecars/main /
COPY --from=build /empty /tmp
CMD ["/main"]
