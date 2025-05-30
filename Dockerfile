FROM golang:1.23-alpine AS build

RUN apk add --no-cache tzdata

# Set timezone ke Asia/Jakarta
ENV TZ=Asia/Jakarta
RUN cp /usr/share/zoneinfo/Asia/Jakarta /etc/localtime && echo "Asia/Jakarta" > /etc/timezone

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Make sure .env file is copied
COPY .env .env


RUN go build -o main cmd/api/main.go

FROM alpine:3.20.1 AS prod

RUN apk add --no-cache tzdata

ENV TZ=Asia/Jakarta
RUN cp /usr/share/zoneinfo/Asia/Jakarta /etc/localtime && echo "Asia/Jakarta" > /etc/timezone


WORKDIR /app
COPY --from=build /app/main /app/main
COPY --from=build /app/.env /app/.env 
# ✅ Copy folder public
COPY --from=build /app/public ./public
COPY --from=build /app/docs ./docs
COPY --from=build /app/templates ./templates
COPY --from=build /app/files ./files
EXPOSE ${APP_PORT}
CMD ["./main"]


