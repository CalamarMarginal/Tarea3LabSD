package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	adminBrokerpb "./adminBrokerpb"

	"google.golang.org/grpc"
)

const ipAdmin string = "0.0.0.0:50059"
const ipDNS1 string = "0.0.0.0:50052"
const ipDNS2 string = "0.0.0.0:50053"
const ipDNS3 string = "0.0.0.0:50054"

type serverA struct{}

/*
func ping(ip string) bool {
	tiempoDeEspera := time.Duration(5 * time.Second)
	_, err := net.DialTimeout("tcp", ip, tiempoDeEspera)
	if err != nil {
		return false
	}
	return true
}*/

func redirectToDNS() string {
	var dnsip string
	for {
		x1 := rand.NewSource(time.Now().UnixNano())
		y1 := rand.New(x1)
		opciondns := y1.Intn(4)
		fmt.Println("Redirigiendo a DNS", opciondns, "...")
		if opciondns == 1 { //redirige a dns1
			//if ping(ipDNS1) {
			dnsip = ipDNS1
			return dnsip
			//}
		} else if opciondns == 2 { //redirige a dns2
			//if ping(ipDNS2) {
			dnsip = ipDNS2
			return dnsip
			//}
		} else if opciondns == 3 { //redirige a dns3
			//if ping(ipDNS3) {
			dnsip = ipDNS3
			return dnsip
			//}
		} else {
			return dnsip
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
		log.Fatalf("Failed to listen %v", err)
	}

	//asignar servidor de grpc a s
	s := grpc.NewServer()

	//se utiliza el archivo que se genera por el protocol buffer y utilizaremos el metodo Register y el nombre del servicio
	// le pasasomos el servidor de grpc (s) y la estructura de un servidor "server"
	adminBrokerpb.RegisterAdminBrokerServiceServer(s, &serverA{})

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
