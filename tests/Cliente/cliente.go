package main

import (
	"context"
	"fmt"
	"log"

	clientepb "./clienteBrokerpb"
	"google.golang.org/grpc"
)

const ipBroker string = "0.0.0.0:50058"

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

func clientBroker(c clientepb.ClienteBrokerServiceClient) {

	var dominio string

	fmt.Println("Starting client&Broker connection")

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

			ipDominio := res.GetIp()
			reloj := res.GetReloj()
			ipDNS := res.GetIpDNS()

			fmt.Println("La ip es ", ipDominio)
			fmt.Println("La reloj es ", reloj)
			fmt.Println("La ipDNS es ", ipDNS)

		} else {
			break
		}
	}

}
