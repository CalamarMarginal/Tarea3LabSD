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
	clientDNSpb "./clientDNSpb"

	"google.golang.org/grpc"
)

const ipBroker string = "10.10.28.70:50058"
const ipDNS1 string = "10.10.28.71:50052"
const ipDNS2 string = "10.10.28.72:50053" //puerto propio
const ipDNS3 string = "10.10.28.73:50054"

const ipDNS1Broker string = "10.10.28.71:50055"
const ipDNS2Broker string = "10.10.28.72:50056" //puerto propio
const ipDNS3Broker string = "10.10.28.73:50057"

const ipDNS1DNS2 string = "10.10.28.72:50050" //puerto propio
const ipDNS1DNS3 string = "10.10.28.73:50051"

var auxiliar int //si el auxiliar es 1 es porque es la primera vez que se crea el archivo

type serverAdmin struct{}
type serverBroker struct{}
type serverDNS struct{}

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
	path := "./ZFDNS2/." + varAux[1] + ".txt"
	aux := readFile(path, req.CommCliente)
	//terminoAux := strings.Split(aux, " ") //recibe ej "algo.com 3.4.5.3"
	//ipDom := terminoAux[1]
	reloj := readFileReloj(path)
	ipDNSpropia := ipDNS2Broker
	res := &brokerDNSpb.DnsClientResponse{
		IpDominio: aux,
		Reloj:     reloj,
		IpDNS:     ipDNSpropia,
	}
	return res, nil
}

func createDomain(dominio string, ip string, comando string) string {
	reloj := "0,1,0"
	aux := strings.Split(dominio, ".")
	extension := aux[1]
	extensionFinal := "." + extension
	path := "./ZFDNS2/" + extensionFinal + ".txt"
	createFile(path)
	pathLog := "./LogDNS2/" + extensionFinal + ".txt"
	createFile(pathLog)
	data := reloj + "?" + dominio + "?" + ip
	clock := writeFile(path, comando, "ZF", data)
	return clock
}

func updateDomain(dominio string, tipoCambio string, parametroNuevo string, comando string) string {
	aux := strings.Split(dominio, ".")
	extension := aux[1]
	extensionFinal := "." + extension
	path := "./ZFDNS2/" + extensionFinal + ".txt"
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
	path := "./ZFDNS2/." + extensionFinal + ".txt"
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
				path2 := "./LogDNS2/." + extension + ".txt"
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
				relojAux := strings.Split(relojAntiguo, ",")
				i, err := strconv.Atoi(relojAux[1])
				if isError(err) {
					return ""
				}
				i++
				s := strconv.Itoa(i)
				relojNuevo := relojAux[0] + "," + s + "," + relojAux[2]
				updateFile(path, relojAntiguo, relojNuevo)

				clock = relojNuevo

				aux = strings.Split(dominio, ".")
				extension := aux[1]
				path2 := "./LogDNS2/." + extension + ".txt"
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
			//--------------Reloj--------------

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
				path2 := "./LogDNS2/." + extension + ".txt"
				text := "update " + dominioFinalAntiguo + " " + valorNuevo
				writeLog(path2, text)
				err = file.Sync()
				if isError(err) {
					return "Dominio no existe"
				}

			} else if tipoDeCambio == "ip" {
				fmt.Println("entre a ip")
				dominioAntiguo := readFile(path, dominio)
				if dominioAntiguo == "" {
					return "Dominio no existe"
				}
				AUX := strings.Split(dominioAntiguo, " ")
				ipFinalAntiguo := AUX[3]

				updateFile(path, ipFinalAntiguo, valorNuevo)

				aux = strings.Split(dominio, ".")
				extension := aux[1]
				path2 := "./LogDNS2/." + extension + ".txt"
				text := "update " + dominio + " " + valorNuevo
				writeLog(path2, text)
				err = file.Sync()
				if isError(err) {
					return "Dominio no existe"
				}
			}
			relojAntiguo := readFileReloj(path)
			relojAux := strings.Split(relojAntiguo, ",")
			i, err := strconv.Atoi(relojAux[1])
			if isError(err) {
				return "Dominio no existe"
			}
			i++
			s := strconv.Itoa(i)
			relojNuevo := relojAux[0] + "," + s + "," + relojAux[2]
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
			i, err := strconv.Atoi(relojAux[1])
			if isError(err) {
				return "Dominio no existe"
			}
			i++
			s := strconv.Itoa(i)
			relojNuevo := relojAux[0] + "," + s + "," + relojAux[2]

			deleteLine(path, dominio)
			updateFile(path, relojAntiguo, relojNuevo)

			clock = relojNuevo

			help := strings.Split(dominio, ".")
			extension := help[1]
			path2 := "./LogDNS2/." + extension + ".txt"
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
		fmt.Println("-----------------------------------------")
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

