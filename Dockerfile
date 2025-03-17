FROM golang:1.24.1-alpine AS builder
WORKDIR /src
COPY . .

# Install musl-dev to make sure we get the required libraries
RUN apk add --no-cache gcc musl-dev libc6-compat make
RUN make build

FROM alpine
COPY --from=builder /src/bin/micro-mysql /

EXPOSE 3306
EXPOSE 5432

ENTRYPOINT ["/micro-mysql"]
CMD []
