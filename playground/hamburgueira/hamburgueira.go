package main

import "fmt"

func hamburgueria() {

	/* Mini teste para basico de Golang */
	var opcao int

	pedido := []string{}

pedidos:
	for {

		fmt.Println("BEM VINDO A HAMBURGUERIA")
		fmt.Println("Informe seu pedido (1 a 2), 3 para sair.")
		fmt.Println("1. Paes \n2. Carnes")
		fmt.Scan(&opcao)

		switch opcao {
		case 1:
		paes:
			for {
				fmt.Println("Escolha um pao:")
				fmt.Println("1. Pao Hamburger \n2. Pao Hotdog \n3. Voltar")
				fmt.Scan(&opcao)

				switch opcao {
				case 1:
					pedido = append(pedido, "Hamburger")
					break paes
				case 2:
					pedido = append(pedido, "Hotdog")
					break paes
				case 3:
					break paes
				default:
					fmt.Println("Favor escolha uma opcao correta.")
				}
			}
			continue
		case 2:

		carnes:
			for {
				fmt.Println("Escolha uma carne:")
				fmt.Println("1. Carne Moida \n2. Frango \n3. Voltar")
				fmt.Scan(&opcao)
				switch opcao {
				case 1:
					pedido = append(pedido, "Carne Moida")
					break carnes
				case 2:
					pedido = append(pedido, "Frango")
					break carnes
				case 3:
					break carnes
				default:
					fmt.Println("Favor escolha uma opcao correta.")
				}
			}

			continue
		case 3:
			break pedidos
		default:
			fmt.Println("Favor escolha uma opcao correta.")

		}

	}

	fmt.Println("Pedido concluido, seu pedido:", pedido)
}

func main() {

	hamburgueria()
}
