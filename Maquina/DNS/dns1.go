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
	"time"

	adminDNSpb "./adminDNSpb"
	brokerDNSpb "./brokerDNSpb"
	clientDNSpb "./clientDNSpb"

	"google.golang.org/grpc"
)

const ipBroker string = "10.10.28.70:50058"
const ipDNS1 string = "10.10.28.71:50052" //puerto propio
const ipDNS2 string = "10.10.28.72:50053"
const ipDNS3 string = "10.10.28.73:50054"

const ipDNS1Broker string = "10.10.28.71:50055" //puerto propio
const ipDNS2Broker string = "10.10.28.72:50056"
const ipDNS3Broker string = "10.10.28.73:50057"

const ipDNS1DNS2 string = "10.10.28.72:50050"
const ipDNS1DNS3 string = "10.10.28.73:50051"

var nuevoReloj string

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
	reloj := "[]"
	/*
		fmt.Println("Tipo comando:", req.TipoComm)
		fmt.Println("Nombre.Dominio:", req.NombreDominio)
		fmt.Println("Tipo cambio:", req.TipoCambio)
		fmt.Println("Parametro nuevo:", req.ParamNuevo)
	*/
	if req.TipoComm == "Create" {
		//create
		reloj = createDomain(req.NombreDominio, req.ParamNuevo, req.TipoComm)
	} else if req.TipoComm == "Update" {
		reloj = updateDomain(req.NombreDominio, req.TipoCambio, req.ParamNuevo, req.TipoComm)
	} else if req.TipoComm == "Delete" {
		//delete
		reloj = deleteDomain(req.NombreDominio, req.TipoComm)
	}

	res := &adminDNSpb.DnsResponse{
		Ack: reloj,
	}
	return res, nil
}

func (*serverBroker) BrokerDNSComm(ctx context.Context, req *brokerDNSpb.ClienteBrRequest) (*brokerDNSpb.DnsClientResponse, error) {
	fmt.Println("Request recibido:", req.CommCliente) //recibe nombre.dom
	varAux := strings.Split(req.CommCliente, ".")
	path := "./ZFDNS1/." + varAux[1] + ".txt"
	aux := readFile(path, req.CommCliente)
	//terminoAux := strings.Split(aux, " ") //recibe ej "algo.com 3.4.5.3"
	//ipDom := terminoAux[1]
	reloj := readFileReloj(path)
	ipDNSpropia := ipDNS1Broker
	res := &brokerDNSpb.DnsClientResponse{
		IpDominio: aux,
		Reloj:     reloj,
		IpDNS:     ipDNSpropia,
	}
	return res, nil
}

func createDomain(dominio string, ip string, comando string) string {
	reloj := "1,0,0"
	aux := strings.Split(dominio, ".")
	extension := aux[1]
	extensionFinal := "." + extension
	path := "./ZFDNS1/" + extensionFinal + ".txt"
	createFile(path)
	pathLog := "./LogDNS1/" + extensionFinal + ".txt"
	createFile(pathLog)
	data := reloj + "?" + dominio + "?" + ip
	clock := writeFile(path, comando, "ZF", data)
	return clock
}