func (*serverDNS) ClientDNS(ctx context.Context, req *clientDNSpb.ClienteDNSRequest) (*clientDNSpb.ClientDNSResponse, error) {
	//da igual que comando sea, el broker solo responde con la ip de una dns
	fmt.Println("Timer:", req.TimeComplete)

	dataLog := recorrerDirectorio("./LogDNS2")
	dataZf := recorrerDirectorio("./ZFDNS2")

	aux := strings.Split(dataZf, "?")

	reloj := ""

	for _, relojAux := range aux {
		if len(relojAux) > 0 {

			auxEspacio := strings.Split(relojAux, " ")
			if len(auxEspacio) > 2 {
				reloj += auxEspacio[1]
				reloj += "?"
			}
		}
	}

	// fmt.Println("reloj", reloj)

	res := &clientDNSpb.ClientDNSResponse{
		Log:   dataLog,
		Reloj: reloj,
	}

	return res, nil

}

func borrarRegistro(path string) {

	err := os.Remove(path)
	if err != nil {
		fmt.Printf("Error eliminando archivo: %v\n", err)
	} else {
		fmt.Println("Eliminado correctamente")
	}
}

func replicandoInfo(log string, zf string) {

	pathZF := ""
	pathLog := ""

	i := 0
	j := 0

	cmd := ""
	extension := ""

	auxLog := strings.Split(log, "?")
	for _, valores := range auxLog {
		valorEspecifico := strings.Split(valores, " ")
		fmt.Println("valorEspecifico: ", valorEspecifico)
		if len(valorEspecifico) > 2 {

			for k, term := range valorEspecifico {
				if k == 0 {

					extension += term
					fmt.Println("La extension es", extension)
					if len(extension) > 2 {

						pathZF = "./ZFDNS2/" + extension
						pathLog = "./LogDNS2/" + extension
						fmt.Println("pathZF", pathZF)
						fmt.Println("pathLog", pathLog)
						borrarRegistro(pathZF)
						borrarRegistro(pathLog)
						createFile(pathZF)
						createFile(pathLog)
					}
					extension = ""

				} else if k != 0 {
					if term == "create" || term == "update" {

						cmd = cmd + term
						i++
						continue
					}
					if i > 0 {
						cmd = cmd + " " + term
						i++
					}
					if i > 2 {
						var file, err = os.OpenFile(pathLog, os.O_APPEND|os.O_WRONLY, 0644)
						if isError(err) {
							fmt.Println(err)
						}
						defer file.Close()
						_, err = fmt.Fprintln(file, cmd)
						if isError(err) {
							fmt.Println(err)
						}
						fmt.Println("escribiendo data ", cmd)
						fmt.Println("en el siguiente path ", pathLog)

						fmt.Println(cmd) //aca esta el comando
						i = 0
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
						var file, err = os.OpenFile(pathLog, os.O_APPEND|os.O_WRONLY, 0644)
						if isError(err) {
							fmt.Println(err)
						}
						defer file.Close()
						_, err = fmt.Fprintln(file, cmd)
						if isError(err) {
							fmt.Println(err)
						}

						fmt.Println("escribiendo data ", cmd)
						fmt.Println("en el siguiente path ", pathLog)

						fmt.Println(cmd)
						cmd = ""
						j = 0

					}
				}

			}
		}
	}

	auxZf := strings.Split(zf, "?")
	data := ""
	cont := 0
	extensiones := ""

	for _, a := range auxZf {

		a_aux := strings.Fields(a)
		fmt.Println(a)
		fmt.Println(a_aux)
		for i, x := range a_aux {
			if i == 0 {

				extension := x
				pathZF = "./ZFDNS2/" + extension
				extensiones += extension
				extensiones += "?"
				fmt.Println("extension", extension)
			}
			if i == 1 {
				reloj := x
				fmt.Println("reloj es ", reloj)
				fmt.Println("escribiendo reloj ", reloj)
				fmt.Println("en el siguiente path ", pathZF)
				var file, err = os.OpenFile(pathZF, os.O_APPEND|os.O_WRONLY, 0644)
				if isError(err) {
					fmt.Println(err)
				}
				defer file.Close()
				_, err = fmt.Fprintln(file, reloj)
				if isError(err) {
					fmt.Println(err)
				}

			} else if i > 1 && cont < 4 {
				data += x
				data += " "
				cont++
			}
			if cont == 4 {
				data += "?"
				cont = 0

			}
		}

	}

	fmt.Println("data", data)
	fmt.Println("extensiones", data)

	aux_2 := strings.Split(data, "?")
	//aux_3 := strings.Split(extensiones, "?")

	//

	cont = 0

	for _, x := range aux_2 {

		if len(x) > 2 {
			aux_4 := strings.Fields(x)
			fmt.Println("aux_4", aux_4)
			fmt.Println("aux_4[0]", aux_4[0])

			separacion := strings.Split(aux_4[0], ".")
			fmt.Println("separacion", separacion)
			extension_sin_punto := separacion[1]
			extension_final := "." + extension_sin_punto + ".txt"
			pathZF = "./ZFDNS2/" + extension_final
			data = strings.Join(aux_4, " ")
			fmt.Println("data es", data)
			fmt.Println("path es ", pathZF)
			var file, err = os.OpenFile(pathZF, os.O_APPEND|os.O_WRONLY, 0644)
			if isError(err) {
				fmt.Println(err)
			}
			defer file.Close()
			_, err = fmt.Fprintln(file, data)
			if isError(err) {
				fmt.Println(err)
			}

			//fmt.Println("aux_2 es",aux_2)

		}
	}

}

