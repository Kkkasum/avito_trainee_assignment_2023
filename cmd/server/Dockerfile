FROM golang:1.24-alpine AS build
RUN apk add --no-cache make
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make .build

FROM alpine:latest
COPY --from=build /app/app /bin/app
CMD ["./bin/app"]