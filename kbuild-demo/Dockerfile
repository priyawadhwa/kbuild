FROM gcr.io/google-appengine/python:latest
COPY context/f* /context/

FROM gcr.io/google-appengine/debian9:latest
RUN apt-get update && apt-get install -y curl
COPY --from=0 /context/ foo/
