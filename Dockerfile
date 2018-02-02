# init
FROM scratch
ADD main /work-dir/
ADD cmd/executor/ca-certificates.crt /etc/ssl/certs/
ADD test/Dockerfile /dockerfile/
ADD cmd/executor/policy.json /etc/containers/