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

	go sendMessages(conn)

	recieveMessages(conn)

}

func sendMessages(conn net.Conn) {
	var opcao int
	defer conn.Close()
	inputReader := bufio.NewReader(os.Stdin)

	writer := bufio.NewWriter(conn)
	fmt.Println("Bem vindo, escolha um:")

	for {
		fmt.Println("Bem Vindo!")
		fmt.Println("1. Enviar mensagem\n2. Registrar-se\n 3. Sair")
		fmt.Scan(&opcao)

		switch opcao {

		case 1:
			fmt.Println("Escreva sua mensagem:")
			msg, err := inputReader.ReadString('\n')

			if err != nil {
				panic(err)
			}

			fmt.Println("Escreva seu destinatario:")

			dest, err := inputReader.ReadString('\n')

			if err != nil {
				panic(err)
			}

			dest = strings.TrimSpace(dest)
			msg = strings.TrimSpace(msg)

			mensagem := "MSG:" + dest + ":" + msg

			fmt.Println(mensagem)
			writer.WriteString(mensagem + "\n")
			writer.Flush()

			continue
		case 2:
			fmt.Println("Escreva seu nome:")
			nome, err := inputReader.ReadString('\n')

			if err != nil {
				panic(err)
			}

			mensagem := "REG:" + "REG" + ":" + nome

			writer.WriteString(mensagem + "\n")
			writer.Flush()
			continue
		case 3:
			conn.Close()
			return
		default:
			fmt.Println("Escolha uma Opcao valida")
			continue

		}

	}
}

func recieveMessages(conn net.Conn) {

	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println("Conn encerrada")
			return
		}

		fmt.Println(msg)

	}

}
