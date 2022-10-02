FROM golang

RUN go version
ENV GOPATH=/



COPY .. ./

RUN cp ./.env.example ./.env

# install psql
RUN apt-get update
RUN apt-get -y install postgresql-client

# make wait-for-postgres.sh executable
RUN chmod +x ./scripts/wait-for-postgres.sh

# build go app
RUN go mod download
RUN go build -o todo-app ./cmd/api/main.go

CMD ["./todo-app"]