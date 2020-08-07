FROM golang
MAINTAINER Mimoja <git@mimoja.de>
 
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o main .

CMD ["/app/main"]
