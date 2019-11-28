package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

var addrs []string

func registerSend(remoteAddr, hostAddr string) {
	remote := fmt.Sprintf("%s:8000", remoteAddr)
	conn, _ := net.Dial("tcp", remote)
	defer conn.Close()

	// Enviar direccion
	fmt.Fprintln(conn, hostAddr)

	// Recibir lista de direcciones
	r := bufio.NewReader(conn)
	strAddrs, _ := r.ReadString('\n')
	var respAddrs []string
	json.Unmarshal([]byte(strAddrs), &respAddrs)

	// agregamos direcciones de nodos a propia libreta
	for _, addr := range respAddrs {
		if addr == remoteAddr {
			return
		}
	}
	addrs = append(respAddrs, remoteAddr)
	fmt.Println(addrs)
}
func registerServer(hostAddr string) {
	host := fmt.Sprintf("%s:8000", hostAddr)
	ln, _ := net.Listen("tcp", host)
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		go handleRegister(conn)
	}
}
func handleRegister(conn net.Conn) {
	defer conn.Close()

	// Recibimos addr del nuevo nodo
	r := bufio.NewReader(conn)
	remoteIp, _ := r.ReadString('\n')
	remoteIp = strings.TrimSpace(remoteIp)

	// respondemos enviando lista de direcciones de nodos actuales
	byteAddrs, _ := json.Marshal(addrs)
	fmt.Fprintf(conn, "%s\n", string(byteAddrs))

	// notificar a nodos actuales de llegada de nuevo nodo
	for _, addr := range addrs {
		notifySend(addr, remoteIp)
	}

	// Agregamos nuevo nodo a la lista de direcciones
	for _, addr := range addrs {
		if addr == remoteIp {
			return
		}
	}
	addrs = append(addrs, remoteIp)
	fmt.Println(addrs)
}
func notifySend(addr, remoteIp string) {
	remote := fmt.Sprintf("%s:8001", addr)
	conn, _ := net.Dial("tcp", remote)
	defer conn.Close()
	fmt.Fprintln(conn, remoteIp)
}
func notifyServer(hostAddr string) {
	host := fmt.Sprintf("%s:8001", hostAddr)
	ln, _ := net.Listen("tcp", host)
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		go handleNotify(conn)
	}
}
func handleNotify(conn net.Conn) {
	defer conn.Close()

	// Recibimos addr del nuevo nodo
	r := bufio.NewReader(conn)
	remoteIp, _ := r.ReadString('\n')
	remoteIp = strings.TrimSpace(remoteIp)

	// Agregamos nuevo nodo a la lista de direcciones
	for _, addr := range addrs {
		if addr == remoteIp {
			return
		}
	}
	addrs = append(addrs, remoteIp)
	fmt.Println(addrs)
}
func main() {
	myIp := "10.142.113.32"
	go registerServer(myIp)

	gin := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese direccion remota: ")
	remoteIp, _ := gin.ReadString('\n')
	remoteIp = strings.TrimSpace(remoteIp)

	if remoteIp != "" {
		registerSend(remoteIp, myIp)
	}

	notifyServer(myIp)
}
