package main

import (
	"net"
)

func main() {

	// Dial... okay, isso e entao o metodo de conexao com um server, ou seja, cliente
	conn, err := net.Dial("tcp", "localhost:8080")

	if err != nil {
		panic(err)
	}

	mensagem := "Olá Servidor"

	// entendi, converte em byter pois conexoes tcp transmitem em bytes
	conn.Write([]byte(mensagem))
}
