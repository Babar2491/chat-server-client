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
	"time"
)

var connection net.Conn
var nick string

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
	if  port < 1 || port > 65535{
		log.Println("Please provide a valid Port for server in command line flags")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var err error
	connection, err = net.Dial("tcp", ipAddrStr + ":" + strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()
	go readMessagesFromServer()
	time.Sleep(1 * time.Second)
	getNick()
	fmt.Println("Your nick is set, You can now start chatting with your friends!")
	getAndSendMessages()
}

func getNick(){
	fmt.Print("Enter your nick \nIt must have:\n- characters (A-Za-z0-9\\_)\n- Max 12 Characters\n> ")

	reader := bufio.NewReader(os.Stdin)
	tNick, err := reader.ReadString('\n')
	if err != nil{
		log.Fatal(err)
	}
	tNick = tNick[:len(tNick) - 1]
	var validNick = regexp.MustCompile(`^[A-Za-z0-9\\_]+$`)
	if ! validNick.MatchString(tNick){
		log.Fatal("Nick must only have characters (A-Za-z0-9\\_)")
	}
	if len(tNick) > 12 {
		log.Fatal("Nick must only have 12 characters)")
	}
	nick = tNick
	writeDataToConnection("NICK " + nick)
}

func readMessagesFromServer() {
	for {
		reader := bufio.NewReader(connection)
		message, err := reader.ReadString('\n')

		if err == io.EOF {
			_ = connection.Close()
			fmt.Println("Connection closed from Server.")
			os.Exit(0)
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
		message = "MSG " + message
		writeDataToConnection(message)
	}
}

func writeDataToConnection(data string){
	if data[len(data) - 1:] != "\n"{
		data += "\n"
	}
	_, err := connection.Write([]byte(data))
	if err != nil {
		log.Println(err)
	}
}
