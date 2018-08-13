package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

var users = make(map[net.Conn]string)
var messages = make(chan string)
var bufMessage []string
var connections = make(chan net.Conn)
var removeConnect = make(chan net.Conn)

func main() {
	port := ":8080"
	network := "tcp"

	listener, err := net.Listen(network, port)
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			connections <- connection
		}
	}()

	for {
		select {
		case connection := <-connections:
			log.Printf("New user at %s", connection.RemoteAddr())
			username := getUsername(connection)
			users[connection] = username

			go readMessage(connection, username)
			go loadMessageBuf(connection, bufMessage)

		case message := <-messages:
			for connection := range users {
				go sendMessage(connection, message)
			}

		case connection := <-removeConnect:
			delete(users, connection)
			log.Printf("%s disconnected", connection.RemoteAddr())
		}
	}
}

func getUsername(connection net.Conn) string {
	io.WriteString(connection, "Please, enter your name:\n")
	reader := bufio.NewReader(connection)
	bytestr, _, _ := reader.ReadLine()
	username := string(bytestr)
	log.Printf("%s join chat", username)
	return username
}

func readMessage(connection net.Conn, username string) {
	reader := bufio.NewReader(connection)
	for msgCounter := 0; ; msgCounter++ {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
			break
		}
		msgline := fmt.Sprintf("%s: %s", username, message)
		log.Printf("New message > %s", msgline)
		messages <- msgline
		bufMessage = append(bufMessage, msgline)
	}

	removeConnect <- connection
}

func sendMessage(connection net.Conn, message string) {
	_, err := io.WriteString(connection, message)
	if err != nil {
		removeConnect <- connection
	}
}

func loadMessageBuf(connection net.Conn, messageBuf []string) {
	for _, message := range messageBuf {
		io.WriteString(connection, message)
	}
}
