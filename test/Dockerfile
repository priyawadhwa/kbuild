FROM gcr.io/google-appengine/debian9:latest
RUN echo "hey" > /etc/foo
RUN echo "baz" > /etc/foo
RUN echo "baz" > /etc/foo2

COPY ./pkg/image/ foo
COPY ./pkg/*/stor*.go /storage/