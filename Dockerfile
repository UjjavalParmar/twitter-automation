FROM golang:1.24 as build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/bot ./cmd/bot

FROM gcr.io/distroless/base-debian12
# ENV TZ=Asia/Kolkata
WORKDIR /app
COPY --from=build /bin/bot /bin/bot
COPY .env .env
VOLUME ["/data"]
ENTRYPOINT ["/bin/bot"]


# FROM golang:1.24
#
# WORKDIR /app
#
# COPY . .
#
# RUN go mod tidy && \
#     go build ./cmd/bot/main.go
#
# EXPOSE 8080
#
# CMD ["./main"]
