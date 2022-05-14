FROM golang:1.18-alpine as build

ARG VERSION_PKG=kexplain/pkg/version
ARG VERSION
ARG GIT_COMMIT

WORKDIR /app

COPY go.mod .
COPY go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -o /kexplain \
  -ldflags "-X ${VERSION_PKG}.version=${VERSION} -X ${VERSION_PKG}.gitCommit=${GIT_COMMIT}" \
  ./cmd/*.go

#####################

FROM alpine:3.15

WORKDIR /

COPY --from=build /kexplain /bin/kexplain

ENTRYPOINT ["/bin/kexplain"]
