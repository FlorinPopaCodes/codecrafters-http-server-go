package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			// Print all incoming data.
			buf := make([]byte, 1024)
			_, err := c.Read(buf)
			if err != nil {
				fmt.Println("Error reading:", err.Error())

			}
			requestLines := strings.Split(string(buf), "\r\n")
			splitFirstLine := strings.Split(requestLines[0], " ")
			path := splitFirstLine[1]

			headers := make(map[string]string)
			// parse headers
			for _, header := range requestLines[1:] {
				if !strings.Contains(header, ": ") {
					continue
				}

				splitHeader := strings.Split(header, ": ")
				headers[splitHeader[0]] = splitHeader[1]
			}

			// Write the response.
			if path == "/" {
				c.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

			} else if strings.HasPrefix(path, "/echo") {
				echo := strings.Split(path, "/echo/")[1]
				response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s\r\n\r\n", len(echo), echo)
				c.Write([]byte(response))
			} else if path == "/user-agent" {
				response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s\r\n\r\n", len(headers["User-Agent"]), headers["User-Agent"])
				c.Write([]byte(response))
			} else {
				c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			}

			// Shut down the connection.
			c.Close()
		}(conn)
	}
}
