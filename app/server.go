package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	// if it exists save the '--directory' flag value to the 'dir' variable
	var dir string
	flag.StringVar(&dir, "directory", "", "the directory to serve files from")
	flag.Parse()

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
		go handleConnection(conn, dir)
	}
}

func handleConnection(c net.Conn, dir string) {
	defer c.Close()
	reader := bufio.NewReader(c)
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}

	parts := strings.Split(strings.TrimSpace(line), " ")
	if len(parts) < 2 {
		fmt.Println("Invalid request line:", line)
		return
	}
	method, path := parts[0], parts[1]

	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading header:", err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "" { // Empty line indicates end of headers
			break
		}
		headerParts := strings.SplitN(line, ": ", 2)
		if len(headerParts) == 2 {
			headers[headerParts[0]] = headerParts[1]
		}
	}

	switch {
	case path == "/":
		c.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	case path == "/user-agent":
		print("User-Agent: ", headers["User-Agent"])
		c.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(headers["User-Agent"]), headers["User-Agent"])))
	case strings.HasPrefix(path, "/echo/"):
		echo := strings.TrimPrefix(path, "/echo/")
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
		c.Write([]byte(response))
	case strings.HasPrefix(path, "/files/") && (method == "GET" || method == "POST"):
		filePath := strings.TrimPrefix(path, "/files/")
		if method == "GET" {
			serveFile(c, dir, filePath)
		} else {
			saveFile(c, reader, dir, filePath)
		}
	default:
		c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func serveFile(c net.Conn, dir, filePath string) {
	fullPath := dir + "/" + filePath
	file, err := os.Open(fullPath)
	if err != nil {
		c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		c.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return
	}

	c.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n", fileInfo.Size())))
	io.Copy(c, file) // Stream the file content directly without loading into memory.
}

func saveFile(c net.Conn, reader *bufio.Reader, dir, filePath string) {
	fullPath := dir + "/" + filePath
	file, err := os.Create(fullPath)
	if err != nil {
		c.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return
	}
	defer file.Close()

	// For simplicity, assuming the rest of the request is the body. In a real application, you'd need to parse headers to find the Content-Length and read that many bytes.
	io.Copy(file, reader) // Stream the body directly into the file.

	c.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
}
