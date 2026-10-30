FROM golang:1.27-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o app ./src

FROM alpine
WORKDIR /app
COPY --from=build /app/app .
CMD ["./app"]