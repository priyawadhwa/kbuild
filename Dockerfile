FROM gcr.io/google-appengine/debian9:latest
ADD main /work-dir/
ADD test/Dockerfile /dockerfile/