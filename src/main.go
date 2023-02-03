package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/websocket"
)

var frontend = `
<html>
	<head>
		<script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/socket.io/2.3.0/socket.io.js"></script>
		<script type="text/javascript" src="https://code.jquery.com/jquery-3.6.0.min.js"></script>
		<style>
			body {
				background-color: black;
				color: green;
				font-family: "Courier New", Courier, monospace;
			}

			table {
				width: 100%;
				border-collapse: collapse;
				margin: 0 auto;
			}

			th, td {
				border: 1px solid green;
				padding: 8px;
				text-align: left;
			}

			th {
				background-color: green;
				color: black;
			}
		</style>
	</head>
	<body>
		<table id="container-table">
			<tr id="container-names"></tr>
			<tr id="container-logs"></tr>
		</table>
		<script>
			const socket = io.connect("http://localhost:8080");

			$.get("http://localhost:8080/containers", containers => {
				containers.forEach(container => {
					const containerNameColumn = $("<th>").text(container);
					$("#container-names").append(containerNameColumn);

					const containerLogColumn = $("<td>")
						.css({
							width: "100%",
							height: "400px",
							overflow: "auto"
						});
					$("#container-logs").append(containerLogColumn);

					socket.on("connect", () => {
						const logSocket = socket.of("/logs?container=${container}");
						logSocket.on("connect", () => {
							logSocket.on("message", data => {
								containerLogColumn.text(containerLogColumn.text() + data);
							});
						});
					});
				});
			});
		</script>
	</body>
</html>

`

// containersEndpoint is a HTTP endpoint that returns the list of container names running on the host
func containersEndpoint(w http.ResponseWriter, r *http.Request) {
	// Create a new Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		// If there is an error creating the client, return a 500 Internal Server Error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// List all the containers running on the host
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		// If there is an error getting the list of containers, return a 500 Internal Server Error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract the names of all the containers and store them in a slice
	var containerNames []string
	for _, container := range containers {
		// The name of the container is stored in the "Names" field, in the format "/containerName"
		containerNames = append(containerNames, container.Names[0][1:])
	}

	// Set the response header to indicate that the response will be in JSON format
	w.Header().Set("Content-Type", "application/json")
	// Encode the container names as JSON and write them to the response body
	json.NewEncoder(w).Encode(containerNames)
}

// logsEndpoint is a WebSocket endpoint that streams the logs of a container to the client
func logsEndpoint(ws *websocket.Conn) {
	// Get the name of the container to retrieve logs for from the query parameters
	container := ws.Request().URL.Query().Get("container")
	if container == "" {
		// If the container name is not provided, send an error message to the client and close the connection
		ws.Write([]byte("container name must be provided as a query parameter"))
		ws.Close()
		return
	}

	// Create a new Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		// If there is an error creating the client, send the error message to the client and close the connection
		ws.Write([]byte(err.Error()))
		ws.Close()
		return
	}

	// Retrieve the logs for the specified container
	logs, err := cli.ContainerLogs(context.Background(), container, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		// If there is an error getting the logs, send the error message to the client and close the connection
		ws.Write([]byte(err.Error()))
		ws.Close()
		return
	}

	// Copy the logs to the WebSocket connection
	_, err = io.Copy(ws, logs)
	if err != nil {
		ws.Write([]byte(err.Error()))
		// Close WebSocket connection
		ws.Close()
		return
	}
}

// main is the main function of the application.
func main() {
	// mux is a ServeMux that will handle incoming HTTP requests.
	mux := http.NewServeMux()

	// Register the containersEndpoint function to handle incoming requests to "/containers".
	mux.HandleFunc("/containers", containersEndpoint)

	// Register the logsEndpoint function to handle incoming WebSocket connections to "/logs".
	mux.Handle("/logs", websocket.Handler(logsEndpoint))

	// Register a function to handle incoming requests to "/".
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Parse a HTML template stored in a variable named frontend.
		tmpl, err := template.New("frontend").Parse(frontend)
		if err != nil {
			// If there was an error parsing the template, return a 500 Internal Server Error response.
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Execute the parsed template and write the result to the HTTP response.
		err = tmpl.Execute(w, nil)
		if err != nil {
			// If there was an error executing the template, return a 500 Internal Server Error response.
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Start a HTTP server that listens on port 8080 and uses the ServeMux to handle incoming requests.
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		// If there was an error starting the server, print the error to the console.
		fmt.Println(err)
	}
}
