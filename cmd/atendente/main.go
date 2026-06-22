package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	conn, err := net.Dial(
		"tcp",
		"localhost:8080",
	)

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	registrar(conn)
	reader := bufio.NewReader(conn)
	resposta, _ := reader.ReadString('\n')

	fmt.Println(
		strings.TrimSpace(resposta),
	)

	menu(conn)
}

func registrar(conn net.Conn) {
	writer := bufio.NewWriter(conn)

	writer.WriteString(
		"REGISTER:TA\n",
	)

	writer.Flush()
}

func menu(conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		fmt.Println("\n===== ATENDIMENTO =====")
		fmt.Println("1 - Chamar senha")
		fmt.Println("2 - Sair")

		var opcao string
		fmt.Print("> ")
		fmt.Scanln(&opcao)

		switch opcao {
		case "1":
			writer.WriteString(
				"NEXT\n",
			)

			writer.Flush()

			resposta, err :=
				reader.ReadString('\n')

			if err != nil {
				return
			}

			senha := strings.TrimSpace(resposta)

			if senha == "NOTICKET" {

				fmt.Println(
					"Nenhuma senha aguardando atendimento.",
				)

				continue
			}

			partes := strings.SplitN(senha, ":", 2)

			if len(partes) == 2 {

				fmt.Println(
					"Atendendo:",
					partes[1],
				)

			}

		case "2":
			return

		default:
			fmt.Println(
				"Opcao invalida",
			)

		}

	}

}
