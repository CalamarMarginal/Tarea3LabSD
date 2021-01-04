package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	adminDNSpb "./adminDNSpb"
	brokerDNSpb "./brokerDNSpb"

	"google.golang.org/grpc"
)

const ipBroker string = "0.0.0.0:50058"
const ipDNS1 string = "0.0.0.0:50052" //puerto propio
const ipDNS2 string = "0.0.0.0:50053"
const ipDNS3 string = "0.0.0.0:50054"

const ipDNS1Broker string = "0.0.0.0:50055" //puerto propio
const ipDNS2Broker string = "0.0.0.0:50056"
const ipDNS3Broker string = "0.0.0.0:50057"

var auxiliar int //si el auxiliar es 1 es porque es la primera vez que se crea el archivo

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

	if req.TipoComm == "Create" {
		//create
		createDomain(req.NombreDominio, req.ParamNuevo, req.TipoComm)
	} else if req.TipoComm == "Update" {
		updateDomain(req.NombreDominio, req.TipoCambio, req.ParamNuevo, req.TipoComm)
	} else if req.TipoComm == "Delete" {
		//delete
		deleteDomain(req.NombreDominio, req.TipoComm)
	}

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

func createDomain(dominio string, ip string, comando string) {
	reloj := "1,0,0"
	aux := strings.Split(dominio, ".")
	extension := aux[1]
	extensionFinal := "." + extension
	path := "./ZFDNS1/" + extensionFinal + ".txt"
	createFile(path)
	pathLog := "./LogDNS1/" + extensionFinal + ".txt"
	createFile(pathLog)
	data := reloj + "?" + dominio + "?" + ip
	writeFile(path, comando, "ZF", data)

}

func updateDomain(dominio string, tipoCambio string, parametroNuevo string, comando string) {
	aux := strings.Split(dominio, ".")
	extension := aux[1]
	extensionFinal := "." + extension
	path := "./ZFDNS1/" + extensionFinal + ".txt"
	data := dominio + "?" + tipoCambio + "?" + parametroNuevo
	writeFile(path, comando, "ZF", data)

}

func createFile(path string) {
	var _, err = os.Stat(path)

	// revisa si el archivo existe o no
	if os.IsNotExist(err) {
		//se crea directorio si es que no existe

		var file, err = os.Create(path)
		if isError(err) {
			return
		}
		fmt.Println("Archivo creado", path)
		auxiliar = 1
		defer file.Close()
	}

}

func deleteDomain(dominio string, comando string) {
	aux := strings.Split(dominio, ".")
	extension := aux[1]
	extensionFinal := "." + extension
	path := "./ZFDNS1/" + extensionFinal + ".txt"
	data := dominio
	writeFile(path, comando, "ZF", data)
}

