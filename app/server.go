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

	// if it exists save the '--directory' flag value to the 'dir' variable
	dir := ""
	if len(os.Args) > 1 && os.Args[1] == "--directory" {
		dir = os.Args[2]
		println(dir)
	}

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
			method, path := splitFirstLine[0], splitFirstLine[1]

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
			} else if strings.HasPrefix(path, "/files") {
				filePath := strings.Split(path, "/files/")[1]

				if method == "GET" {
					file, err := os.Open(dir + "/" + filePath)
					if err != nil {
						c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
						c.Close()
						return
					}
					defer file.Close()

					fileInfo, err := file.Stat()
					if err != nil {
						c.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
						c.Close()
						return
					}

					fileSize := fileInfo.Size()
					fileContent := make([]byte, fileSize)
					_, err = file.Read(fileContent)
					if err != nil {
						c.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
						c.Close()
						return
					}

					response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s\r\n\r\n", fileSize, fileContent)
					c.Write([]byte(response))
				} else if method == "POST" {
					file, err := os.Create(dir + "/" + filePath)
					if err != nil {
						c.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
						c.Close()
						return
					}

					defer file.Close()

					bodyStartIndex := strings.Index(string(buf), "\r\n\r\n") + 4
					body := buf[bodyStartIndex:]
					_, err = file.Write(body)
					if err != nil {
						c.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
						c.Close()
						return
					}

					c.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
				} else {
					c.Write([]byte("HTTP/1.1 405 Method Not Allowed\r\n\r\n"))
				}

			} else {
				c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			}

			// Shut down the connection.
			c.Close()
		}(conn)
	}
}