func updateDomain(dominio string, tipoCambio string, parametroNuevo string, comando string) string {
	aux := strings.Split(dominio, ".")
	extension := aux[1]
	extensionFinal := "." + extension
	path := "./ZFDNS1/" + extensionFinal + ".txt"
	data := dominio + "?" + tipoCambio + "?" + parametroNuevo
	clock := writeFile(path, comando, "ZF", data)
	return clock

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

func deleteDomain(dominio string, comando string) string {
	aux := strings.Split(dominio, ".")
	extension := aux[1]
	extensionFinal := "." + extension
	path := "./ZFDNS1/" + extensionFinal + ".txt"
	data := dominio
	clock := writeFile(path, comando, "ZF", data)
	return clock
}

func writeFile(path string, comando string, archivo string, data string) string {

	// Open file using READ & WRITE permission.
	clock := "[]" //valor dummy
	var file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if isError(err) {
		return "Dominio no existe"
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
					return ""
				}
				_, err = fmt.Fprintln(file, formato)
				if isError(err) {
					return ""
				}

				aux = strings.Split(dominio, ".")
				extension := aux[1]
				path2 := "./LogDNS1/." + extension + ".txt"
				text := "create " + dominio + " " + ip

				writeLog(path2, text)
				//cuando el archivo no existe
				auxiliar = 0

				clock = reloj

				err = file.Sync()
				if isError(err) {
					return ""
				}
			} else {
				aux := strings.Split(data, "?")

				// reloj := aux[0]
				dominio := aux[1]
				ip := aux[2]

				formato := dominio + " IN A " + ip
				_, err = fmt.Fprintln(file, formato)
				if isError(err) {
					return ""
				}

				relojAntiguo := readFileReloj(path)
				// fmt.Println("reloj", relojAntiguo)
				relojAux := strings.Split(relojAntiguo, ",")
				i, err := strconv.Atoi(relojAux[0])
				if isError(err) {
					return ""
				}
				i++
				s := strconv.Itoa(i)
				relojNuevo := s + "," + relojAux[1] + "," + relojAux[2]
				updateFile(path, relojAntiguo, relojNuevo)

				clock = relojNuevo

				aux = strings.Split(dominio, ".")
				extension := aux[1]
				path2 := "./LogDNS1/." + extension + ".txt"
				text := "create " + dominio + " " + ip
				writeLog(path2, text)
				//cuando el archivo no existe
				auxiliar = 0

				err = file.Sync()
				if isError(err) {
					return ""
				}

			}

		} else if comando == "Update" {

			aux := strings.Split(data, "?")
			dominio := aux[0]
			tipoDeCambio := aux[1]
			valorNuevo := aux[2]

			if tipoDeCambio == "name" {
				var dominioFinalAntiguo string
				dominioAntiguo := readFile(path, dominio)
				AUX := strings.Split(dominioAntiguo, " ")
				dominioFinalAntiguo = AUX[0]
				if dominioFinalAntiguo == "" {
					return "Dominio no existe"
				}
				updateFile(path, dominioFinalAntiguo, valorNuevo)

				aux = strings.Split(dominioFinalAntiguo, ".")
				extension := aux[1]
				path2 := "./LogDNS1/." + extension + ".txt"
				text := "update " + dominioFinalAntiguo + " " + valorNuevo
				writeLog(path2, text)
				err = file.Sync()
				if isError(err) {
					return "Dominio no existe"
				}

			} else if tipoDeCambio == "ip" {
				//fmt.Println("entre a ip")
				dominioAntiguo := readFile(path, dominio)
				if dominioAntiguo == "" {
					return "Dominio no existe"
				}
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
					return "Dominio no existe"
				}
			}
			//--------------Reloj--------------
			relojAntiguo := readFileReloj(path)
			relojAux := strings.Split(relojAntiguo, ",")
			i, err := strconv.Atoi(relojAux[0])
			if isError(err) {
				return "Dominio no existe"
			}
			i++
			s := strconv.Itoa(i)
			relojNuevo := s + "," + relojAux[1] + "," + relojAux[2]
			updateFile(path, relojAntiguo, relojNuevo)

			clock = relojNuevo

		} else if comando == "Delete" {

			aux := readFile(path, data) //obtenemos el termino que necesitamos reemplazar por una linea en blanco
			terminosAux := strings.Split(aux, " ")
			dominio := terminosAux[0]
			if dominio == "" {
				return "Dominio no existe"
			}
			relojAntiguo := readFileReloj(path)
			// fmt.Println("reloj", relojAntiguo)
			relojAux := strings.Split(relojAntiguo, ",")
			i, err := strconv.Atoi(relojAux[0])
			if isError(err) {
				return "Dominio no existe"
			}
			i++
			s := strconv.Itoa(i)
			relojNuevo := s + "," + relojAux[1] + "," + relojAux[2]

			deleteLine(path, dominio)
			updateFile(path, relojAntiguo, relojNuevo)

			clock = relojNuevo

			help := strings.Split(dominio, ".")
			extension := help[1]
			path2 := "./LogDNS1/." + extension + ".txt"
			text := "delete " + dominio
			writeLog(path2, text)
			err = file.Sync()
			if isError(err) {
				return "Dominio no existe"
			}
		}
	}

	// Save file changes.

	fmt.Println("File Updated Successfully.")
	return clock
}

