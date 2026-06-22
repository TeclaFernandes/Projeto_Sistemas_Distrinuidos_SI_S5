package main

import (
	"SASE-Projeto/internal/logger"
	"SASE-Projeto/internal/models"
	"SASE-Projeto/internal/queque"
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

var clientes = make(map[net.Conn]models.Client)
var contadorTA int

var clientesMutex sync.Mutex
var contadorMutex sync.Mutex
var filaMutex sync.Mutex

func main() {

	const PORTA = ":8080"

	logger.Info("Iniciando Server na porta " + PORTA)
	ouvinte, err := net.Listen("tcp", PORTA)

	if err != nil {
		panic(err)
	}

	fila := &queque.Fila{}

	for {

		conn, err := ouvinte.Accept()

		if err != nil {
			logger.Info("Conexao nao efetuada com sucesso")
			continue
		}

		go handleConnection(conn, fila)

	}

}

func handleConnection(conn net.Conn, fila *queque.Fila) {

	defer func() {
		conn.Close()

		clientesMutex.Lock()
		delete(clientes, conn)
		clientesMutex.Unlock()

	}()

	logger.Info("Nova conexao aceita")

	// INICIO DA VALIDACAO
	leitor := bufio.NewReader(conn)
	msg, err := leitor.ReadString('\n')

	escritor := bufio.NewWriter(conn)

	if err != nil {
		logger.Info("Conexao Encerrada...")
		return
	}

	msg = strings.TrimSpace(msg)
	partes := strings.SplitN(msg, ":", 2)

	if len(partes) < 2 {
		logger.Info("Validacao falhou: dados insuficientes ou formato invalido")
		return
	}

	comando := partes[0]
	id := partes[1]

	logger.Info("Validando novo cliente...")

	if comando != "REGISTER" {
		logger.Info("Validacao falhou: Primeiro comando nao solicitou REGISTER")
		return
	}

	tipoCliente, success := criaCliente(id, conn)

	if !success {
		logger.Info("Validacao falhou: REGISTERID fora dos parametros (TA,TS,TV)")
		enviaMsg(escritor, "FAIL")

		return
	}

	enviaMsg(escritor, "SUCCESS")
	logger.Info("Cliente validado, tipo: " + id)

	// LOOP PRINCIPAL -- WIP
	for {

		msg, err := recebeMsg(leitor)

		if err != nil {
			logger.Info("Cliente desconectado")
			return
		}

		partes := strings.SplitN(msg, ":", 2)

		if len(partes) < 1 {
			logger.Info("Erro: Comando enviado vazio")
			continue
		}

		comando := partes[0]

		switch comando {
		// BLOCO GERAR SENHAS
		case "GENERATE":

			// VALIDA SE O CLIENTE E O CORRETO
			if tipoCliente != "TS" {

				logger.Info("Requsicao GENERATE negada tipo: " + tipoCliente)
				enviaMsg(escritor, "DENIED")
				continue
			}

			if len(partes) < 2 {
				logger.Info("Erro: Comando GENERATE falta tipo de senha (P ou N)")
				continue
			}

			switch partes[1] {
			case "N":
				filaMutex.Lock()
				senha := fila.AdicionaNormal()
				filaMutex.Unlock()

				enviaMsg(escritor, fmt.Sprintf("CREATED:N%d", senha.Numero))
				logger.Info(fmt.Sprintf("Nova Senha criada: N%d", senha.Numero))

			case "P":
				filaMutex.Lock()
				senha := fila.AdicionaPrioritaria()
				filaMutex.Unlock()

				enviaMsg(escritor, fmt.Sprintf("CREATED:P%d", senha.Numero))
				logger.Info(fmt.Sprintf("Nova Senha criada: P%d", senha.Numero))

			default:
				logger.Info("Tipo de senha nao reconhecido pelo server: " + partes[1])
			}

		// BLOCO CHAMA PROXIMA SENHA
		case "NEXT":

			if tipoCliente != "TA" {

				logger.Info("Requsicao NEXT negada tipo: " + tipoCliente)
				enviaMsg(escritor, "DENIED")

				continue
			}

			filaMutex.Lock()
			senha, existe := fila.Next()
			filaMutex.Unlock()

			clientesMutex.Lock()
			guiche := clientes[conn].Guiche
			clientesMutex.Unlock()

			if !existe {
				logger.Info("Nao ha senhas na fila")

				enviaMsg(escritor, "NOTICKET")

				continue
			}

			broadCastTVs(senha, guiche)

			enviaMsg(escritor, fmt.Sprintf("CURRENTTICKET:%s%d", senha.Tipo, senha.Numero))

		default:
			logger.Info("Erro: comando desconhecido")
			continue
		}

	}

}

func broadCastTVs(senha queque.Senha, guiche int) {

	clientesMutex.Lock()

	clientesTV := make(map[net.Conn]models.Client)

	for conn, cliente := range clientes {
		clientesTV[conn] = cliente
	}

	clientesMutex.Unlock()

	for conn, cliente := range clientesTV {

		if cliente.Tipo != "TV" {
			continue
		}

		escritor := bufio.NewWriter(conn)

		resposta := fmt.Sprintf(
			"DISPLAY:%s%d:%d\n",
			senha.Tipo,
			senha.Numero,
			guiche,
		)

		enviaMsg(escritor, resposta)
	}
}

func criaCliente(id string, conn net.Conn) (string, bool) {

	switch id {
	case "TV", "TS":
		cliente := models.Client{
			Tipo: id,
		}

		clientesMutex.Lock()
		clientes[conn] = cliente
		clientesMutex.Unlock()
		return id, true

	case "TA":

		contadorMutex.Lock()

		contadorTA++
		guiche := contadorTA

		contadorMutex.Unlock()

		cliente := models.Client{
			Tipo:   id,
			Guiche: guiche,
		}

		clientesMutex.Lock()
		clientes[conn] = cliente
		clientesMutex.Unlock()

		return id, true

	default:
		return "", false

	}

}

func enviaMsg(escritor *bufio.Writer, msg string) {

	_, err := escritor.WriteString(msg + "\n")

	if err != nil {
		logger.Info("Falha ao enviar msg: " + msg)
		return
	}

	escritor.Flush()
}

func recebeMsg(leitor *bufio.Reader) (string, error) {

	msg, err := leitor.ReadString('\n')

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(msg), nil
}
