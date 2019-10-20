package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	connection       net.Conn
	nick             string
	disconnected     bool
	outMessages      chan string
	dontShowMessages []string
}

var client = Client{disconnected: true, outMessages: make(chan string)}
var mu sync.Mutex

func main() {
	var ipAddrStr string
	var port int
	flag.StringVar(&ipAddrStr, "ip", "", "(REQUIRED) IP of the chat server (x.x.x.x)")
	flag.IntVar(&port, "port", 0, "(REQUIRED) Port number of the chat server (1 - 65535)")
	flag.Parse()

	ipaddress := net.ParseIP(ipAddrStr)
	if len(ipAddrStr) < 7 || ipaddress == nil {
		log.Println("Please provide a valid IP Address for server in command line flags")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if port < 1 || port > 65535 {
		log.Println("Please provide a valid Port for server in command line flags")
		flag.PrintDefaults()
		os.Exit(1)
	}
	fmt.Println("Initializing ...")
	go maintainConnectionToServer(ipAddrStr, strconv.Itoa(port))
	time.Sleep(2 * time.Second)
	go readMessagesFromServer()
	time.Sleep(2 * time.Second)
	go writeDataToConnection()
	getNick()
	fmt.Println("Your nick is set, You can now start chatting with your friends!")
	getAndSendMessages()
}

func maintainConnectionToServer(serverIP string, serverPort string) {
	for {
		clientDisconnected := client.disconnected
		if clientDisconnected {
			connection, err := net.Dial("tcp", serverIP+":"+serverPort)
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}
			mu.Lock()
			client.connection = connection
			client.disconnected = false
			mu.Unlock()
			if client.nick != "" {
				client.outMessages <- "NICK " + client.nick
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func getNick() {
	fmt.Print("Enter your nick \nIt must have:\n- characters (A-Za-z0-9\\_)\n- Max 12 Characters\n> ")

	reader := bufio.NewReader(os.Stdin)
	tNick, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	tNick = tNick[:len(tNick)-1]
	var validNick = regexp.MustCompile(`^[A-Za-z0-9\\_]+$`)
	if !validNick.MatchString(tNick) {
		log.Fatal("Nick must only have characters (A-Za-z0-9\\_)")
	}
	if len(tNick) > 12 {
		log.Fatal("Nick must only have 12 characters)")
	}
	client.nick = tNick
	client.outMessages <- "NICK " + client.nick
}

func readMessagesFromServer() {
	for {
		if client.disconnected {
			time.Sleep(2 * time.Second)
			continue
		}
		reader := bufio.NewReader(client.connection)
		message, err := reader.ReadString('\n')

		if err == io.EOF {
			client.disconnected = true
		}
		dontShow := false
		for index, dontShowMessage := range client.dontShowMessages {
			if dontShowMessage == message {
				dontShow = true
				client.dontShowMessages = append(client.dontShowMessages[:index], client.dontShowMessages[index+1:]...)
				break
			}
		}
		if dontShow {
			continue
		}
		fmt.Println("Received: " + message)
	}
}

func getAndSendMessages() {
	for {
		reader := bufio.NewReader(os.Stdin)
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if len(message) < 1 {
			continue
		}
		formattedMsg := "MSG " + message
		client.outMessages <- formattedMsg
	}
}

func writeDataToConnection() {
	for {
		if client.disconnected {
			time.Sleep(2 * time.Second)
			continue
		}
		data := <-client.outMessages
		msg := data
		if msg[len(msg)-1:] != "\n" {
			msg += "\n"
		}
		_, err := client.connection.Write([]byte(msg))
		if err != nil {
			client.outMessages <- data
		}
		if strings.HasPrefix(msg, "MSG ") {
			client.dontShowMessages = append(client.dontShowMessages, msg[:4]+client.nick+msg[3:])
		} else if strings.HasPrefix(msg, "NICK ") {
			client.nick = msg[5 : len(msg)-1]
		}
	}
}