func deleteLine(ruta string, name string) {
	input, err := ioutil.ReadFile(ruta)
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, name) {
			lines[i] = " "
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(ruta, []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}
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
		fmt.Println("Error", err.Error())
		return "No encontrado"
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
		fmt.Println("Error", err.Error())
		return "No encontrado"
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
		log.Printf("Failed to listen %v", err)
	}

	//asignar servidor de grpc a s
	s := grpc.NewServer()

	//se utiliza el archivo que se genera por el protocol buffer y utilizaremos el metodo Register y el nombre del servicio
	// le pasasomos el servidor de grpc (s) y la estructura de un servidor "server"
	adminDNSpb.RegisterAdminDNSServiceServer(s, &serverAdmin{})

	if err := s.Serve(lis); err != nil {
		log.Printf("Failed to serve %v", err)
	}
}

func ServerB() { //servidor para broker
	fmt.Println("DNS broker server is running")

	lis, err := net.Listen("tcp", ipDNS1Broker) //este puerto usa el broker para conectarse

	if err != nil {
		log.Printf("Failed to listen %v", err)
	}

	//asignar servidor de grpc a s
	s := grpc.NewServer()

	//se utiliza el archivo que se genera por el protocol buffer y utilizaremos el metodo Register y el nombre del servicio
	// le pasasomos el servidor de grpc (s) y la estructura de un servidor "server"
	brokerDNSpb.RegisterBrokerDNSServiceServer(s, &serverBroker{})

	if err := s.Serve(lis); err != nil {
		log.Printf("Failed to serve %v", err)
	}
}

func clientDNS2(wg *sync.WaitGroup) {
	fmt.Println("Go clientDNS2 is running")

	//se invoca el localhost con grpc
	//sacamos el Tlc para simplificar la conexion por certificados y seguridades
	//no conectamos al servicio en el localhost en el puerto 50052
	cc, err := grpc.Dial(ipDNS1DNS2, grpc.WithInsecure())

	if err != nil {
		log.Printf("Failed to connect")
	}

	//se ejecuta al final del ciclo de vida de la funcion
	defer cc.Close()

	c := clientDNSpb.NewClientDNSServiceClient(cc)
	clientDNS1DNS2(c)

	defer wg.Done()

	return
}

func clientDNS1DNS2(c clientDNSpb.ClientDNSServiceClient) {
	//se crea un request basada en una estructura del protocol buffer
	req := &clientDNSpb.ClienteDNSRequest{
		TimeComplete: "timerListo",
	}

	res, err := c.ClientDNS(context.Background(), req)

	if err != nil {
		log.Printf("Error, calling DNS2: \n")
	}

	log := res.GetLog()
	// fmt.Println("LOooooooooooooooooooooooooooooooog", log)
	reloj := res.GetReloj()

	// fmt.Println("DNS 2 --> log: : ", log)
	// fmt.Println("DNS 2 --> reloj: : ", reloj)

	auxDominio := strings.Split(log, "?")

	for i, nombreDominio := range auxDominio {
		if len(nombreDominio) > 0 {
			auxNombreDominio := strings.Split(nombreDominio, " ")
			auxReloj := strings.Split(reloj, "?")
			nombreDominio := auxNombreDominio[0]
			// fmt.Println("nombreDominio", nombreDominio)
			// fmt.Println("reloj", auxReloj[i])
			comprobacionRelojes("./ZFDNS1", log, nombreDominio, auxReloj[i])
		}
	}
	// recorrerDirectorioRelojNuevo("./ZFDNS1")

	return

}

func clientDNS2confirmation(wg *sync.WaitGroup) {
	fmt.Println("Go clientDNS3 is running")

	//se invoca el localhost con grpc
	//sacamos el Tlc para simplificar la conexion por certificados y seguridades
	//no conectamos al servicio en el localhost en el puerto 50052
	cc, err := grpc.Dial(ipDNS1DNS2, grpc.WithInsecure())

	if err != nil {
		log.Printf("Failed to connect")
	}

	c := clientDNSpb.NewClientDNSServiceClient(cc)
	clientDNS1DNS2Confirmation(c)
	defer wg.Done()

	//se ejecuta al final del ciclo de vida de la funcion
	defer cc.Close()
	return
}

func clientDNS1DNS2Confirmation(c clientDNSpb.ClientDNSServiceClient) {

	dataLog := recorrerDirectorio("./LogDNS1")
	dataZf := recorrerDirectorio("./ZFDNS1")

	req := &clientDNSpb.ClientDNSRequestConfirmation{
		Log: dataLog,
		Zf:  dataZf,
	}

	res, err := c.ClientDNSConfirmation(context.Background(), req)

	if err != nil {
		log.Printf("Error calling DNS2 : \n")
	}

	log.Printf("DNS2 responde: %v", res)

	return
}

