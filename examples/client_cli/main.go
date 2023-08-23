package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/MSSkowron/GRPCChatter/pkg/client"
)

const (
	serverAddress = ":5000"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read username from console: %s", err)
	}
	username = strings.Trim(username, "\r\n")

	c := client.NewClient(username, serverAddress)
	defer c.Disconnect()

	shortCode, err := c.CreateChatRoom("myChatRoom", "password")
	if err != nil {
		log.Fatalf("Failed to create chat room: %s", err)
	}

	if err := c.JoinChatRoom(shortCode, "password"); err != nil {
		log.Fatalf("Failed to join chat room: %s", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	receiveCh := make(chan struct{}, 1)
	sendCh := make(chan struct{}, 1)

	go receiveAndPrintMessages(c, sendCh, receiveCh, wg)
	go readAndSendMessage(c, receiveCh, sendCh, wg)

	wg.Wait()

	close(receiveCh)
	close(sendCh)

	os.Exit(0)
}

func receiveAndPrintMessages(c *client.Client, sendStopCh chan<- struct{}, receiveStopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-receiveStopCh:
			return
		default:
			msg, err := c.Receive()
			if err != nil {
				if errors.Is(err, client.ErrConnectionClosed) || errors.Is(err, client.ErrConnectionNotExists) || errors.Is(err, client.ErrStreamNotExists) {
					log.Println("Failed to receive message: lost connection with the server")
				} else {
					log.Printf("Failed to receive message: %s\n", err)
				}

				sendStopCh <- struct{}{}

				return
			}

			fmt.Printf("[%s]: %s\n", msg.Sender, msg.Body)
		}
	}
}

func readAndSendMessage(c *client.Client, sendStopCh chan<- struct{}, receiveStopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-receiveStopCh:
			return
		default:
			reader := bufio.NewReader(os.Stdin)
			msg, err := reader.ReadString('\n')
			if err != nil {
				log.Printf(" Failed to read message from console: %s\n", err)

				sendStopCh <- struct{}{}

				return
			}
			msg = strings.Trim(msg, "\r\n")

			if err := c.Send(msg); err != nil {
				if errors.Is(err, client.ErrConnectionClosed) || errors.Is(err, client.ErrConnectionNotExists) || errors.Is(err, client.ErrStreamNotExists) {
					log.Println("Failed to send message: lost connection with the server")
				} else {
					log.Printf("Failed to send message: %s\n", err)
				}

				sendStopCh <- struct{}{}

				return
			}
		}
	}
}
