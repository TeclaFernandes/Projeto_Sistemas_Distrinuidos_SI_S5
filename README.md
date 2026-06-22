# Sistema de Atendimento Distribuído

## 1. Introdução

Este projeto é um sistema de atendimento distribuído desenvolvido em Go para fins acadêmicos. Ele demonstra comunicação TCP entre um servidor de controle e três tipos de terminais: totens de senha (`TS`), terminais de atendimento (`TA`) e terminais de exibição (`TV`).

O sistema organiza a geração e o atendimento de senhas, com prioridade para senhas de atendimento urgente, e aplica uma regra simples de balanceamento entre senhas normais e prioritárias.

## 2. Objetivos

- Implementar um servidor TCP capaz de autenticar e diferenciar tipos de terminais.
- Criar um protocolo textual simples para comunicação entre cliente e servidor.
- Desenvolver uma fila de senhas com priorização.
- Demonstrar concorrência e sincronização em Go usando goroutines e mutexes.

## 3. Estrutura do projeto

```text
sase/
├── cmd/
│   ├── atendente/
│   │   └── main.go
│   ├── senhas/
│   │   └── main.go
│   ├── server/
│   │   └── main.go
│   └── tv/
│       └── main.go
├── internal/
│   ├── logger/
│   │   └── logger.go
│   ├── models/
│   │   └── cliente.go
│   └── queque/
│       ├── queque.go
│       └── senha.go
├── playground/
├── go.mod
└── README.md
```

## 3. Especie de Diagrama do projeto

```text
                +----------------+
                |      TS        |
                | TerminalSenha  |
                +----------------+
                        |
                        |
                   TCP Socket
                        |
                        v

                +----------------+
                |      SRV       |
                |    Servidor    |
                +----------------+
                   /          \
                  /            \
                 v              v

        +---------------+   +---------------+
        |      TA       |   |      TV       |
        | Atendimento   |   | Visualização  |
        +---------------+   +---------------+
```

## 4. Componentes do sistema

### 4.1 Servidor (`cmd/server/main.go`)

O servidor central é responsável por:

- Aceitar conexões TCP na porta `8080`.
- Verificar o registro inicial de cada terminal.
- Processar comandos de geração de senhas e de chamada de próxima senha.
- Enviar mensagens de display para os terminais `TV`.
- Manter o estado da fila de senhas de forma concorrente.

O servidor utiliza os pacotes internos para registro de clientes, fila de senhas e logs.

### 4.2 Totem de senhas (`cmd/senhas/main.go`)

O totêm de senhas (`TS`) é um cliente que permite a criação de senhas:

- `GENERATE:N` para senha normal.
- `GENERATE:P` para senha prioritária.

Após o comando, o cliente recebe uma resposta do servidor indicando a senha criada.

### 4.3 Terminal de atendimento (`cmd/atendente/main.go`)

O terminal de atendimento (`TA`) solicita a próxima senha disponível usando o comando:

- `NEXT`

Em retorno, ele recebe: `CURRENTTICKET:<tipo><número>` ou `NOTICKET` quando não existem senhas.

### 4.4 Terminal de exibição (`cmd/tv/main.go`)

O terminal de exibição (`TV`) permanece conectado ao servidor e recebe atualizações do tipo:

- `DISPLAY:<tipo><número>:<guiche>`

Essa mensagem deve ser exibida em tempo real para informar a senha chamada e o guichê de atendimento.

## 5. Protocolos de comunicação

### 5.1 Formato geral

Todos os comandos e respostas são enviados como linhas de texto terminadas em `\n`.

Formato padrão:

```text
COMANDO:ATRIBUTO1:ATRIBUTO2\n
```

### 5.2 Registro inicial

O primeiro comando enviado por qualquer terminal deve ser de registro:

```text
REGISTER:TV
REGISTER:TS
REGISTER:TA
```

Se o registro for válido, o servidor responde com `SUCCESS`.

### 5.3 Comandos do terminal TS

- `GENERATE:N` — cria uma nova senha normal.
- `GENERATE:P` — cria uma nova senha prioritária.

Resposta esperada:

- `CREATED:N<number>`
- `CREATED:P<number>`

### 5.4 Comando do terminal TA

- `NEXT` — solicita a próxima senha da fila.

Respostas possíveis:

- `CURRENTTICKET:<tipo><número>`
- `NOTICKET`

### 5.5 Mensagem para TVs

- `DISPLAY:<tipo><número>:<guiche>`

Exemplo:

```text
DISPLAY:N12:2
```

### 5.6 Códigos de status

- `SUCCESS` — cliente registrado com sucesso.
- `FAIL` — falha na autenticação ou registro.
- `DENIED` — comando enviado por terminal não autorizado.

## 6. Modelo de fila de senhas

A lógica de atendimento está em `internal/queque/queque.go`.

### 6.1 Tipos de senhas

- `N` — senha normal.
- `P` — senha prioritária.

### 6.2 Regra de atendimento

1. Atende até duas senhas normais consecutivas.
2. Após duas senhas normais, atende uma senha prioritária se houver.
3. Se não houver senhas normais, atende a próxima prioritária.
4. Se não houver nenhuma senha, retorna `NOTICKET`.

Essa política garante que senhas prioritárias sejam atendidas com frequência, mas sem deixar as senhas normais sem progresso.

## 7. Descrição dos pacotes internos

### 7.1 `internal/queque`

- `Fila` — guarda as filas de senhas normais e prioritárias.
- `AdicionaNormal()` — cria e adiciona uma senha normal.
- `AdicionaPrioritaria()` — cria e adiciona uma senha prioritária.
- `Next()` — retorna a próxima senha a ser atendida.

### 7.2 `internal/models`

- `Client` — estrutura que representa um cliente conectado, com campos `Tipo` e `Guiche`.

### 7.3 `internal/logger`

- `Info(msg string)` — função de log para imprimir mensagens com timestamp.

## 8. Como compilar

1. Abra o terminal na pasta `sase/`.
2. Compile o servidor:

```bash
go build -o bin/server ./cmd/server
```

3. Compile os terminais:

```bash
go build -o bin/ts ./cmd/senhas
go build -o bin/ta ./cmd/atendente
go build -o bin/tv ./cmd/tv
```

## 9. Como executar

1. Execute o servidor:

```bash
./bin/server
```

2. Em terminais separados, execute os clientes:

```bash
./bin/tv
./bin/ts
./bin/ta
```

3. No terminal `TS`, gere senhas.
4. No terminal `TA`, chame a próxima senha.
5. O terminal `TV` exibe a senha e o guichê.

## 10. Exemplo de uso

1. Inicie o servidor.
2. Abra um `TV` para receber atualizações de display.
3. Abra um `TS` e gere senhas: `1` para normal, `2` para prioritária.
4. Abra um `TA` e pressione `1` para chamar a próxima senha.
5. Observe o `TV` mostrar `DISPLAY:<tipo><número>:<guiche>`.

## 11. Conclusão

Este projeto oferece uma aplicação distribuída que integra comunicação de rede, sincronização concorrente e lógica de prioridade em filas. A documentação e a arquitetura permitem extensões futuras para interfaces mais sofisticadas, persistência e autenticação.

### Status do comando Generate

Para o cliente TS:

```js
CREATED // Retorno da criação de senha
Ex.: CREATED:N12
```

# Status do comando Next

Para o cliente TA:

```js
CURRENTTICKET // Retorno da senha atual Ex.: CURRENTTICKET:N25

NOTICKET // Não há mais senhas a serem chamadas
```