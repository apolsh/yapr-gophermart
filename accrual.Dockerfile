FROM golang:1.18-alpine
WORKDIR /application

ADD ./cmd/accrual/accrual_linux_amd64 /application/accrual_linux_amd64

RUN apk add --no-cache libc6-compat
CMD ["/application/accrual_linux_amd64", "-a=accrual:8282"]