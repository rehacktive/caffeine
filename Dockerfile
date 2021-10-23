FROM golang:1.17-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /caffeine

FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /caffeine /caffeine

EXPOSE 8000

USER nonroot:nonroot

ENTRYPOINT ["/caffeine"]