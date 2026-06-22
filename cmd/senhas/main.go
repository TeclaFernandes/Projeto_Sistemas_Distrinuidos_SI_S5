package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	registrar(conn)

	reader := bufio.NewReader(conn)

	resposta, _ := reader.ReadString('\n')

	fmt.Println(strings.TrimSpace(resposta))

	menu(conn)
}

func registrar(conn net.Conn) {
	writer := bufio.NewWriter(conn)

	writer.WriteString("REGISTER:TS\n")

	writer.Flush()
}

func menu(conn net.Conn) {
	input := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	for {
		fmt.Println("\n===== Totem =====")
		fmt.Println("1 - Senha Normal")
		fmt.Println("2 - Senha Prioritaria")
		fmt.Println("3 - Sair")

		var opcao string

		fmt.Print("> ")

		opcao, _ = input.ReadString('\n')
		opcao = strings.TrimSpace(opcao)

		switch opcao {
		case "1":
			writer.WriteString("GENERATE:N\n")
			writer.Flush()

		case "2":
			writer.WriteString("GENERATE:P\n")
			writer.Flush()

		case "3":
			return

		default:
			fmt.Println("Opcao invalida")
			continue
		}

		resposta, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println("Servidor desconectado")
			return
		}

		senha := strings.TrimSpace(resposta)

		partes := strings.SplitN(senha, ":", 2)

		if len(partes) == 2 {
			fmt.Println(
				"Sua senha:",
				partes[1],
			)
		}

	}
}
