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

	fmt.Printf("Enter REST server address: ")
	restServerAddress, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read username from console: %s\n", err)
	}
	restServerAddress = strings.Trim(restServerAddress, "\r\n")

	fmt.Printf("Enter gRPC server address: ")
	grpcServerAddress, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read username from console: %s\n", err)
	}
	grpcServerAddress = strings.Trim(grpcServerAddress, "\r\n")

	c := client.NewClient(restServerAddress, grpcServerAddress)
	defer c.Disconnect()

	fmt.Printf("\n")
	for {
		fmt.Println("Menu:")
		fmt.Println("1. Register")
		fmt.Println("2. Login")
		fmt.Println("3. Create chat room")
		fmt.Println("4. Join chat room")
		fmt.Println("5. Quit")
		fmt.Print("> ")

		choice, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read choice from console: %s\n", err)
		}

		choice = strings.Trim(choice, "\r\n")
		switch choice {
		case "1":
			fmt.Print("Enter user name: ")
			userName, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("Failed to read username name: %s\n", err)
			}
			userName = strings.Trim(userName, "\r\n")

			fmt.Print("Enter password: ")
			password, err := readPassword()
			if err != nil {
				log.Fatalf("Failed to read password: %s\n", err)
			}

			if err := c.Register(userName, password); err != nil {
				log.Printf("\nFailed to register: %s\n", err)
				continue
			}

			fmt.Printf("\nRegistered.\n")
		case "2":
			fmt.Print("Enter user name: ")
			userName, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("Failed to read username name: %s\n", err)
			}
			userName = strings.Trim(userName, "\r\n")

			fmt.Print("Enter password: ")
			password, err := readPassword()
			if err != nil {
				log.Fatalf("Failed to read password: %s\n", err)
			}

			if err := c.Login(userName, password); err != nil {
				log.Printf("\nFailed to log in: %s\n", err)
				continue
			}

			fmt.Printf("\nLogged in.\n")
		case "3":
			fmt.Print("Enter name: ")
			name, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("Failed to read chat room name: %s\n", err)
			}
			name = strings.Trim(name, "\r\n")

			fmt.Print("Enter password: ")
			password, err := readPassword()
			if err != nil {
				log.Fatalf("Failed to read password: %s\n", err)
			}

			shortCode, err := c.CreateChatRoom(name, password)
			if err != nil {
				log.Printf("\nFailed to create chat room: %s\n", err)
				continue
			}

			fmt.Printf("\nChat room created. Short code: %s\n", shortCode)
		case "4":
			fmt.Print("Enter short code: ")
			shortCode, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("Failed to read chat room short code: %s\n", err)
			}
			shortCode = strings.Trim(shortCode, "\r\n")

			fmt.Print("Enter password: ")
			password, err := readPassword()
			if err != nil {
				log.Fatalf("Failed to read password: %s\n", err)
			}

			if err := c.JoinChatRoom(shortCode, password); err != nil {
				log.Printf("\nFailed to join chat room: %s\n", err)
				continue
			}

			fmt.Printf("\nJoined chat room.\n")

			wg := &sync.WaitGroup{}
			wg.Add(2)

			receiveCh := make(chan struct{}, 1)
			sendCh := make(chan struct{}, 1)

			go receiveAndPrintMessages(c, sendCh, receiveCh, wg)
			go readAndSendMessages(c, receiveCh, sendCh, wg)

			wg.Wait()

			close(receiveCh)
			close(sendCh)
		case "5":
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
				if errors.Is(err, client.ErrConnectionClosed) || errors.Is(err, client.ErrConnectionNotExist) || errors.Is(err, client.ErrStreamNotExist) {
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

func readAndSendMessages(c *client.Client, sendStopCh chan<- struct{}, receiveStopCh <-chan struct{}, wg *sync.WaitGroup) {
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
				if errors.Is(err, client.ErrConnectionClosed) || errors.Is(err, client.ErrConnectionNotExist) || errors.Is(err, client.ErrStreamNotExist) {
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
