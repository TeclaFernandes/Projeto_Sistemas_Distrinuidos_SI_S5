package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

var clients = make(map[string]net.Conn)

func main() {

	fmt.Println("Iniciando conexao...")

	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		panic(err)
	}

	for {

		conn, err := listener.Accept()

		if err != nil {
			panic(err)
		}

		go handleConnection(conn)

	}
}

func handleConnection(conn net.Conn) {

	defer conn.Close()
	fmt.Println("Nova conexao aceita")

	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println("Conexao Encerrada: ", err)
			return
		}
		msg = strings.TrimSpace(msg)

		parts := strings.SplitN(msg, ":", 3)

		if len(parts) < 3 {
			continue
		}

		command := parts[0]
		target := parts[1]
		text := parts[2]

		switch command {
		case "REG":

			fmt.Println("Efetuando Registro de novo cliente")

			text = strings.TrimSpace(text)
			clients[text] = conn
		case "MSG":

			targetConn, exists := clients[target]

			if !exists {
				continue
			}

			fmt.Println(targetConn, exists)
			writerTarget := bufio.NewWriter(targetConn)
			writerTarget.WriteString(text + "\n")
			writerTarget.Flush()

		}

	}
}
