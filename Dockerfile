FROM debian:10-slim
WORKDIR /app
ADD ./todo-go /app
ADD ./demodata /app/demodata

CMD ["/app/todo-go"]