func (*serverDNS) ClientDNSConfirmation(ctx context.Context, req *clientDNSpb.ClientDNSRequestConfirmation) (*clientDNSpb.ClientDNSResponseConfirmation, error) {
	fmt.Println("log-----:", req.GetLog())
	fmt.Println("zf------:", req.GetZf())

	res := &clientDNSpb.ClientDNSResponseConfirmation{
		Ack: "replicando informacion en el DNS2",
	}

	replicandoInfo(req.GetLog(), req.GetZf())

	return res, nil
}

func ServerA() { //servidor para admin
	fmt.Println("DNS admin server is running")

	lis, err := net.Listen("tcp", ipDNS2) //este puerto usa el admin para conectarse

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

	lis, err := net.Listen("tcp", ipDNS2Broker) //este puerto usa el broker para conectarse

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

func ServerDNS1() {

	fmt.Println("DNS2&DNS1 server is running")

	lis, err := net.Listen("tcp", ipDNS1DNS2) //este puerto usa el broker para conectarse

	if err != nil {
		log.Printf("Failed to listen %v", err)
	}

	//asignar servidor de grpc a s
	s := grpc.NewServer()

	//se utiliza el archivo que se genera por el protocol buffer y utilizaremos el metodo Register y el nombre del servicio
	// le pasasomos el servidor de grpc (s) y la estructura de un servidor "server"
	clientDNSpb.RegisterClientDNSServiceServer(s, &serverDNS{})

	if err := s.Serve(lis); err != nil {
		log.Printf("Failed to serve %v", err)
	}
}

func main() {

	var wg sync.WaitGroup
	wg.Add(4)

	//server de admin y server de broker
	go ServerA()
	go ServerB()
	go ServerDNS1()
	wg.Wait()
	return
}
