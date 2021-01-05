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

var lastDNSVisited string

var contRedirect int //anti loops recursivos

var dictDom = map[string]string{}

func opcionComando() int {
	var opcion int
	for {
		fmt.Printf("Elija una opcion: \n 1.Create \n 2.Update \n 3.Delete \n 4.Salir \n")
		fmt.Scanln(&opcion)
		if opcion != 1 && opcion != 2 && opcion != 3 && opcion != 4 {
			fmt.Printf("Debe elegir una opcion viable (1, 2, 3 o 4) .. \n")
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
			log.Printf("Error: \n%v", err)
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
			log.Printf("Error: \n%v", err)
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
			log.Printf("Error: \n%v", err)
		}
		time.Sleep(1000 * time.Millisecond)
		log.Printf("Redirigido a: %v", res.IpDNS)
		return res.IpDNS
	}
}

func sendCmdToDNS(c adminDNSpb.AdminDNSServiceClient, comandoInfo string, comm int) {
	contRedirect++
	if contRedirect > 2 {
		fmt.Println("No existe")
		contRedirect = 0
		return
	}

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
			log.Printf("Error: \n%v", err)
		}
		time.Sleep(1000 * time.Millisecond)
		log.Printf("DNS dice %v", res.Ack)
		vaux := res.Ack + "?" + lastDNSVisited
		dictDom[name] = vaux
		fmt.Println(dictDom[name])
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
			log.Printf("Error: \n%v", err)
		}
		time.Sleep(1000 * time.Millisecond)
		log.Printf("DNS dice %v", res.Ack)

		if res.Ack == "Dominio no existe" {
			_, ok := dictDom[name]
			fmt.Println("Buscando DNS actualizado")
			aux4 := strings.Split(dictDom[name], "?")
			if ok == true { //existe en el diccionario
				fmt.Println("Redirigiendo a ", aux4[1], "...")
				connectToDNS(aux4[1], comandoInfo, comm)
				return
			}
			fmt.Println("El dominio no se ha registrado")
			return

		}

		vaux := res.Ack + "?" + lastDNSVisited //reloj + dns
		dictDom[name] = vaux
		fmt.Println(dictDom[name])
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
			log.Printf("Error: \n%v", err)
		}
		time.Sleep(1000 * time.Millisecond)
		log.Printf("DNS dice %v", res.Ack)

		if res.Ack == "Dominio no existe" {
			_, ok := dictDom[comandoInfo]
			fmt.Println("Buscando DNS actualizado")
			aux4 := strings.Split(dictDom[comandoInfo], "?")
			if ok == true { //existe en el diccionario
				fmt.Println("Redirigiendo a ", aux4[1], "...")
				connectToDNS(aux4[1], comandoInfo, comm)
				return
			}
			fmt.Println("El dominio no se ha registrado")
			return

		}

	}

}

func connectToDNS(ipConnect string, comando string, tipocom int) {
	cc, err := grpc.Dial(ipConnect, grpc.WithInsecure())

	if err != nil {
		log.Printf("Failed to connect %v", err)
	} else {
		fmt.Println("Conectado al DNS:", ipConnect)
		//se ejecuta al final del ciclo de vida de la funcion
		defer cc.Close()
	}
	c := adminDNSpb.NewAdminDNSServiceClient(cc)

	lastDNSVisited = ipConnect

	sendCmdToDNS(c, comando, tipocom)
}

func main() {

	contRedirect = 0

	for {

		cc, err := grpc.Dial(ipBroker, grpc.WithInsecure())

		if err != nil {
			log.Printf("Failed to connect %v", err)
		} else {
			fmt.Println("Conectado al Broker")
			//se ejecuta al final del ciclo de vida de la funcion
			defer cc.Close()
		}
		c := adminBrokerpb.NewAdminBrokerServiceClient(cc)

		comd := opcionComando()
		if comd == 1 {
			fmt.Println("No agregue create|update|delete al ingresar el comando")
			fmt.Println("Formato Create: <nombre.dominio> <ip>")
		} else if comd == 2 {
			fmt.Println("No agregue create|update|delete al ingresar el comando")
			fmt.Println("Formato Update: <nombre.dominio> <name|ip> <parametro>")
		} else if comd == 3 {
			fmt.Println("No agregue create|update|delete al ingresar el comando")
			fmt.Println("Formato Delete: <nombre.dominio> ")
		} else {
			break
		}

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		coman := scanner.Text()

		ipRedirect := sendCmdToBroker(c, coman, comd)
		fmt.Println(ipRedirect)

		connectToDNS(ipRedirect, coman, comd)
	}

}
