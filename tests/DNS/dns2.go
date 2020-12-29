package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	adminDNSpb "./adminDNSpb"
	brokerDNSpb "./brokerDNSpb"

	"google.golang.org/grpc"
)

const ipBroker string = "0.0.0.0:50058"
const ipDNS1 string = "0.0.0.0:50052"
const ipDNS2 string = "0.0.0.0:50053" //puerto propio
const ipDNS3 string = "0.0.0.0:50054"

const ipDNS1Broker string = "0.0.0.0:50055"
const ipDNS2Broker string = "0.0.0.0:50056" //puerto propio
const ipDNS3Broker string = "0.0.0.0:50057"

type serverAdmin struct{}
type serverBroker struct{}

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

func (*serverBroker) BrokerDNSComm(ctx context.Context, req *brokerDNSpb.ClienteBrRequest) (*brokerDNSpb.DnsClientResponse, error) {
	fmt.Println("Request recibido:", req.CommCliente)
	ack := "tu pagina esta en 10.11.12.13"
	reloj := "0.0.0"
	res := &brokerDNSpb.DnsClientResponse{
		IpDominio: ack,
		Reloj:     reloj,
	}
	return res, nil
}

func ServerA() { //servidor para admin
	fmt.Println("DNS admin server is running")

	lis, err := net.Listen("tcp", ipDNS2) //este puerto usa el admin para conectarse

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

func ServerB() { //servidor para broker
	fmt.Println("DNS broker server is running")

	lis, err := net.Listen("tcp", ipDNS2Broker) //este puerto usa el broker para conectarse

	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	//asignar servidor de grpc a s
	s := grpc.NewServer()

	//se utiliza el archivo que se genera por el protocol buffer y utilizaremos el metodo Register y el nombre del servicio
	// le pasasomos el servidor de grpc (s) y la estructura de un servidor "server"
	brokerDNSpb.RegisterBrokerDNSServiceServer(s, &serverBroker{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}

func main() {

	var wg sync.WaitGroup
	wg.Add(4)

	//server de admin y server de broker
	go ServerA()
	go ServerB()
	wg.Wait()
	return
}
