FROM node:14 AS node-build
WORKDIR /app

COPY . .

RUN npm ci
RUN npx tailwindcss -i ./styles/tailwind.css -o ./.build/static/app.css

FROM golang:1.23.2-alpine AS go-build
WORKDIR /app

COPY . .
COPY --from=node-build /app/.build .build

ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GO111MODULE=on

RUN go mod tidy

RUN go build -v -o ./.build/bin/torque ./app.go
RUN chmod +x ./.build/bin/torque

FROM alpine:latest AS production
COPY --from=go-build /app/.build/bin/torque /bin/torque

EXPOSE 8080
ENTRYPOINT [ "/bin/torque" ]