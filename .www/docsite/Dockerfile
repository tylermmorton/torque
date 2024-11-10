FROM alpine:latest AS gitrevision

RUN apk add git
RUN git clone https://github.com/tylermmorton/torque.git
WORKDIR /torque
RUN git rev-parse --short HEAD > .gitrevision

FROM node:14 AS node-build

WORKDIR /app
COPY . .

RUN npm ci
RUN npx tailwindcss -i ./styles/tailwind.css -o ./.build/static/app.css
RUN npx vite build

FROM golang:1.23.2-alpine AS torque-symbols

RUN apk add git
RUN git clone https://github.com/tylermmorton/torque.git
WORKDIR /go/torque/.www/docsite

ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GO111MODULE=on

RUN go mod tidy
RUN go run ./cmd/symbols/main.go -i ../../ -o content/symbols/torque.json

FROM golang:1.23.2-alpine AS go-build
WORKDIR /app

COPY . .
COPY --from=node-build /app/.build .build
COPY --from=gitrevision /torque/.gitrevision .gitrevision
COPY --from=torque-symbols /go/torque/.www/docsite/content/symbols/torque.json /content/symbols/torque.json

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