func clientDNS3(wg *sync.WaitGroup) {
	fmt.Println("Go clientDNS3 is running")

	//se invoca el localhost con grpc
	//sacamos el Tlc para simplificar la conexion por certificados y seguridades
	//no conectamos al servicio en el localhost en el puerto 50052
	cc, err := grpc.Dial(ipDNS1DNS3, grpc.WithInsecure())

	if err != nil {
		log.Printf("Failed to connect")
	}

	//se ejecuta al final del ciclo de vida de la funcion
	defer cc.Close()

	c := clientDNSpb.NewClientDNSServiceClient(cc)
	clientDNS1DNS3(c)

	defer wg.Done()

	return
}

func clientDNS1DNS3(c clientDNSpb.ClientDNSServiceClient) {
	//se crea un request basada en una estructura del protocol buffer
	req := &clientDNSpb.ClienteDNSRequest{
		TimeComplete: "timerListo",
	}

	res, err := c.ClientDNS(context.Background(), req)

	if err != nil {
		log.Printf("Error, calling DNS3: \n")
	}

	log := res.GetLog()
	reloj := res.GetReloj()

	// fmt.Println("DNS 2 --> log: : ", log)
	// fmt.Println("DNS 2 --> reloj: : ", reloj)

	auxDominio := strings.Split(log, "?")

	for i, nombreDominio := range auxDominio {
		if len(nombreDominio) > 0 {
			auxNombreDominio := strings.Split(nombreDominio, " ")
			auxReloj := strings.Split(reloj, "?")
			nombreDominio := auxNombreDominio[0]
			// fmt.Println("nombreDominio", nombreDominio)
			// fmt.Println("reloj", auxReloj[i])
			comprobacionRelojes("./ZFDNS1", log, nombreDominio, auxReloj[i])
		}
	}
	// recorrerDirectorioRelojNuevo("./ZFDNS1")

	return

}

func clientDNS3confirmation(wg *sync.WaitGroup) {
	fmt.Println("Go clientDNS3 is running")

	//se invoca el localhost con grpc
	//sacamos el Tlc para simplificar la conexion por certificados y seguridades
	//no conectamos al servicio en el localhost en el puerto 50052
	cc, err := grpc.Dial(ipDNS1DNS3, grpc.WithInsecure())

	if err != nil {
		log.Printf("Failed to connect")
	}

	c := clientDNSpb.NewClientDNSServiceClient(cc)
	clientDNS1DNS3Confirmation(c)
	defer wg.Done()

	//se ejecuta al final del ciclo de vida de la funcion
	defer cc.Close()
	return
}

func clientDNS1DNS3Confirmation(c clientDNSpb.ClientDNSServiceClient) {
	dataLog := recorrerDirectorio("./LogDNS1")
	dataZf := recorrerDirectorio("./ZFDNS1")

	req := &clientDNSpb.ClientDNSRequestConfirmation{
		Log: dataLog,
		Zf:  dataZf,
	}

	res, err := c.ClientDNSConfirmation(context.Background(), req)

	if err != nil {
		log.Printf("Error calling DNS2 : \n")
	}

	log.Printf("DNS3 responde: %v", res)

	return
}

