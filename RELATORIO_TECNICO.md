# Relatório Técnico: Sistema de Atendimento Distribuído (SASE)

## Autores

- **Tecla Fernandes Oliveora** — Graduanda em Sistemas de Informações

- **Gustavo Eliel de Figueiredo Silva** — Graduando em Sistemas de Informações

- **Fernando Barbosa da Silva** — Graduando em Sistemas de Informações

---

## Informações do Projeto

**Data:** 16 de junho de 2026  
**Instituição:** Instituto Federal de Educação, Ciência e Tecnologia do Ceará (IFCE)  
**Disciplina:** Sistemas Distribuídos (5º Semestre)  
**Projeto:** Sistema de Atendimento Distribuído  
**Linguagem:** Go 1.26.3  

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Arquitetura do Sistema](#2-arquitetura-do-sistema)
3. [Detalhamento das Funções](#3-detalhamento-das-funções)
4. [Soluções de Software Adotadas](#4-soluções-de-software-adotadas)
5. [Recursos de Interface](#5-recursos-de-interface)
6. [Protocolos de Comunicação](#6-protocolos-de-comunicação)
7. [Sincronização e Concorrência](#7-sincronização-e-concorrência)
8. [Fluxos de Dados](#8-fluxos-de-dados)
9. [Análise de Estruturas de Dados](#9-análise-de-estruturas-de-dados)
10. [Considerações de Desempenho](#10-considerações-de-desempenho)
11. [Tratamento de Erros](#11-tratamento-de-erros)
12. [Conclusões](#12-conclusões)

---

## 1. Visão Geral

O Sistema de Atendimento Distribuído (SASE) é uma aplicação cliente-servidor desenvolvida em Go que demonstra conceitos fundamentais de sistemas distribuídos:

- **Comunicação de rede:** Sockets TCP para troca de mensagens entre componentes.
- **Arquitetura distribuída:** Um servidor central coordena múltiplos clientes heterogêneos.
- **Processamento concorrente:** Uso de goroutines para atender múltiplas conexões simultâneas.
- **Sincronização:** Mutexes para evitar condições de corrida (race conditions).
- **Protocolo customizado:** Protocolo textual simples baseado em comandos e atributos.
- **Lógica de priorização:** Fila com duas prioridades atendidas segundo regra determinística.

O sistema simula um ambiente de atendimento (banco, hospital, farmácia) onde clientes recebem senhas e são chamados para atendimento em guichês específicos.

---

## 2. Arquitetura do Sistema

### 2.1 Visão Geral da Arquitetura

```
┌─────────────────────────────────────────────────────────────┐
│                     SERVIDOR CENTRAL                        │
│                   (cmd/server/main.go)                      │
│                    Porta: 8080 (TCP)                        │
└─────────┬────────────────────────────┬──────────────────────┘
          │                            │
          │ Goroutines para cada      │
          │ conexão de cliente        │
          │                            │
    ┌─────▼─────┐  ┌──────────────┐  ┌▼─────────────┐
    │  TS (Totem │  │  TA          │  │  TV          │
    │  de Senha) │  │  (Atendente) │  │  (Exibição)  │
    └───────────┘  └──────────────┘  └──────────────┘
         │                │                    │
         │                │             (Recebe Broadcast)
         │                │
    GENERATE:N/P       NEXT             DISPLAY:<tipo>:<guiche>
    GENERATE:P         NEXT
```

### 2.2 Componentes Principais

**1. Servidor (`cmd/server/main.go`)**
- Listener TCP na porta 8080
- Aceita múltiplas conexões simultâneas
- Cria uma goroutine para cada cliente
- Gerencia fila compartilhada de senhas
- Sincroniza acesso a recursos com mutexes

**2. Clientes**
- `cmd/senhas/main.go` (TS): Gera senhas
- `cmd/atendente/main.go` (TA): Chama próxima senha
- `cmd/tv/main.go` (TV): Recebe e exibe mensagens

**3. Pacotes Internos**
- `internal/queque/`: Lógica de fila com priorização
- `internal/models/`: Estruturas de dados
- `internal/logger/`: Função de logging com timestamp

---

## 3. Detalhamento das Funções

### 3.1 Servidor Principal (`cmd/server/main.go`)

#### **`func main()`**

```go
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
```

**Descrição:**
- Inicializa o listener TCP na porta 8080.
- Cria uma instância da fila compartilhada.
- Executa um loop infinito aceitando conexões.
- Para cada conexão, dispara uma goroutine que chama `handleConnection()`.

**Decisão de Design:** O uso de goroutines permite que o servidor atenda múltiplos clientes concorrentemente sem bloquear.

---

#### **`func handleConnection(conn net.Conn, fila *queque.Fila)`**

```go
func handleConnection(conn net.Conn, fila *queque.Fila) {
    defer func() {
        conn.Close()
        clientesMutex.Lock()
        delete(clientes, conn)
        clientesMutex.Unlock()
    }()
    
    leitor := bufio.NewReader(conn)
    msg, err := leitor.ReadString('\n')
    // ... validação de REGISTER
    
    // Loop principal de processamento de comandos
    for {
        msg, err := recebeMsg(leitor)
        // ... switch para GENERATE ou NEXT
    }
}
```

**Descrição:**
- Gerencia a conexão de um cliente individual.
- Valida o registro inicial (`REGISTER:tipo`).
- Processa comandos em loop até desconexão.
- Libera recursos e remove cliente do mapa no encerramento (defer).

**Fases:**
1. **Validação de registro:** Verifica se o primeiro comando é `REGISTER` com tipo válido.
2. **Criação de cliente:** Associa tipo (`TV`, `TS`, `TA`) e número de guichê (para `TA`).
3. **Loop de processamento:** Recebe e processa comandos indefinidamente.

---

#### **`func criaCliente(id string, conn net.Conn) (string, bool)`**

```go
func criaCliente(id string, conn net.Conn) (string, bool) {
    switch id {
    case "TV", "TS":
        cliente := models.Client{ Tipo: id }
        clientesMutex.Lock()
        clientes[conn] = cliente
        clientesMutex.Unlock()
        return id, true
    case "TA":
        contadorMutex.Lock()
        contadorTA++
        guiche := contadorTA
        contadorMutex.Unlock()
        
        cliente := models.Client{ Tipo: id, Guiche: guiche }
        clientesMutex.Lock()
        clientes[conn] = cliente
        clientesMutex.Unlock()
        return id, true
    default:
        return "", false
    }
}
```

**Descrição:**
- Valida o tipo de cliente.
- Para `TA`, atribui um número de guichê único (contador incrementado).
- Armazena cliente no mapa global `clientes`.
- Usa lock/unlock para evitar race condition.

**Lógica de Guichês:**
- Cada `TA` recebe um número único crescente.
- Usado para identificar qual guichê chamou a senha.

---

#### **`func broadCastTVs(senha queque.Senha, guiche int)`**

```go
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
        resposta := fmt.Sprintf("DISPLAY:%s%d:%d\n", senha.Tipo, senha.Numero, guiche)
        enviaMsg(escritor, resposta)
    }
}
```

**Descrição:**
- Envia mensagem `DISPLAY` para todos os clientes `TV` conectados.
- Cria cópia do mapa de clientes para iterar fora do lock.
- Filtra apenas clientes do tipo `TV`.
- Envia a senha e o número de guichê.

**Decisão de Design:** 
- Cria cópia local do mapa para minimizar tempo de lock.
- Evita deadlock ao tentar enviar mensagem enquanto mantém lock.

---

#### **`func enviaMsg(escritor *bufio.Writer, msg string) e recebeMsg(leitor *bufio.Reader)`**

```go
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
```

**Descrição:**
- `enviaMsg`: Envia mensagem com newline e executa flush.
- `recebeMsg`: Recebe mensagem até newline e remove espaçamento.

**Utilidade:** Abstraem a comunicação TCP, garantindo consistência no protocolo.

---

### 3.2 Fila de Senhas (`internal/queque/queque.go`)

#### **`func (q *Fila) AdicionaNormal() Senha`**

```go
func (q *Fila) AdicionaNormal() Senha {
    q.ContadorNormal++
    senha := Senha{ Tipo: "N", Numero: q.ContadorNormal }
    q.FilaNormal = append(q.FilaNormal, senha)
    return senha
}
```

**Descrição:**
- Incrementa contador de senhas normais.
- Cria nova estrutura `Senha` com tipo "N".
- Adiciona ao final da fila.
- Retorna a senha criada para resposta ao cliente.

---

#### **`func (q *Fila) AdicionaPrioritaria() Senha`**

```go
func (q *Fila) AdicionaPrioritaria() Senha {
    q.ContadorPrioritario++
    senha := Senha{ Tipo: "P", Numero: q.ContadorPrioritario }
    q.FilaPrioritaria = append(q.FilaPrioritaria, senha)
    return senha
}
```

**Descrição:**
- Idêntico a `AdicionaNormal`, mas para senhas prioritárias.
- Mantém contador separado.
- Armazena em fila separada.

---

#### **`func (q *Fila) Next() (Senha, bool)`**

```go
func (q *Fila) Next() (Senha, bool) {
    // Regra: após 2 normais, atende 1 prioritária
    if q.NormalServido >= 2 && len(q.FilaPrioritaria) > 0 {
        senha := q.FilaPrioritaria[0]
        q.FilaPrioritaria = q.FilaPrioritaria[1:]
        q.NormalServido = 0
        return senha, true
    }
    
    // Atende senha normal
    if len(q.FilaNormal) > 0 {
        senha := q.FilaNormal[0]
        q.FilaNormal = q.FilaNormal[1:]
        q.NormalServido++
        return senha, true
    }
    
    // Se não há normal, atende prioritária
    if len(q.FilaPrioritaria) > 0 {
        senha := q.FilaPrioritaria[0]
        q.FilaPrioritaria = q.FilaPrioritaria[1:]
        return senha, true
    }
    
    // Fila vazia
    return Senha{}, false
}
```

**Descrição:**
- Implementa a regra de atendimento com priorização.
- Prioridade: 2 normais + 1 prioritária (se houver).
- Se não houver normal, atende prioritária imediatamente.
- Retorna `(Senha{}, false)` quando fila está vazia.

**Algoritmo de Priorização:**

| Ordem | Condição | Ação |
|-------|----------|------|
| 1 | `NormalServido >= 2 && FilaPrioritaria.len > 0` | Atende prioritária, reseta contador |
| 2 | `FilaNormal.len > 0` | Atende normal, incrementa contador |
| 3 | `FilaPrioritaria.len > 0` | Atende prioritária |
| 4 | Nenhuma | Retorna falso (sem ticket) |

---

### 3.3 Cliente Totem de Senhas (`cmd/senhas/main.go`)

#### **`func main()`**

```go
func main() {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil { panic(err) }
    defer conn.Close()
    
    registrar(conn)
    reader := bufio.NewReader(conn)
    resposta, _ := reader.ReadString('\n')
    fmt.Println(strings.TrimSpace(resposta))
    menu(conn)
}
```

**Descrição:**
- Conecta ao servidor na porta 8080.
- Registra como `TS`.
- Aguarda confirmação de sucesso.
- Exibe menu.

---

#### **`func menu(conn net.Conn)`**

```go
func menu(conn net.Conn) {
    input := bufio.NewReader(os.Stdin)
    writer := bufio.NewWriter(conn)
    reader := bufio.NewReader(conn)
    
    for {
        fmt.Println("\n===== Totem =====")
        fmt.Println("1 - Senha Normal")
        fmt.Println("2 - Senha Prioritaria")
        fmt.Println("3 - Sair")
        
        opcao, _ := input.ReadString('\n')
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
        
        fmt.Println("Sua senha:", strings.TrimSpace(resposta))
    }
}
```

**Descrição:**
- Loop interativo com menu de opções.
- Opção 1: Gera senha normal (`GENERATE:N`).
- Opção 2: Gera senha prioritária (`GENERATE:P`).
- Opção 3: Encerra conexão.
- Exibe senha gerada após resposta do servidor.

**Interface de Usuário (CLI):**
```
===== Totem =====
1 - Senha Normal
2 - Senha Prioritaria
3 - Sair
> 1
Sua senha: CREATED:N1
```

---

### 3.4 Cliente Atendente (`cmd/atendente/main.go`)

#### **`func main()` e `func menu(conn net.Conn)`**

```go
func main() {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil { panic(err) }
    defer conn.Close()
    
    registrar(conn)
    reader := bufio.NewReader(conn)
    resposta, _ := reader.ReadString('\n')
    fmt.Println(strings.TrimSpace(resposta))
    menu(conn)
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
            writer.WriteString("NEXT\n")
            writer.Flush()
            
            resposta, err := reader.ReadString('\n')
            if err != nil {
                return
            }
            
            fmt.Println("Atendendo:", strings.TrimSpace(resposta))
        case "2":
            return
        default:
            fmt.Println("Opcao invalida")
        }
    }
}
```

**Descrição:**
- Menu simples com 2 opções.
- Opção 1: Solicita próxima senha (`NEXT`).
- Opção 2: Encerra.
- Exibe a senha retornada ou mensagem de fila vazia.

**Interface de Usuário (CLI):**
```
===== ATENDIMENTO =====
1 - Chamar senha
2 - Sair
> 1
Atendendo: CURRENTTICKET:N1
```

---

### 3.5 Cliente Exibição (`cmd/tv/main.go`)

#### **`func main()` e `func escutaMsg(conn net.Conn)`**

```go
func main() {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil { panic(err) }
    
    registrar(conn)
    escutaMsg(conn)
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
        logger.Info(msg)
    }
}
```

**Descrição:**
- Registra como `TV`.
- Entra em loop de escuta.
- Recebe e exibe mensagens de broadcast continuamente.
- Termina se houver erro na conexão.

**Saída Esperada (CLI):**
```
[14:23:45] DISPLAY:N1:1
[14:23:50] DISPLAY:P1:2
[14:24:15] DISPLAY:N2:1
```

---

### 3.6 Logger (`internal/logger/logger.go`)

```go
func Info(msg string) {
    fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), msg)
}
```

**Descrição:**
- Função simples de logging com timestamp (HH:MM:SS).
- Usada em todos os componentes para rastreabilidade.

---

## 4. Soluções de Software Adotadas

### 4.1 Padrão Arquitetural: Cliente-Servidor

**Descrição:** Um servidor central coordena múltiplos clientes heterogêneos.

**Vantagens:**
- Centralização de lógica complexa (fila, priorização).
- Facilita broadcast (envio para múltiplos TVs).
- Ponto único de controle.

**Desvantagens:**
- Ponto único de falha.
- Pode gerar gargalo de performance com muitos clientes.

### 4.2 Comunicação TCP com Protocolo Textual

**Descrição:** Mensagens ASCII terminadas em newline.

**Razões:**
- Simplicidade: Legível e fácil de debugar.
- Compatibilidade: Funciona em qualquer plataforma.
- Sem compressão: Overhead mínimo para este projeto.

**Formato:**
```
COMANDO:ATRIBUTO1:ATRIBUTO2\n
```

### 4.3 Goroutines para Concorrência

**Descrição:** Uma goroutine por cliente conectado.

```go
for {
    conn, _ := ouvinte.Accept()
    go handleConnection(conn, fila)  // Goroutine leve
}
```

**Benefícios:**
- Suporta milhares de conexões simultâneas.
- Modelo de programação simples.
- Escalabilidade superior ao modelo de threads.

### 4.4 Mutexes para Sincronização

**Descrição:** Protege acesso a recursos compartilhados.

```go
var clientesMutex sync.Mutex
var filaMutex sync.Mutex
var contadorMutex sync.Mutex
```

**Recursos Sincronizados:**
- `clientesMutex`: Mapa `clientes` (conexões ativas).
- `filaMutex`: Fila de senhas.
- `contadorMutex`: Contador de guichês.

**Pattern:**
```go
mutex.Lock()
// Operação crítica
mutex.Unlock()
```

### 4.5 Fila Prioritária com Regra Determinística

**Implementação:**
- Duas listas: `FilaNormal[]` e `FilaPrioritaria[]`.
- Contador: `NormalServido` para rastrear sequência.

**Regra:**
```
Se NormalServido >= 2 E existem prioridade:
    Atende prioritária, reseta NormalServido = 0
Senão se existem normais:
    Atende normal, incrementa NormalServido
Senão se existem prioritárias:
    Atende prioritária
Senão:
    Sem ticket
```

### 4.6 Map para Gerenciamento de Clientes

```go
var clientes = make(map[net.Conn]models.Client)
```

**Vantagens:**
- Lookup O(1) de cliente por conexão.
- Remoção e inserção eficientes.
- Iteração simples para broadcast.

### 4.7 Defer para Limpeza de Recursos

```go
defer func() {
    conn.Close()
    clientesMutex.Lock()
    delete(clientes, conn)
    clientesMutex.Unlock()
}()
```

**Garantia:** Recursos são liberados mesmo em caso de erro ou panic.

---

## 5. Recursos de Interface

### 5.1 Totem de Senhas (TS)

**Tipo:** Interface CLI (Command Line Interface)

**Menu:**
```
===== Totem =====
1 - Senha Normal
2 - Senha Prioritaria
3 - Sair
> [entrada do usuário]
```

**Fluxo:**
1. Conecta ao servidor.
2. Exibe confirmação de registro.
3. Mostra menu.
4. Aguarda entrada (1, 2 ou 3).
5. Envia comando ao servidor.
6. Exibe senha gerada.
7. Retorna ao menu.

**Saída Exemplo:**
```
SUCCESS
===== Totem =====
1 - Senha Normal
2 - Senha Prioritaria
3 - Sair
> 1
Sua senha: CREATED:N1
```

---

### 5.2 Terminal de Atendimento (TA)

**Tipo:** Interface CLI simples

**Menu:**
```
===== ATENDIMENTO =====
1 - Chamar senha
2 - Sair
> [entrada do usuário]
```

**Fluxo:**
1. Conecta ao servidor e recebe número de guichê.
2. Mostra confirmação.
3. Exibe menu.
4. Opção 1: Chama `NEXT`, exibe senha.
5. Opção 2: Encerra.

**Saída Exemplo:**
```
SUCCESS
===== ATENDIMENTO =====
1 - Chamar senha
2 - Sair
> 1
Atendendo: CURRENTTICKET:N1
```

---

### 5.3 Terminal de Exibição (TV)

**Tipo:** Painel de Exibição (Reader/Display)

**Comportamento:**
- Sem menu interativo.
- Recebe mensagens de broadcast continuamente.
- Exibe com timestamp.

**Saída Exemplo:**
```
[14:23:45] DISPLAY:N1:1
[14:23:47] DISPLAY:N2:2
[14:24:12] DISPLAY:P1:1
[14:25:00] DISPLAY:N3:2
```

**Formato:** `[HH:MM:SS] DISPLAY:<tipo><número>:<guiche>`

---

### 5.4 Servidor (SRV)

**Tipo:** Log de eventos do servidor

**Saída Exemplo:**
```
[14:23:40] Iniciando Server na porta :8080
[14:23:42] Nova conexao aceita
[14:23:42] Validando novo cliente...
[14:23:42] Cliente validado, tipo: TV
[14:23:45] Validando novo cliente...
[14:23:45] Cliente validado, tipo: TS
[14:23:46] Nova Senha criada: N1
[14:23:50] Validando novo cliente...
[14:23:50] Cliente validado, tipo: TA
[14:23:52] Nova Senha criada: N2
```

**Informações Registradas:**
- Inicialização do servidor.
- Novas conexões aceitas.
- Validação de clientes.
- Senhas criadas.
- Erros e falhas.

---

## 6. Protocolos de Comunicação

### 6.1 Especificação do Protocolo

**Padrão Geral:**
```
COMANDO:ATRIBUTO1:ATRIBUTO2\n
```

Todas as mensagens terminam com `\n` (newline).

---

### 6.2 Fases de Comunicação

#### **Fase 1: Registro Inicial**

**Cliente envia:**
```
REGISTER:TV\n
REGISTER:TS\n
REGISTER:TA\n
```

**Servidor responde:**
```
SUCCESS\n        (se válido)
FAIL\n          (se inválido)
```

**Lógica do Servidor:**
- Valida se o primeiro comando é `REGISTER`.
- Valida se o tipo está em `{TV, TS, TA}`.
- Se `TA`, atribui número de guichê único.
- Envia `SUCCESS` ou `FAIL`.

---

#### **Fase 2: Geração de Senhas (apenas TS)**

**Cliente `TS` envia:**
```
GENERATE:N\n    (senha normal)
GENERATE:P\n    (senha prioritária)
```

**Servidor responde:**
```
CREATED:N<número>\n
CREATED:P<número>\n
```

**Exemplos:**
```
Cliente: GENERATE:N
Servidor: CREATED:N1

Cliente: GENERATE:P
Servidor: CREATED:P1
```

---

#### **Fase 3: Chamada de Senhas (apenas TA)**

**Cliente `TA` envia:**
```
NEXT\n
```

**Servidor responde:**
```
CURRENTTICKET:N<número>\n    (se houver senha)
CURRENTTICKET:P<número>\n
NOTICKET\n                    (se fila vazia)
```

**Exemplos:**
```
Cliente: NEXT
Servidor: CURRENTTICKET:N1

Cliente: NEXT
Servidor: CURRENTTICKET:P1

Cliente: NEXT
Servidor: NOTICKET
```

---

#### **Fase 4: Broadcast para TVs**

**Servidor envia para todos `TV` conectados:**
```
DISPLAY:<tipo><número>:<guiche>\n
```

**Exemplos:**
```
DISPLAY:N1:1
DISPLAY:P2:2
DISPLAY:N3:1
```

---

### 6.3 Códigos de Erro/Status

| Código | Significado | Contexto |
|--------|-------------|----------|
| `SUCCESS` | Registro válido | Após `REGISTER` correto |
| `FAIL` | Registro inválido | Tipo inválido ou comando inicial errado |
| `DENIED` | Comando não autorizado | Ex.: `TS` enviando `NEXT`, ou `TV` enviando `GENERATE` |
| `CREATED:<tipo><num>` | Senha criada | Resposta a `GENERATE` |
| `CURRENTTICKET:<tipo><num>` | Senha chamada | Resposta a `NEXT` |
| `NOTICKET` | Fila vazia | Resposta a `NEXT` sem senhas |

---

### 6.4 Validações de Protocolo

**Ordem de Recebimento:**
1. Primeiro comando DEVE ser `REGISTER:tipo`.
2. Outros comandos após validação.

**Comandos por Tipo:**
- `TV`: Apenas recebe (sem enviar comandos após `REGISTER`).
- `TS`: Apenas `GENERATE:N` e `GENERATE:P`.
- `TA`: Apenas `NEXT`.

**Violação:** Servidor envia `DENIED`.

---

## 7. Sincronização e Concorrência

### 7.1 Problema de Condição de Corrida (Race Condition)

**Cenário Crítico:**
```
Goroutine 1 (TA) executa: Next() em Fila
Goroutine 2 (TS) executa: AdicionaNormal() em Fila
Simultaneidade → Resultado imprevisível
```

### 7.2 Solução: Mutexes

```go
var filaMutex sync.Mutex

// No servidor:
filaMutex.Lock()
senha := fila.Next()
filaMutex.Unlock()

filaMutex.Lock()
senha := fila.AdicionaNormal()
filaMutex.Unlock()
```

**Garante:** Apenas uma operação acessa a fila por vez.

### 7.3 Sincronização de Clientes

```go
var clientesMutex sync.Mutex

// Adicionar cliente:
clientesMutex.Lock()
clientes[conn] = cliente
clientesMutex.Unlock()

// Iterar clientes:
clientesMutex.Lock()
for conn, cliente := range clientes { ... }
clientesMutex.Unlock()
```

### 7.4 Sincronização de Contador

```go
var contadorTA int
var contadorMutex sync.Mutex

// Obter próximo guichê:
contadorMutex.Lock()
contadorTA++
guiche := contadorTA
contadorMutex.Unlock()
```

**Garante:** Cada `TA` recebe número único.

### 7.5 Padrão: Cópia Local Para Reduzir Lock

```go
func broadCastTVs(...) {
    clientesMutex.Lock()
    clientesTV := make(map[net.Conn]models.Client)
    for conn, cliente := range clientes {
        clientesTV[conn] = cliente
    }
    clientesMutex.Unlock()
    
    // Agora iteramos sem lock
    for conn, cliente := range clientesTV {
        // Enviar mensagem...
    }
}
```

**Benefício:** Minimiza tempo de lock, evita deadlock em envio.

### 7.6 Goroutines

**Modelo:**
```
main() → net.Listen() → Loop aceitação
                           ↓
                      go handleConnection(conn, fila)
                           ↓
                      Goroutine processando cliente
```

**Número de Goroutines:** O(n), onde n = número de clientes conectados.

**Exemplo: 100 clientes = ~100 goroutines ativas.**

---

## 8. Fluxos de Dados

### 8.1 Fluxo: Geração de Senha Normal (TS)

```
┌─────────────────────────────────────────────────────┐
│ 1. Cliente TS conecta e envia: REGISTER:TS          │
└─────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────┐
│ 2. Servidor valida e responde: SUCCESS              │
└─────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────┐
│ 3. Cliente TS envia: GENERATE:N                      │
└─────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────┐
│ 4. Servidor: Lock filaMutex                         │
│    - Incrementa ContadorNormal (1 → 2)              │
│    - Cria Senha{Tipo:"N", Numero:1}                 │
│    - Adiciona a FilaNormal[]                        │
│    - Unlock filaMutex                               │
└─────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────┐
│ 5. Servidor responde: CREATED:N1                    │
└─────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────┐
│ 6. Cliente TS recebe e exibe: Sua senha: CREATED:N1│
└─────────────────────────────────────────────────────┘
```

---

### 8.2 Fluxo: Chamada de Senha (TA)

```
┌──────────────────────────────────────────────────────┐
│ 1. Cliente TA conecta e envia: REGISTER:TA          │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│ 2. Servidor valida, atribui Guiche=1, responde SUCCESS
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│ 3. Cliente TA envia: NEXT                           │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│ 4. Servidor: Lock filaMutex                         │
│    - Chama fila.Next()                              │
│    - Validação: NormalServido=0, retorna N1         │
│    - FilaNormal = FilaNormal[1:]                    │
│    - NormalServido++                                │
│    - Unlock filaMutex                               │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│ 5. Servidor: chama broadCastTVs(N1, 1)              │
│    - Cria cópia de clientes                         │
│    - Filtra apenas TV                               │
│    - Envia DISPLAY:N1:1 para cada TV                │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│ 6. Servidor responde TA: CURRENTTICKET:N1          │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│ 7. TA recebe: Atendendo: CURRENTTICKET:N1          │
│    TV recebe: [14:25:30] DISPLAY:N1:1              │
└──────────────────────────────────────────────────────┘
```

---

### 8.3 Fluxo: Priorização (2 Normais → 1 Prioritária)

```
Fila Inicial:
  FilaNormal: [N1, N2, N3]
  FilaPrioritaria: [P1, P2]
  NormalServido: 0

NEXT #1:
  Condição: NormalServido=0, FilaNormal.len=3
  Ação: Retorna N1, NormalServido=1
  
NEXT #2:
  Condição: NormalServido=1, FilaNormal.len=2
  Ação: Retorna N2, NormalServido=2
  
NEXT #3:
  Condição: NormalServido=2, FilaPrioritaria.len=2
  Ação: Retorna P1, NormalServido=0
  
NEXT #4:
  Condição: NormalServido=0, FilaNormal.len=1
  Ação: Retorna N3, NormalServido=1

Resultado: N1 → N2 → P1 → N3 → P2 → (nenhuma)
```

---

## 9. Análise de Estruturas de Dados

### 9.1 Estrutura `Fila`

```go
type Fila struct {
    FilaNormal      []Senha  // Slice dinâmico
    FilaPrioritaria []Senha  // Slice dinâmico
    ContadorNormal      int  // Contador
    ContadorPrioritario int  // Contador
    NormalServido       int  // Contador
}
```

**Complexidade:**

| Operação | Complexidade | Razão |
|----------|--------------|-------|
| `AdicionaNormal()` | O(1)* | Append ao final |
| `AdicionaPrioritaria()` | O(1)* | Append ao final |
| `Next()` | O(1) | Remove primeiro elemento |

*Amortizado: ocasionalmente O(n) ao redimensionar, mas raramente.

**Remoção do Início:**
```go
senha := FilaNormal[0]
FilaNormal = FilaNormal[1:]  // Cria novo slice, O(n) operações
```

**Otimização Possível:** Usar fila circular ou índice de início.

### 9.2 Estrutura `Client`

```go
type Client struct {
    Tipo   string  // "TV", "TS", "TA"
    Guiche int     // Número único para TA
}
```

**Simplicidade:** Apenas 2 campos, sem nested structs.

### 9.3 Estrutura `Senha`

```go
type Senha struct {
    Tipo   string  // "N" ou "P"
    Numero int     // Sequencial
}
```

**Representação:** Completa para mensagens de protocolo.

### 9.4 Map `clientes`

```go
var clientes = make(map[net.Conn]models.Client)
```

**Propriedades:**
- **Chave:** Conexão (pointer único por conexão).
- **Valor:** Estrutura de cliente.
- **Lookup:** O(1) em média.
- **Iteração:** O(n) onde n = número de clientes.

**Uso:**
- Broadcast para TVs.
- Consulta de guichê de TA.
- Limpeza ao desconectar.

---

## 10. Considerações de Desempenho

### 10.1 Escalabilidade

**Número de Clientes Suportados:**
- Limitação: Memory + FD (file descriptors).
- Cada goroutine ≈ 2KB (Go runtime).
- Cada conexão ≈ buffers de I/O.

**Estimativa:** 10.000 a 100.000 clientes simultâneos em servidor moderno.

**Gargalo Primário:** Mutex da fila (serialização de acesso).

### 10.2 Latência

**Operação: Gerar Senha**
```
Cliente: GENERATE:N (escrita)
    ↓ ~1ms (rede)
Servidor: Processa (lock fila, adiciona)
    ↓ ~0.1ms (memória)
Servidor: Responde CREATED:N1 (escrita)
    ↓ ~1ms (rede)
Cliente: Recebe
    ↓ ~2ms total (aproximado)
```

### 10.3 Contenção de Mutex

**Cenário:** Muitos TA e TS concorrentes.

```go
// Todos tentam acessar filaMutex simultaneamente
TS: filaMutex.Lock()  ← Obtém
TA: filaMutex.Lock()  ← Espera
TS: filaMutex.Lock()  ← Espera
```

**Solução Possível:** Substituir por estrutura lock-free ou RwMutex.

### 10.4 Bandwidth

**Mensagem Típica:** ~30 bytes (ex: `DISPLAY:P10:25\n`).

**Throughput:** 1.000 operações/seg = ~30 KB/s (negligenciável).

---

## 11. Tratamento de Erros

### 11.1 Erros de Conexão

**Caso: Cliente não consegue conectar**
```go
conn, err := net.Dial("tcp", "localhost:8080")
if err != nil {
    panic(err)  // Termina cliente
}
```

**Melhoria:** Implementar retry com backoff exponencial.

### 11.2 Erros de I/O

**Caso: Leitura falha**
```go
msg, err := leitor.ReadString('\n')
if err != nil {
    logger.Info("Cliente desconectado")
    return  // Encerra goroutine
}
```

**Comportamento:** Fecha conexão, remove cliente do mapa.

### 11.3 Registro Inválido

```go
if comando != "REGISTER" {
    logger.Info("Validacao falhou")
    return
}
```

**Resposta:** Sem resposta, conexão fechada.

**Melhoria:** Enviar `FAIL` antes de fechar.

### 11.4 Comando Não Autorizado

```go
if tipoCliente != "TS" {
    logger.Info("Requisicao GENERATE negada")
    enviaMsg(escritor, "DENIED")
    continue
}
```

**Comportamento:** Envia `DENIED`, continua aguardando.

---

## 12. Conclusões

### 12.1 Avaliação Geral

O **Sistema de Atendimento Distribuído (SASE)** é uma implementação educacional sólida que demonstra:

✅ **Conceitos Implementados:**
- Comunicação TCP cliente-servidor
- Protocolo textual customizado
- Concorrência com goroutines
- Sincronização com mutexes
- Fila com priorização
- Broadcasting para múltiplos clientes

✅ **Qualidades:**
- Código limpo e modular
- Separação clara de responsabilidades (cmd/, internal/)
- Logging para rastreabilidade
- Tratamento básico de erros

⚠️ **Limitações Educacionais:**
- Sem persistência (estado perdido ao reiniciar).
- Sem autenticação ou segurança.
- Protocolo simples (sem versioning ou extensibilidade).
- Erro no registro inicial (não envia `FAIL`).

### 12.2 Possíveis Extensões

| Extensão | Impacto | Complexidade |
|----------|--------|--------------|
| Banco de dados (histórico) | Alta | Alta |
| Interface web/GUI | Média | Alta |
| Autenticação por token | Média | Média |
| Reconexão automática | Média | Média |
| Persistência de fila | Alta | Média |
| Replicação de servidor | Alta | Muito Alta |
| Compressão de protocolo | Baixa | Baixa |

### 12.3 Adequação para Fins Acadêmicos

✅ **Objetivos Atingidos:**
- Arquitetura distribuída clara.
- Demonstração prática de concorrência.
- Protocolo bem documentado.
- Código legível para aprendizado.

**Recomendação:** Projeto adequado para disciplina de **Sistemas Distribuídos**, servindo como ponto de partida para discussões sobre escalabilidade, segurança e arquitetura.

---

## Referências e Recursos

### Documentação Go

- **net:** Comunicação de rede
- **sync:** Mutexes e sincronização
- **bufio:** Leitura/escrita com buffering
- **fmt:** Formatação de strings

### Conceitos de Sistemas Distribuídos

- Padrão cliente-servidor
- Concorrência e sincronização
- Protocolo de aplicação
- Tratamento de falhas

---

**Fim do Relatório Técnico**

*Relatório preparado para fins acadêmicos.*  
*Data: 16 de junho de 2026*
