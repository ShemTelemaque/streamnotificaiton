# Running with Podman

This guide explains how to run the Stream Notification application using Podman instead of Docker.

## Prerequisites

- [Podman](https://podman.io/getting-started/installation) installed on your system
- [Podman Compose](https://github.com/containers/podman-compose) (optional, for using docker-compose.yml)

## Option 1: Using Podman Compose

Podman Compose provides a Docker Compose compatible experience:

1. Clone the repository and navigate to the project directory

2. Build and start the containers:
   ```bash
   podman-compose up -d
   ```

3. View logs:
   ```bash
   podman-compose logs -f
   ```

4. Stop the containers:
   ```bash
   podman-compose down
   ```

## Option 2: Using Podman Directly

If you prefer not to use Podman Compose, you can use Podman commands directly:

1. Create a pod for the application:
   ```bash
   podman pod create --name streamnotification-pod -p 8080:8080 -p 5432:5432
   ```

2. Create a volume for PostgreSQL data:
   ```bash
   podman volume create postgres_data
   ```

3. Run the PostgreSQL container:
   ```bash
   podman run -d --pod streamnotification-pod \
     --name streamnotification-db \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=streamnotification \
     -v postgres_data:/var/lib/postgresql/data \
     postgres:14-alpine
   ```

4. Build the application image:
   ```bash
   podman build -t streamnotification-app .
   ```

5. Run the application container:
   ```bash
   podman run -d --pod streamnotification-pod \
     --name streamnotification-app \
     -e PORT=8080 \
     -e ENVIRONMENT=production \
     -e DB_HOST=localhost \
     -e DB_PORT=5432 \
     -e DB_USER=postgres \
     -e DB_PASSWORD=postgres \
     -e DB_NAME=streamnotification \
     -e TWITCH_CLIENT_ID=your_client_id \
     -e TWITCH_CLIENT_SECRET=your_client_secret \
     -e DISCORD_BOT_TOKEN=your_discord_bot_token \
     streamnotification-app
   ```

   Note: Replace the Twitch and Discord credentials with your actual values.

6. Check the running containers:
   ```bash
   podman ps
   ```

7. View logs:
   ```bash
   podman logs -f streamnotification-app
   ```

8. Stop and remove containers:
   ```bash
   podman pod stop streamnotification-pod
   podman pod rm streamnotification-pod
   ```

## Important Notes for Podman

1. **Rootless Mode**: Podman can run in rootless mode, which is more secure. The commands above work in both rootless and root mode.

2. **SELinux**: If you're using SELinux, you might need to add `:z` or `:Z` to volume mounts:
   ```bash
   -v postgres_data:/var/lib/postgresql/data:Z
   ```

3. **Network**: When using a pod, containers can communicate with each other using `localhost` instead of service names.

4. **Windows Users**: If you're using Podman on Windows, you might need to use the WSL2 backend.

## Troubleshooting

- If you encounter permission issues with volumes, try running Podman with root privileges or adjust the SELinux context.
- If the application can't connect to the database, ensure the database container is running and the connection parameters are correct.
- For networking issues, check that the containers are in the same pod or network.