# Docker Container Log Viewer

Docker Container Log Viewer is a web application that allows you to view the logs of your Docker containers in real-time.

## Prerequisites

- Docker
- Docker Compose

## Usage

1. Clone this repository:
`git clone https://github.com/andewkuehne/docker-container-log-viewer.git`


2. Navigate to the project directory:
`cd docker-container-log-viewer`



3. Start the application using Docker Compose:
`docker-compose up`



4. Open your web browser and go to `http://localhost:8080`. You will see a list of all your running Docker containers. Click on the logs link next to the container you want to view logs for.

## Troubleshooting

If you encounter any issues, please check the logs of the `docker-container-log-viewer` service in the Docker Compose output for hints on what might be causing the problem.

## License

This project is licensed under the GPL License - see the LICENSE file for details.