func comprobacionRelojes(folder string, log string, nombreDominio string, reloj string) {
	archivos, err := ioutil.ReadDir(folder)
	if err != nil {
		fmt.Println(err)
	}

	flag := 0
	posicion := 0

	// fmt.Println("nombreDominio", nombreDominio)

	for j, archivo := range archivos {
		fmt.Println("archivo", archivo.Name())
		fmt.Println("nombreDominio", nombreDominio)
		posicion = j // para los dns que no existen
		fmt.Println("posicion", posicion)
		if archivo.Name() == nombreDominio {
			flag = 1

			path := folder + "/" + archivo.Name()
			input, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Println(err)
			}

			lines := strings.Split(string(input), "\n")

			if reloj != lines[0] {
				auxSplitDNSpropio := strings.Split(lines[0], ",")
				auxSplitDNSexterno := strings.Split(reloj, ",")

				for i, valor := range auxSplitDNSpropio {

					valorDNSpropia, err2 := strconv.Atoi(valor)
					if err2 != nil {
						fmt.Println(err)
					}
					valorDNSexterna, err3 := strconv.Atoi(auxSplitDNSexterno[i])
					if err3 != nil {
						fmt.Println(err)
					}
					if valor != auxSplitDNSexterno[i] {

						if valorDNSpropia < valorDNSexterna {
							merge(j, nombreDominio, log)
							nuevoReloj += auxSplitDNSexterno[i]
							nuevoReloj += ","
						} else {
							nuevoReloj += valor
							nuevoReloj += ","
						}

					} else {
						nuevoReloj += valor
						nuevoReloj += ","
					}

				}
				fmt.Println("nuevo Reloj es ", nuevoReloj)

				fixRelojes(nombreDominio, nuevoReloj)
				nuevoReloj = ""

			}

		}

	}

	if flag == 0 {
		fmt.Println("reloj !!!", reloj)
		fmt.Println("no existe en este dns", nombreDominio)
		merge(posicion, nombreDominio, log)
		fixRelojes(nombreDominio, reloj)

	}

}

func fixRelojes(nombreDominio string, nuevoReloj string) {
	fmt.Println("--------------------------Nombre Dominio-----------------", nombreDominio)
	fmt.Println("--------------------------nuevoReloj -----------------", nuevoReloj)

	relojAntiguo := ""

	i := nuevoReloj
	aux_zf := strings.Split(i, ",")
	fmt.Println(aux_zf)
	fixReloj := ""
	for i, x := range aux_zf {
		if i < 2 {
			fmt.Println(x)
			fixReloj += x
			fixReloj += ","

		} else {
			fixReloj += x
		}
	}

	fmt.Println(fixReloj)

	archivos, err := ioutil.ReadDir("./ZFDNS1")
	if err != nil {
		log.Fatal(err)
	}

	for _, archivo := range archivos {

		fmt.Println("Nombre:", archivo.Name())
		// fmt.Println("Tamaño:", archivo.Size())
		// fmt.Println("Modo:", archivo.Mode())
		// fmt.Println("Ultima modificación:", archivo.ModTime())
		// fmt.Println("Es directorio?:", archivo.IsDir())

		if archivo.Name() == nombreDominio {

			path := "./ZFDNS1/" + archivo.Name()
			fmt.Println("path es :", path)
			input, err := ioutil.ReadFile(path)
			if err != nil {
				log.Fatalln(err)
			}
			lines := strings.Split(string(input), "\n")

			for _, line := range lines {
				if strings.Contains(line, ",") {
					relojAntiguo = line
					fmt.Println("RELOJ ANTIGUO ES", relojAntiguo)
					break
				}

			}
			input, err2 := ioutil.ReadFile(path)
			if err2 != nil {
				fmt.Println(err2)
				os.Exit(1)
			}

			output := bytes.Replace(input, []byte(relojAntiguo), []byte(nuevoReloj), 1)

			if err2 = ioutil.WriteFile(path, output, 0666); err2 != nil {
				fmt.Println(err2)
				os.Exit(1)
			}

		}

	}

}