func writeFile(path string, comando string, archivo string, data string) {

	// Open file using READ & WRITE permission.
	var file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if isError(err) {
		return
	}
	defer file.Close()

	fmt.Println("path", path)
	fmt.Println("archivo", archivo)
	fmt.Println("data", data)

	if archivo == "ZF" {

		if comando == "Create" {

			if auxiliar == 1 {
				aux := strings.Split(data, "?")

				reloj := aux[0]
				dominio := aux[1]
				ip := aux[2]

				formato := dominio + " IN A " + ip

				_, err = fmt.Fprintln(file, reloj)
				if isError(err) {
					return
				}
				_, err = fmt.Fprintln(file, formato)
				if isError(err) {
					return
				}

				aux = strings.Split(dominio, ".")
				extension := aux[1]
				path2 := "./LogDNS1/." + extension + ".txt"
				text := "create " + dominio + " " + ip

				writeLog(path2, text)
				//cuando el archivo no existe
				auxiliar = 0

				err = file.Sync()
				if isError(err) {
					return
				}
			} else {
				aux := strings.Split(data, "?")

				// reloj := aux[0]
				dominio := aux[1]
				ip := aux[2]

				formato := dominio + " IN A " + ip
				_, err = fmt.Fprintln(file, formato)
				if isError(err) {
					return
				}

				relojAntiguo := readFileReloj(path)
				// fmt.Println("reloj", relojAntiguo)
				reloj_aux := strings.Split(relojAntiguo, ",")
				i, err := strconv.Atoi(reloj_aux[0])
				if isError(err) {
					return
				}
				i += 1
				s := strconv.Itoa(i)
				relojNuevo := s + "," + reloj_aux[1] + "," + reloj_aux[2]
				updateFile(path, relojAntiguo, relojNuevo)

				aux = strings.Split(dominio, ".")
				extension := aux[1]
				path2 := "./LogDNS1/." + extension + ".txt"
				text := "create " + dominio + " " + ip
				writeLog(path2, text)
				//cuando el archivo no existe
				auxiliar = 0

				err = file.Sync()
				if isError(err) {
					return
				}

			}

		} else if comando == "Update" {
			//--------------Reloj--------------
			relojAntiguo := readFileReloj(path)
			reloj_aux := strings.Split(relojAntiguo, ",")
			i, err := strconv.Atoi(reloj_aux[0])
			if isError(err) {
				return
			}
			i += 1
			s := strconv.Itoa(i)
			relojNuevo := s + "," + reloj_aux[1] + "," + reloj_aux[2]
			updateFile(path, relojAntiguo, relojNuevo)

			aux := strings.Split(data, "?")
			dominio := aux[0]
			tipoDeCambio := aux[1]
			valorNuevo := aux[2]

			if tipoDeCambio == "name" {
				var dominioFinalAntiguo string
				dominioAntiguo := readFile(path, dominio)
				AUX := strings.Split(dominioAntiguo, " ")
				dominioFinalAntiguo = AUX[0]
				updateFile(path, dominioFinalAntiguo, valorNuevo)

				aux = strings.Split(dominioFinalAntiguo, ".")
				extension := aux[1]
				path2 := "./LogDNS1/." + extension + ".txt"
				text := "update " + dominioFinalAntiguo + " " + valorNuevo
				writeLog(path2, text)
				err = file.Sync()
				if isError(err) {
					return
				}

			} else if tipoDeCambio == "ip" {
				fmt.Println("entre a ip")
				dominioAntiguo := readFile(path, dominio)
				AUX := strings.Split(dominioAntiguo, " ")
				ipFinalAntiguo := AUX[3]

				updateFile(path, ipFinalAntiguo, valorNuevo)

				aux = strings.Split(dominio, ".")
				extension := aux[1]
				path2 := "./LogDNS1/." + extension + ".txt"
				text := "update " + dominio + " " + valorNuevo
				writeLog(path2, text)
				err = file.Sync()
				if isError(err) {
					return
				}
			}

		} else if comando == "Delete" {

			relojAntiguo := readFileReloj(path)
			// fmt.Println("reloj", relojAntiguo)
			reloj_aux := strings.Split(relojAntiguo, ",")
			i, err := strconv.Atoi(reloj_aux[0])
			if isError(err) {
				return
			}
			i += 1
			s := strconv.Itoa(i)
			relojNuevo := s + "," + reloj_aux[1] + "," + reloj_aux[2]
			aux := readFile(path, data) //obtenemos el termino que necesitamos reemplazar por una linea en blanco
			terminos_aux := strings.Split(aux, " ")
			dominio := terminos_aux[0]
			In := terminos_aux[1]
			A := terminos_aux[2]
			Ip := terminos_aux[3]
			replace := " "

			updateFile(path, dominio, replace)
			updateFile(path, In, replace)
			updateFile(path, A, replace)
			updateFile(path, Ip, replace)
			updateFile(path, relojAntiguo, relojNuevo)

			help := strings.Split(dominio, ".")
			extension := help[1]
			path2 := "./LogDNS1/." + extension + ".txt"
			text := "delete " + dominio
			writeLog(path2, text)
			err = file.Sync()
			if isError(err) {
				return
			}
		}
	}

	// Save file changes.

	fmt.Println("File Updated Successfully.")
}

func writeLog(path string, text string) {

	var file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if isError(err) {
		return
	}

	_, err = fmt.Fprintln(file, text)
	if isError(err) {
		return
	}

	err = file.Sync()
	if isError(err) {
		return
	}
}

func updateFile(path string, terminoAntiguo string, terminoNuevo string) {
	fmt.Println("Termino antiguo", terminoAntiguo)
	fmt.Println("termino nuevo", terminoNuevo)
	input, err2 := ioutil.ReadFile(path)
	if err2 != nil {
		fmt.Println(err2)
		os.Exit(1)
	}

	output := bytes.Replace(input, []byte(terminoAntiguo), []byte(terminoNuevo), 1)

	if err2 = ioutil.WriteFile(path, output, 0666); err2 != nil {
		fmt.Println(err2)
		os.Exit(1)
	}
	return
}

func readFileReloj(path string) string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		reloj := scanner.Text()
		defer file.Close()
		return reloj
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return ""
}

func readFile(path string, termino string) string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		dominio := scanner.Text()
		res := strings.Contains(dominio, termino)
		if res == true {
			return dominio
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return ""
}

func isError(err error) bool {
	if err != nil {
		fmt.Println("--------")
		fmt.Println("Error", err.Error())
	}

	return (err != nil)
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

func ServerB() { //servidor para broker
	fmt.Println("DNS broker server is running")

	lis, err := net.Listen("tcp", ipDNS1Broker) //este puerto usa el broker para conectarse

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
