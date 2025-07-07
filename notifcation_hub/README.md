# setup:

- docker build -f deployments/Dockerfile.api -t notification_hub:latest .
- docker run -p 8080:8080 notification_hub:latest
