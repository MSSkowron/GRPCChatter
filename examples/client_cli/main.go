package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/MSSkowron/GRPCChatter/pkg/client"
	"golang.org/x/term"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter username: ")
	userName, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read username from console: %s\n", err)
	}
	userName = strings.Trim(userName, "\r\n")

	fmt.Printf("Enter server address: ")
	serverAddress, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read username from console: %s\n", err)
	}
	serverAddress = strings.Trim(serverAddress, "\r\n")

	c := client.NewClient(userName, serverAddress)
	defer c.Disconnect()

	fmt.Printf("\n")
	for {
		fmt.Println("Menu:")
		fmt.Println("1. Create chat room")
		fmt.Println("2. Join chat room")
		fmt.Println("3. Quit")
		fmt.Print("> ")

		choice, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read choice from console: %s\n", err)
		}
		choice = strings.Trim(choice, "\r\n")

		switch choice {
		case "1":
			fmt.Print("Enter chat room name: ")
			roomName, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("Failed to read chat room name: %s\n", err)
			}
			roomName = strings.Trim(roomName, "\r\n")

			fmt.Print("Enter password: ")
			password, err := readPassword()
			if err != nil {
				log.Fatalf("Failed to read password: %s\n", err)
			}

			shortCode, err := c.CreateChatRoom(roomName, password)
			if err != nil {
				log.Printf("\nFailed to create chat room: %s\n", err)
				continue
			}

			fmt.Printf("\nChat room created. Short code: %s\n", shortCode)
		case "2":
			fmt.Print("Enter chat room short code: ")
			shortCode, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("Failed to read chat room short code: %s\n", err)
			}
			shortCode = strings.Trim(shortCode, "\r\n")

			fmt.Print("Enter chat room password: ")
			password, err := readPassword()
			if err != nil {
				log.Fatalf("Failed to read password: %s\n", err)
			}

			if err := c.JoinChatRoom(shortCode, password); err != nil {
				log.Printf("\nFailed to join chat room: %s\n", err)
				continue
			}

			fmt.Printf("\nSuccessfully joined the chat room.\n")

			wg := &sync.WaitGroup{}
			wg.Add(2)

			receiveCh := make(chan struct{}, 1)
			sendCh := make(chan struct{}, 1)

			go receiveAndPrintMessages(c, sendCh, receiveCh, wg)
			go readAndSendMessage(c, receiveCh, sendCh, wg)

			wg.Wait()

			close(receiveCh)
			close(sendCh)
		case "3":
			fmt.Printf("\nGoodbye!\n")
			return
		default:
			fmt.Printf("\nInvalid choice. Please select a valid option.\n")
		}
	}
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

func readPassword() (string, error) {
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(passwordBytes), nil
}
