package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

const (
	sendnum = iota
	attack
	retreat
)

const localAddr = "192.168.197.1:8002"

var addrs = []string{
	"192.168.197.1:8001"}

type tmsg struct {
	Code int
	Addr string
	Num  int
}

var chRecord chan map[string]int //grabar canal

func main() {
	chRecord = make(chan map[string]int, 1) //creo un canal de mapas con string de llaves y valores int
	chRecord <- map[string]int{} //le agreagas al canal un map de array de int con clave string
	go server()  //ejecutar el servidor en paralelo 
	var op string
	for {
		fmt.Println("Jalamos? [y/n]: ")
		fmt.Scanln(&op)
		var num int
		if op == "y" {
			num = attack // num  = 1
		} else {
			num = retreat // num = 2
		}
		record := <-chRecord //envia el valor del canal a record
		record[localAddr] = num //se le guarda el numero a un arreglo q tiene la ip server
		chRecord <- record   // se guarda  record en el canal chRecord
		for _, addr := range addrs {
			send(addr, tmsg{sendnum, localAddr, num}) // envia el mensaje a todas las direcciones
		}
	}
}
func send(remoteAddr string, msg tmsg) {
	if conn, err := net.Dial("tcp", remoteAddr); err != nil {
		log.Println("Can't dial", remoteAddr)  // no se pudo conectar a la direccion remota
	} else {
		defer conn.Close()   //se cierra la conexion
		enc := json.NewEncoder(conn)  // se codifica la conexion
		fmt.Println("Sending", remoteAddr, msg)  // se envia al acceso remoto el mensaje
		enc.Encode(msg) // se codifica el mensaje
	}
}
func server() {
	if ln, err := net.Listen("tcp", localAddr); err != nil {
		log.Panicln("Can't listen on", localAddr)  // no hay conexion local 
	} else {
		defer ln.Close()
		for {
			if conn, err := ln.Accept(); err != nil {
				log.Println("Can't accept connection") // no se puede aceptar conexion con el host remoto
			} else {
				go handle(conn) // se encarga de manejar el mensaje
			}
		}
	}
}
func handle(conn net.Conn) {
	defer conn.Close()  //cierra la conexion
	dec := json.NewDecoder(conn)  // decodifica la conexion
	var msg tmsg
	if err := dec.Decode(&msg); err != nil {
		log.Println("Can't understand", conn.RemoteAddr()) //no se puede leer o entender el mensaje
	} else {
		fmt.Println("Got", msg) // se obtuvo el mensaje
		switch msg.Code {
		case sendnum:
			consensus(msg) // se envia
		}
	}
}
func consensus(msg tmsg) {
	record := <-chRecord // el canal envia record
	record[msg.Addr] = msg.Num 
	if len(record) == len(addrs)+1 {
		contAttack := 0
		contRetreat := 0
		for _, op := range record {
			if op == attack {
				contAttack++
			} else {
				contRetreat++
			}
		}
		if contAttack > contRetreat {
			fmt.Println("ATTACK!")
		} else {
			fmt.Println("FLEE!")
		}
		record = map[string]int{}
	}
	chRecord <- record
}
