# setup:

How to Run Kafka separately:

- Run the docker image of `Kafka` and `zoo-keeper` (image details in docker-compose.yml)

Run docker compose(recommended):

- cd notification_hub
- docker compose build
- docker compose up -d

See logs:

- docker compose logs -f api
- docker compose logs -f email

Running API and Worker Separately (Manual Build):

- docker build -f deployments/Dockerfile.api -t notification_hub:latest .
- docker run -p 8080:8080 notification_hub:latest

- docker build -f deployments/Dockerfile.email -t email:latest .
- docker run -p 3000:3000 email:latest

Stop everything:

- docker compose down

Example curl request:

```sh
curl -X POST http://localhost:8080/api/notify/signup \
     -H "Content-Type: application/json" \
     -d '{"username":"jsn", "password":"qwerty"}'

```

Response Example:

```json
{
  "message": "Signup event sent to Kafka"
}
```

Log:

```md
email-1 | 2025/07/10 09:54:26 jsn qwerty
```
