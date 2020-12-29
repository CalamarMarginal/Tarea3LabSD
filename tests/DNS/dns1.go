package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	adminDNSpb "./adminDNSpb"

	"google.golang.org/grpc"
)

const ipBroker string = "0.0.0.0:50058"
const ipDNS1 string = "0.0.0.0:50052" //ip propia
const ipDNS2 string = "0.0.0.0:50053"
const ipDNS3 string = "0.0.0.0:50054"

type serverAdmin struct{}

/*
func ping(ip string) bool {
	tiempoDeEspera := time.Duration(5 * time.Second)
	_, err := net.DialTimeout("tcp", ip, tiempoDeEspera)
	if err != nil {
		return false
	}
	return true
}*/

func (*serverAdmin) AdminDNSComm(ctx context.Context, req *adminDNSpb.CommandAdminDNS) (*adminDNSpb.DnsResponse, error) {
	//da igual que comando sea, el broker solo responde con la ip de una dns
	fmt.Println("Tipo comando:", req.TipoComm)
	fmt.Println("Nombre.Dominio:", req.NombreDominio)
	fmt.Println("Tipo cambio:", req.TipoCambio)
	fmt.Println("Parametro nuevo:", req.ParamNuevo)
	ack := "escuche tu comando"
	res := &adminDNSpb.DnsResponse{
		Ack: ack,
	}
	return res, nil
}

func ServerA() { //servidor para admin
	fmt.Println("DNS admin server is running")

	lis, err := net.Listen("tcp", ipDNS1) //este puerto usa el admin para conectarse

	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	//asignar servidor de grpc a s
	s := grpc.NewServer()

	//se utiliza el archivo que se genera por el protocol buffer y utilizaremos el metodo Register y el nombre del servicio
	// le pasasomos el servidor de grpc (s) y la estructura de un servidor "server"
	adminDNSpb.RegisterAdminDNSServiceServer(s, &serverAdmin{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}

func main() {

	var wg sync.WaitGroup
	wg.Add(4)

	//server de admin y server de cliente
	go ServerA()
	//go ClientServer()
	wg.Wait()
	return
}
