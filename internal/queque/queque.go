package queque

/* Fila prioritaria para registro de senhas P e N */

type Fila struct {

	// Lista para armazenar senhas P e N
	FilaNormal      []Senha
	FilaPrioritaria []Senha

	// Contador para ir progredindo conforme novas senhas criadas
	ContadorNormal      int
	ContadorPrioritario int

	// Para contar a cada 2 senhas N
	NormalServido int
}

func (q *Fila) AdicionaNormal() Senha {

	q.ContadorNormal++

	// Cria nova senha
	senha := Senha{
		Tipo:   "N",
		Numero: q.ContadorNormal,
	}

	// Add senha a fila
	q.FilaNormal = append(q.FilaNormal, senha)

	return senha

}

func (q *Fila) AdicionaPrioritaria() Senha {

	q.ContadorPrioritario++

	senha := Senha{
		Tipo:   "P",
		Numero: q.ContadorPrioritario,
	}

	q.FilaPrioritaria = append(q.FilaPrioritaria, senha)

	return senha

}

func (q *Fila) Next() (Senha, bool) {

	// Se ja se passaram 2 senhas normais
	if q.NormalServido >= 2 && len(q.FilaPrioritaria) > 0 {

		// Pega senha no inicio da lista
		senha := q.FilaPrioritaria[0]

		// Remove senha do inicio
		q.FilaPrioritaria = q.FilaPrioritaria[1:]

		// Reinicia contador
		q.NormalServido = 0

		return senha, true
	}

	// Chama senha normal
	if len(q.FilaNormal) > 0 {

		senha := q.FilaNormal[0]

		q.FilaNormal = q.FilaNormal[1:]

		// Adiciona 1 para normal servido
		q.NormalServido++

		return senha, true
	}

	// Caso nao tenha normal, mas tenha prioritario
	if len(q.FilaPrioritaria) > 0 {

		senha := q.FilaPrioritaria[0]

		q.FilaPrioritaria = q.FilaPrioritaria[1:]

		return senha, true
	}

	// Se tiver nada na fila
	return Senha{}, false

}
