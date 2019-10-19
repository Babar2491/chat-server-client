package main

import
(
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var version string = "1.0.0"
var mu sync.Mutex
var allClients []Client


type Client struct{
	conn net.Conn
	nick string
}

func (c *Client) readMessages(){
	reader := bufio.NewReader(c.conn)
	for {
		data, err := reader.ReadString('\n')
		if err != nil{
			if err == io.EOF {
				var identity string
				if len(c.nick) > 1 {
					identity = c.nick
				}else{
					identity = c.conn.RemoteAddr().String()
				}
				log.Printf("Client %s has disconnected", identity)
				var i int
				for i = range allClients {
					if allClients[i] == *c {
						break
					}
				}
				mu.Lock()
				allClients = append(allClients[:i], allClients[i+1:]...)
				fmt.Printf("Number of clients now: %d", len(allClients))
				mu.Unlock()
				return
			}
			log.Println(err)
		}
		if len(data) < 1 {
			continue
		}

		data = data[:len(data) - 1]
		if strings.HasPrefix(data, "NICK "){
			nick := data[5:]
			var validNick = regexp.MustCompile(`^[A-Za-z0-9\\_]+$`)
			if ! validNick.MatchString(nick){
				_ = c.sendMessage("Nick must only have characters (A-Za-z0-9\\_)")
				continue
			}
			if len(nick) > 12{
				_ = c.sendMessage("Nick must only have 12 characters)")
				continue
			}
			c.nick = nick
			_ = c.sendMessage("OK")
		} else if strings.HasPrefix(data, "MSG ") {
			if len(c.nick) == 0 {
				_ = c.sendMessage("Error: No nick set")
				continue
			}
			if len(data[4:]) > 255 {
				_ = c.sendMessage("Error: Message length should be <= 255")
			}
			broadcastMessage(data[:4] + c.nick + data[3:])
		} else {
			log.Println(data)
			_ = c.sendMessage("Error: malformed data")
		}
		log.Println(data)
	}

}

func (c *Client) sendMessage(msg string) error{
	if msg[len(msg) - 1:] != "\n"{
		msg += "\n"
	}
	_, err := c.conn.Write([]byte(msg))
	return err
}

func main() {
	var port = flag.Int("port", 5000, "Port to run server at")
	flag.Parse()

	log.Println("Starting server on Port: " + strconv.Itoa(*port))

	ln, err := net.Listen("tcp4", ":" + strconv.Itoa(*port))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Println("Server Started")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("Client connected: ")
		fmt.Println(conn.RemoteAddr())
		go registerClient(conn)
	}
}

func registerClient(conn net.Conn){
	client := Client{conn:conn}
	mu.Lock()
	allClients = append(allClients, client)
	mu.Unlock()
	err := client.sendMessage("Hello " + version)
	if err!= nil{
		fmt.Println(err)
	}
	go client.readMessages()
}

func broadcastMessage(msg string){
	mu.Lock()
	clients := make([]Client, len(allClients))
	copy(clients, allClients)
	fmt.Printf("Number of clients now: %d", len(clients))
	mu.Unlock()

	for _, client := range clients{
		_ = client.sendMessage(msg)
	}
}
