package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/MSSkowron/GRPCChatter/pkg/client"
)

const (
	serverAddress = ":5000"
)

func main() {
	fmt.Print("Enter your username: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		os.Exit(1)
	}
	username := scanner.Text()

	c := client.NewClient(username, serverAddress)
	defer c.Close()

	if err := c.Join(); err != nil {
		log.Fatalf("Failed to join the chat: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	stopCh := make(chan struct{})

	go receiveAndPrintMessages(c, stopCh, wg)
	go readAndSendMessage(c, stopCh, wg)

	wg.Wait()
}

func receiveAndPrintMessages(c *client.Client, stopCh chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-stopCh:
			return
		default:
			msg, err := c.Receive()
			if err != nil {
				if errors.Is(err, client.ErrConnectionClosed) || errors.Is(err, client.ErrConnectionNotExists) || errors.Is(err, client.ErrStreamNotExists) {
					log.Println("Error receiving message: lost connection with the server")
				} else {
					log.Printf("Error receiving message: %s\n", err)
				}

				select {
				case <-stopCh:
					return
				case stopCh <- struct{}{}:
					return
				}
			}

			fmt.Printf("[%s]: %s\n", msg.Sender, msg.Body)
		}
	}
}

func readAndSendMessage(c *client.Client, stopCh chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-stopCh:
			return
		default:
			if err := c.Send("hello"); err != nil {
				if errors.Is(err, client.ErrConnectionClosed) || errors.Is(err, client.ErrConnectionNotExists) || errors.Is(err, client.ErrStreamNotExists) {
					log.Println("Error sending message: lost connection with the server")
				} else {
					log.Printf("Error sending message: %s\n", err)
				}

				select {
				case <-stopCh:
					return
				case stopCh <- struct{}{}:
					return
				}
			}
		}
	}
}
