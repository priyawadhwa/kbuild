# init
FROM scratch
ADD main /bin/
ADD cmd/executor/ca-certificates.crt /etc/ssl/certs/
ADD test/Dockerfile /dockerfile/