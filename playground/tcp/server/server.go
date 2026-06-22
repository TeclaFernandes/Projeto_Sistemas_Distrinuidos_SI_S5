package main

import (
	"fmt"
	"net"
)

func main() {

	// listener da porta, provavel que espera alguem se conectar com ele (cliente)
	// Retorna uma porta, e se houver erro nao da raise, bem diferente
	// primeiro arg: Tipo de conexao (tcp, tcp4, tcp6, unix or unixpacket)
	// Imagino que ":8080" possa significar qualuer IP ou IP gerado ou IP da maquina, enquanto a porta e 8080
	// Significa que pode definir um ip como 123.456.789.111:8080 ou algo assim
	listener, err := net.Listen("tcp", ":8080")

	// "panic" pode ser o mesmo que raise em python, considerando que para o codigo e recebe err como arg
	if err != nil {
		panic(err)
	}

	// Descoberto metodo Addr para mostrar address
	fmt.Println("Porta criada:", listener.Addr())

	// Basicamente, pelo que entendo, ele espera algum cliente se conectar e aceita, mas como esse programa e procedural,
	// fico sem entender como isso funciona
	conn, err := listener.Accept()

	if err != nil {
		panic(err)
	}

	fmt.Println("Cliente Conectado")

	// pelo nome, buffer seria o tempo para verificar uma mensagem do cliente?
	// Basicamente um array ou algo parecido, nao entendo
	buffer := make([]byte, 1024)

	// Parece ser o caso
	n, err := conn.Read(buffer)

	if err != nil {
		panic(err)
	}

	// Fico confuso nesse buffer agora sendo usado como arg na conversao de string... que por algum motivo se transforma em mensagem
	mensagem := string(buffer[:n])

	fmt.Println("Mensagem recebida:", mensagem)
}
