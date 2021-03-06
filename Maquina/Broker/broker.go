package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	adminBrokerpb "./adminBrokerpb"
	brokerDNSpb "./brokerDNSpb"
	clientBrokerpb "./clienteBrokerpb"

	"google.golang.org/grpc"
)

const ipAdmin string = "10.10.28.70:50059"
const ipClient string = "10.10.28.70:50058"
const ipDNS1 string = "10.10.28.71:50052"
const ipDNS2 string = "10.10.28.72:50053"
const ipDNS3 string = "10.10.28.73:50054"

const ipDNS1Broker string = "10.10.28.71:50055"
const ipDNS2Broker string = "10.10.28.72:50056"
const ipDNS3Broker string = "10.10.28.73:50057"

type serverA struct{}
type serverC struct{}

/*
func ping(ip string) bool {
	tiempoDeEspera := time.Duration(5 * time.Second)
	_, err := net.DialTimeout("tcp", ip, tiempoDeEspera)
	if err != nil {
		return false
	}
	return true
}*/

func BrokerDns(c brokerDNSpb.BrokerDNSServiceClient, dominio string) string {
	fmt.Println("broker client is running")

	req := &brokerDNSpb.ClienteBrRequest{
		CommCliente: dominio,
	}

	res, err := c.BrokerDNSComm(context.Background(), req)

	if err != nil {
		log.Printf("Error calling BrokerDNSComm RPC: \n%v", err)
	}

	ipDominio := res.GetIpDominio()
	reloj := res.GetReloj()
	ipDNS := res.GetIpDNS()

	fmt.Println("la ip es ", ipDominio)
	fmt.Println("el reloj es ", reloj)
	fmt.Println("el ipDNS es ", ipDNS)

	response := ipDominio + "?" + reloj + "?" + ipDNS

	return response

}

func clientDns(dominio string) string {

	ipDns := redirectToDNS2()

	fmt.Println("client DNS is running")

	//se invoca el localhost con grpc
	//sacamos el Tlc para simplificar la conexion por certificados y seguridades
	//no conectamos al servicio en el localhost en el puerto 50052
	cc, err := grpc.Dial(ipDns, grpc.WithInsecure())

	if err != nil {
		log.Printf("Failed to connect %v", err)
	}

	//se ejecuta al final del ciclo de vida de la funcion
	defer cc.Close()

	c := brokerDNSpb.NewBrokerDNSServiceClient(cc)

	response := BrokerDns(c, dominio)

	return response

}

func (*serverC) ClienteBroker(ctx context.Context, req *clientBrokerpb.ClienteRequest) (*clientBrokerpb.BrokerResponse, error) {

	fmt.Println("Cliente Broker server is running")

	dominio := req.GetDominio()

	fmt.Println("El dominio es ", dominio)

	response := clientDns(dominio)
	split := strings.Split(response, "?")

	ip := split[0]
	reloj := split[1]
	ipDNS := split[2]

	res := &clientBrokerpb.BrokerResponse{
		Ip:    ip,
		Reloj: reloj,
		IpDNS: ipDNS,
	}
	return res, nil

}

func redirectToDNS() string {
	var dnsip string
	for {
		x1 := rand.NewSource(time.Now().UnixNano())
		y1 := rand.New(x1)
		opciondns := y1.Intn(4)
		if opciondns != 0 {

			fmt.Println("Redirigiendo a DNS", opciondns, "...")
			if opciondns == 1 { //redirige a dns1
				dnsip = ipDNS1
				return dnsip
			} else if opciondns == 2 { //redirige a dns2
				dnsip = ipDNS2
				return dnsip
			} else if opciondns == 3 { //redirige a dns3
				dnsip = ipDNS3
				return dnsip
			}
		}
	}
}
func redirectToDNS2() string {
	var dnsip string
	for {
		x1 := rand.NewSource(time.Now().UnixNano())
		y1 := rand.New(x1)
		opciondns := y1.Intn(4)
		if opciondns != 0 {

			fmt.Println("Redirigiendo a DNS", opciondns, "...")
			if opciondns == 1 { //redirige a dns1
				dnsip = ipDNS1Broker
				return dnsip
			} else if opciondns == 2 { //redirige a dns2
				dnsip = ipDNS2Broker
				return dnsip
			} else if opciondns == 3 { //redirige a dns3
				dnsip = ipDNS3Broker
				return dnsip
			}
		}
	}
}

func (*serverA) AdminBrokerComm(ctx context.Context, req *adminBrokerpb.CommandAdmin) (*adminBrokerpb.RedirectDNS, error) {
	//da igual que comando sea, el broker solo responde con la ip de una dns
	ack := redirectToDNS()
	res := &adminBrokerpb.RedirectDNS{
		IpDNS: ack,
	}
	return res, nil
}

func ServerA() { //servidor para admin
	fmt.Println("Broker admin server is running")

	lis, err := net.Listen("tcp", ipAdmin)

	if err != nil {
		log.Printf("Failed to listen %v", err)
	}

	//asignar servidor de grpc a s
	s := grpc.NewServer()

	//se utiliza el archivo que se genera por el protocol buffer y utilizaremos el metodo Register y el nombre del servicio
	// le pasasomos el servidor de grpc (s) y la estructura de un servidor "server"
	adminBrokerpb.RegisterAdminBrokerServiceServer(s, &serverA{})

	if err := s.Serve(lis); err != nil {
		log.Printf("Failed to serve %v", err)
	}
}

func ServerC() {
	fmt.Println("Broker client server is running")

	lis, err := net.Listen("tcp", ipClient)

	if err != nil {
		log.Printf("Failed to listen %v", err)
	}

	//asignar servidor de grpc a s
	s := grpc.NewServer()

	//se utiliza el archivo que se genera por el protocol buffer y utilizaremos el metodo Register y el nombre del servicio
	// le pasasomos el servidor de grpc (s) y la estructura de un servidor "server"
	clientBrokerpb.RegisterClienteBrokerServiceServer(s, &serverC{})

	if err := s.Serve(lis); err != nil {
		log.Printf("Failed to serve %v", err)
	}
}

func main() {

	var wg sync.WaitGroup
	wg.Add(2)

	//server de admin y server de cliente
	go ServerA()
	go ServerC()
	wg.Wait()
	return
}
