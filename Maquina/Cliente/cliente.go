package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	clientepb "./clienteBrokerpb"
	"google.golang.org/grpc"
)

const ipBroker string = "10.10.28.70:50058"

var dictConsulta = map[string]string{}

func checkDic() {
	fmt.Println("Valores en diccionario:")
	for k, v := range dictConsulta {
		fmt.Printf("%s -> %s\n", k, v)
	}
}

func main() {
	fmt.Println("Go client is running")

	//se invoca el localhost con grpc
	//sacamos el Tlc para simplificar la conexion por certificados y seguridades
	//no conectamos al servicio en el localhost en el puerto 50052
	cc, err := grpc.Dial(ipBroker, grpc.WithInsecure())

	if err != nil {
		log.Printf("Failed to connect %v", err)
	}

	//se ejecuta al final del ciclo de vida de la funcion
	defer cc.Close()

	c := clientepb.NewClienteBrokerServiceClient(cc)
	clientBroker(c)
}

func opcionViable() int {

	var opcion2 int
	for {
		fmt.Printf("Eliga una opcion: \n 1. Ingresar nombre del sitio a buscar\n 2. Salir \n")
		fmt.Scanln(&opcion2)

		if opcion2 != 1 && opcion2 != 2 {
			fmt.Printf("Debe elegir una opcion viable (1 o 2) .. \n")
		} else {
			break
		}
	}
	return opcion2
}

func compareReloj(reloj1 string, reloj2 string) bool { //formato "x,y,z"
	//funcion revisa si reloj consultado es el mas reciente
	splR1 := strings.Split(reloj1, ",")
	splR2 := strings.Split(reloj2, ",")

	for n := range splR1 {
		i1, err := strconv.Atoi(splR1[n])
		if err != nil {
			fmt.Println(err)
		}
		i2, err2 := strconv.Atoi(splR2[n])
		if err2 != nil {
			fmt.Println(err)
		}
		if i2 > i1 { //valor reloj es mayor en el visitado
			continue
		} else {
			return false
		}
	}

	return true
}

func clientBroker(c clientepb.ClienteBrokerServiceClient) {

	var dominio string

	fmt.Println("Starting Client & Broker connection")

	for {
		var opcion = opcionViable()

		if opcion == 1 {

			fmt.Printf("Ingrese el dominio: ")
			fmt.Scanln(&dominio)

			req := &clientepb.ClienteRequest{
				Dominio: dominio,
			}

			res, err := c.ClienteBroker(context.Background(), req)

			if err != nil {
				log.Printf("Error calling ClientBroker RPC: \n%v", err)
			}
			auxDir := strings.Split(res.GetIp(), " ")
			x := len(auxDir)
			if x != 0 || res.GetIp() != "" { //respuesta no vacia
				if res.GetIp() != "No encontrado" {
					ipDominio := auxDir[x-1]
					reloj := res.GetReloj()
					ipDNS := res.GetIpDNS()

					_, ok := dictConsulta[dominio]
					if ok == true {
						maux := strings.Split(dictConsulta[dominio], "?")
						relojAnt := maux[0]
						if compareReloj(relojAnt, reloj) == false { //nuestra version es mas reciente o concurrente
							fmt.Println("La ip del dominio es ", maux[2]) //guardamos ip en memoria para no hacer otra consulta
							fmt.Println("El reloj es ", relojAnt)
							fmt.Println("La DNS que respondio es ", maux[1])
						} else { //respuesta es mas reciente
							fmt.Println("La ip del dominio es ", ipDominio)
							fmt.Println("El reloj es ", reloj)
							fmt.Println("La DNS que respondio es ", ipDNS)
							dictConsulta[dominio] = reloj + "?" + ipDNS + "?" + ipDominio
						}
					} else {
						fmt.Println("La ip del dominio es ", ipDominio)
						fmt.Println("El reloj es ", reloj)
						fmt.Println("La DNS que respondio es ", ipDNS)
						dictConsulta[dominio] = reloj + "?" + ipDNS + "?" + ipDominio
					}

				} else {
					fmt.Println("No se ha encontrado el dominio")
				}
			} else {
				_, ok2 := dictConsulta[dominio]
				if ok2 == true {
					waux := strings.Split(dictConsulta[dominio], "?")
					fmt.Println("La ip del dominio es ", waux[2])
					fmt.Println("El reloj es ", waux[0])
					fmt.Println("La DNS que respondio es ", waux[1])
				} else {
					fmt.Println("No se ha encontrado el dominio")
				}
			}
			checkDic()
		} else {
			break
		}
	}

}
