# init
FROM scratch
ADD main /work-dir/
ADD cmd/executor/ca-certificates.crt /etc/ssl/certs/
ADD cmd/executor/policy.json /etc/containers/
ADD appender/docker-credential-gcr_linux_amd64-1.4.1.tar.gz /usr/local/bin/
ADD appender/config.json /root/.docker/
ADD test/Dockerfile /dockerfile/