func merge(posicion int, nombreDominio string, log string) {

	i := 0
	j := 0
	k := 0
	cmd := ""

	aux := strings.Split(log, "?")
	fmt.Println("aux", aux)

	valorAiterar := aux[posicion]

	// fmt.Println("valor a iterar", valorAiterar)

	valorEspecifico := strings.Split(valorAiterar, " ")

	for _, term := range valorEspecifico {
		if term == "create" {
			cmd = cmd + term
			i++
			continue
		}
		if i > 0 {
			cmd = cmd + " " + term
			i++
		}
		if i > 2 {
			//aca esta el comando
			aux := strings.Split(cmd, " ")
			dominio := aux[1]
			ip := aux[2]
			fmt.Println("dominio es: ", dominio)
			fmt.Println("ip es: ", ip)
			i = 0
			cmd = ""

			clock := createDomain(dominio, ip, "Create")
			fmt.Println("clock: ", clock)

		}
		if term == "update" {
			cmd = cmd + term
			k++
			continue
		}
		if k > 0 {
			cmd = cmd + " " + term
			k++
		}
		if k > 2 {
			//aca esta el comando
			extension := strings.Split(nombreDominio, ".")
			if strings.HasSuffix(cmd, extension[1]) {
				aux := strings.Split(cmd, " ")
				dominioAntiguo := aux[1]
				dominioNuevo := aux[2]
				fmt.Println("dominioAntiguo es: ", dominioAntiguo)
				fmt.Println("dominioNuevo es: ", dominioNuevo)
				clock := updateDomain(dominioAntiguo, "name", dominioNuevo, "Update")
				fmt.Println("clock: ", clock)

			} else {
				aux := strings.Split(cmd, " ")
				dominioAntiguo := aux[1]
				ip := aux[2]
				fmt.Println("dominioAntiguo es: ", dominioAntiguo)
				fmt.Println("ip es: ", ip)
				clock := updateDomain(dominioAntiguo, "ip", ip, "Update")
				fmt.Println("clock: ", clock)

			}

			k = 0
			cmd = ""
		}

		if term == "delete" {
			cmd = cmd + term
			j++
			continue
		}
		if j > 0 {
			cmd = cmd + " " + term
			j++
		}
		if j > 1 {
			aux := strings.Split(cmd, " ")
			dominio := aux[1] //aca esta el comando
			fmt.Println("dominio es: ", dominio)
			j = 0
			cmd = ""
			clock := deleteDomain(dominio, "Delete")
			fmt.Println("clock: ", clock)

		}
	}

}

func recorrerDirectorioRelojNuevo(folder string) {
	archivos, err := ioutil.ReadDir(folder)
	if err != nil {
		log.Fatal(err)
	}

	data := ""

	relojes := strings.Split(nuevoReloj, ",")

	fmt.Println("relojes: ", relojes)

	// relojes_cantidad := int(len(relojes) / 3)

	for _, archivo := range archivos {
		data += archivo.Name()
		data += " "
		fmt.Println("Nombre:", archivo.Name())
		// fmt.Println("Tamaño:", archivo.Size())
		// fmt.Println("Modo:", archivo.Mode())
		// fmt.Println("Ultima modificación:", archivo.ModTime())
		// fmt.Println("Es directorio?:", archivo.IsDir())
		path := folder + "/" + archivo.Name()
		input, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalln(err)
		}

		lines := strings.Split(string(input), "\n")

		for i, line := range lines {
			if strings.Contains(line, ",") {
				// lines[i] = relojes[i]
				fmt.Println("relojes[i]", relojes[i])
			}
		}

		data += "?"

	}
	fmt.Println("data", data)

}

func recorrerDirectorio(folder string) string {
	archivos, err := ioutil.ReadDir(folder)
	if err != nil {
		log.Fatal(err)
	}

	data := ""

	for _, archivo := range archivos {
		data += archivo.Name()
		data += " "
		fmt.Println("Nombre:", archivo.Name())
		// fmt.Println("Tamaño:", archivo.Size())
		// fmt.Println("Modo:", archivo.Mode())
		// fmt.Println("Ultima modificación:", archivo.ModTime())
		// fmt.Println("Es directorio?:", archivo.IsDir())
		path := folder + "/" + archivo.Name()
		input, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalln(err)
		}
		lines := strings.Split(string(input), "\n")
		for _, line := range lines {
			data += line
			data += " "
			fmt.Println(line)
		}
		data += "?"
	}
	fmt.Println("data", data)
	return data
}

func main() {
	// recorrerDirectorio("./LogDNS1")
	// recorrerDirectorio("./ZFDNS1")
	var wg sync.WaitGroup
	wg.Add(4)

	//server de admin y server de broker
	go ServerA()
	go ServerB()
	go func() {
		for {
			var wg2 sync.WaitGroup
			wg2.Add(4)

			timer2 := time.NewTimer(20 * time.Second)
			<-timer2.C
			// go clientDNS3()
			go clientDNS2(&wg2)
			timer3 := time.NewTimer(5 * time.Second)
			<-timer3.C
			go clientDNS2confirmation(&wg2)
			timer4 := time.NewTimer(5 * time.Second)
			<-timer4.C
			go clientDNS3(&wg2)
			timer5 := time.NewTimer(5 * time.Second)
			<-timer5.C
			go clientDNS3confirmation(&wg2)
			timer6 := time.NewTimer(5 * time.Second)
			<-timer6.C
			fmt.Println("300 segundos transcurridos")
			wg2.Wait()
		}
	}()
	wg.Wait()
	return
}
