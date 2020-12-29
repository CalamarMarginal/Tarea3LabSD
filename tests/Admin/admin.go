package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	adminBrokerpb "./adminBrokerpb"
	adminDNSpb "./adminDNSpb"

	"google.golang.org/grpc"
)

const ipBroker string = "0.0.0.0:50059"

func opcionComando() int {
	var opcion int
	for {
		fmt.Printf("Elija una opcion: \n 1.Create \n 2.Update \n 3.Delete \n")
		fmt.Scanln(&opcion)
		if opcion != 1 && opcion != 2 && opcion != 3 {
			fmt.Printf("Debe elegir una opcion viable (1, 2 o 3) .. \n")
		} else {
			break
		}
	}
	return opcion
}

func sendCmdToBroker(c adminBrokerpb.AdminBrokerServiceClient, comandoInfo string, comm int) string {

	if comm == 1 { //create
		ncom := "Create"
		aux := strings.Split(comandoInfo, " ")
		name := aux[0]
		ip := aux[1]
		req := &adminBrokerpb.CommandAdmin{
			TipoComm:      ncom,
			NombreDominio: name,
			TipoCambio:    "-1",
			ParamNuevo:    ip,
		}
		res, err := c.AdminBrokerComm(context.Background(), req)
		if err != nil {
			log.Fatalf("Error: \n%v", err)
		}
		time.Sleep(1000 * time.Millisecond)
		log.Printf("Redirigido a: %v", res.IpDNS)
		return res.IpDNS
	} else if comm == 2 { //update
		ncom := "Update"
		aux := strings.Split(comandoInfo, " ")
		name := aux[0]
		tipo := aux[1]
		param := aux[2]
		req := &adminBrokerpb.CommandAdmin{
			TipoComm:      ncom,
			NombreDominio: name,
			TipoCambio:    tipo,
			ParamNuevo:    param,
		}
		res, err := c.AdminBrokerComm(context.Background(), req)
		if err != nil {
			log.Fatalf("Error: \n%v", err)
		}
		time.Sleep(1000 * time.Millisecond)
		log.Printf("Redirigido a: %v", res.IpDNS)
		return res.IpDNS
	} else { //delete
		ncom := "Delete"
		req := &adminBrokerpb.CommandAdmin{
			TipoComm:      ncom,
			NombreDominio: comandoInfo,
			TipoCambio:    "-1",
			ParamNuevo:    "-1",
		}
		res, err := c.AdminBrokerComm(context.Background(), req)
		if err != nil {
			log.Fatalf("Error: \n%v", err)
		}
		time.Sleep(1000 * time.Millisecond)
		log.Printf("Redirigido a: %v", res.IpDNS)
		return res.IpDNS
	}
}

func sendCmdToDNS(c adminDNSpb.AdminDNSServiceClient, comandoInfo string, comm int) {

	if comm == 1 { //create
		ncom := "Create"
		aux := strings.Split(comandoInfo, " ")
		name := aux[0]
		ip := aux[1]
		req := &adminDNSpb.CommandAdminDNS{
			TipoComm:      ncom,
			NombreDominio: name,
			TipoCambio:    "-1",
			ParamNuevo:    ip,
		}
		res, err := c.AdminDNSComm(context.Background(), req)
		if err != nil {
			log.Fatalf("Error: \n%v", err)
		}
		time.Sleep(1000 * time.Millisecond)
		log.Printf("DNS dice %v", res.Ack)

	} else if comm == 2 { //update
		ncom := "Update"
		aux := strings.Split(comandoInfo, " ")
		name := aux[0]
		tipo := aux[1]
		param := aux[2]
		req := &adminDNSpb.CommandAdminDNS{
			TipoComm:      ncom,
			NombreDominio: name,
			TipoCambio:    tipo,
			ParamNuevo:    param,
		}
		res, err := c.AdminDNSComm(context.Background(), req)
		if err != nil {
			log.Fatalf("Error: \n%v", err)
		}
		time.Sleep(1000 * time.Millisecond)
		log.Printf("DNS dice %v", res.Ack)
	} else { //delete
		ncom := "Delete"
		req := &adminDNSpb.CommandAdminDNS{
			TipoComm:      ncom,
			NombreDominio: comandoInfo,
			TipoCambio:    "-1",
			ParamNuevo:    "-1",
		}
		res, err := c.AdminDNSComm(context.Background(), req)
		if err != nil {
			log.Fatalf("Error: \n%v", err)
		}
		time.Sleep(1000 * time.Millisecond)
		log.Printf("DNS dice %v", res.Ack)
	}

}

func connectToDNS(ipConnect string, comando string, tipocom int) {
	cc, err := grpc.Dial(ipConnect, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Failed to connect %v", err)
	} else {
		fmt.Println("Conectado al DNS:", ipConnect)
		//se ejecuta al final del ciclo de vida de la funcion
		defer cc.Close()
	}
	c := adminDNSpb.NewAdminDNSServiceClient(cc)

	sendCmdToDNS(c, comando, tipocom)
}

func main() {

	cc, err := grpc.Dial(ipBroker, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Failed to connect %v", err)
	} else {
		fmt.Println("Conectado al Broker")
		//se ejecuta al final del ciclo de vida de la funcion
		defer cc.Close()
	}
	c := adminBrokerpb.NewAdminBrokerServiceClient(cc)

	comd := opcionComando()
	fmt.Println("No agregue create|update|delete al ingresar el comando")

	if comd == 1 {
		fmt.Println("Formato Create: <nombre.dominio> <ip>")
	} else if comd == 2 {
		fmt.Println("Formato Update: <nombre.dominio> <name|ip> <parametro>")
	} else if comd == 3 {
		fmt.Println("Formato Delete: <nombre.dominio> ")
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	coman := scanner.Text()

	ipRedirect := sendCmdToBroker(c, coman, comd)
	fmt.Println(ipRedirect)

	connectToDNS(ipRedirect, coman, comd)

}