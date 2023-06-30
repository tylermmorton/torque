FROM node:14 as node-build

WORKDIR /app
COPY . .

RUN npm ci
RUN npx tailwindcss -i ./styles/tailwind.css -o ./.build/static/app.css

FROM golang:1.20-alpine as go-build
COPY . .
COPY ../go.mod .
COPY ../go.sum .
COPY --from=node-build /app/.build .build

ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV TMPL_BIND_TYPE="embed"

RUN go mod tidy
RUN go generate ./...
RUN go build -v -o ./torque ./cmd/torque
RUN chmod +x ./torque

FROM alpine:latest as production
COPY --from=go-build /app/torque /bin/torque

EXPOSE 3030
ENTRYPOINT [ "/bin/torque" ]