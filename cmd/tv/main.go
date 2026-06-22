package main

import (
	"SASE-Projeto/internal/logger"
	"bufio"
	"fmt"
	"net"
	"strings"
)

const comando = "REGISTER:TV\n"

func main() {

	conn, err := net.Dial("tcp", "localhost:8080")

	if err != nil {
		panic(err)
	}

	registrar(conn)

	escutaMsg(conn)

}

func registrar(conn net.Conn) {

	escritor := bufio.NewWriter(conn)

	escritor.WriteString(comando)

	escritor.Flush()

}

func escutaMsg(conn net.Conn) {

	defer conn.Close()

	leitor := bufio.NewReader(conn)

	for {

		msg, err := leitor.ReadString('\n')

		if err != nil {
			fmt.Println(err)
			return
		}

		msg = strings.TrimSpace(msg)

		partes := strings.Split(msg, ":")

		if len(partes) == 3 {

			senha := partes[1]
			guiche := partes[2]

			logger.Info(
				fmt.Sprintf(
					"Senha atual: %s - Guichê %s",
					senha,
					guiche,
				),
			)

		}

	}

}
