FROM golang:alpine AS build

WORKDIR /opt/build

COPY . .

RUN go build main.go

FROM scratch AS prod

WORKDIR /opt/app

COPY --from=build /opt/build/main ./

CMD ["./main"]