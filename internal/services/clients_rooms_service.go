package services

import (
	"errors"
	"sync"
)

var (
	ErrRoomAlreadyExist = errors.New("room with the short code already exists")
	ErrRoomDoesNotExist = errors.New("room does not exist")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrClientNotFound   = errors.New("client not found")
	ErrClientClosed     = errors.New("client closed")
)

// ClientsRoomsService is an interface that defines the methods required for clients rooms management.
type ClientsRoomsService interface {
	RoomExists(shortCode string) bool
	CheckPassword(shortCode string, password string) error
	CreateRoom(shortCode string, name string, password string) error
	DeleteRoom(shortCode string) error
	AddClientToRoom(shortCode string, clientName string) error
	RemoveClientFromRoom(shortCode string, clientName string) error
	IsClientInRoom(shortCode string, clientName string) (bool, error)
	BroadcastMessageToRoom(shortCode string, message *Message) error
	GetClientMessage(shortCode string, clientName string) (*Message, error)
}

type Message struct {
	Sender string
	Body   string
}

// ClientsRoomsServiceImpl implements the ClientRoomsService interface.
type ClientsRoomsServiceImpl struct {
	mu    sync.RWMutex
	rooms map[string]*room

	maxMessageQueueSize int
}

type room struct {
	shortCode string
	name      string
	password  string

	clients []*client
}

type client struct {
	name         string
	messageQueue chan *Message
}

// NewClientRoomsServiceImpl creates a new ClientRoomsServiceImpl instance with the specified short code length.
func NewClientsRoomsServiceImpl(maxMessageQueueSize int) *ClientsRoomsServiceImpl {
	return &ClientsRoomsServiceImpl{
		maxMessageQueueSize: maxMessageQueueSize,
		rooms:               make(map[string]*room),
	}
}

func (crs *ClientsRoomsServiceImpl) RoomExists(shortCode string) bool {
	crs.mu.RLock()
	defer crs.mu.RUnlock()

	_, ok := crs.rooms[shortCode]
	return ok
}

func (crs *ClientsRoomsServiceImpl) CheckPassword(shortCode, password string) error {
	crs.mu.RLock()
	defer crs.mu.RUnlock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		return ErrRoomDoesNotExist
	}

	if room.password == password {
		return nil
	}

	return ErrInvalidPassword
}

func (crs *ClientsRoomsServiceImpl) CreateRoom(shortCode, name, password string) error {
	crs.mu.Lock()
	defer crs.mu.Unlock()

	if _, ok := crs.rooms[shortCode]; ok {
		return ErrRoomAlreadyExist
	}

	crs.rooms[shortCode] = &room{
		shortCode: shortCode,
		name:      name,
		password:  password,
		clients:   make([]*client, 0),
	}

	return nil
}

func (crs *ClientsRoomsServiceImpl) DeleteRoom(shortCode string) error {
	crs.mu.Lock()
	defer crs.mu.Unlock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		return ErrRoomDoesNotExist
	}

	for _, client := range room.clients {
		close(client.messageQueue)
	}

	delete(crs.rooms, shortCode)

	return nil
}

func (crs *ClientsRoomsServiceImpl) AddClientToRoom(shortCode string, clientName string) error {
	crs.mu.Lock()
	defer crs.mu.Unlock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		return ErrRoomDoesNotExist
	}

	room.clients = append(room.clients, &client{
		name:         clientName,
		messageQueue: make(chan *Message, crs.maxMessageQueueSize),
	})

	return nil
}

func (crs *ClientsRoomsServiceImpl) RemoveClientFromRoom(shortCode string, clientName string) error {
	crs.mu.Lock()
	defer crs.mu.Unlock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		return ErrRoomDoesNotExist
	}

	for idx, client := range room.clients {
		if client.name == clientName {
			close(client.messageQueue)
			room.clients = append(room.clients[:idx], room.clients[idx+1:]...)
			break
		}
	}

	return nil
}

func (crs *ClientsRoomsServiceImpl) IsClientInRoom(shortCode string, clientName string) (bool, error) {
	crs.mu.RLock()
	defer crs.mu.RUnlock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		return false, ErrRoomDoesNotExist
	}

	for _, client := range room.clients {
		if client.name == clientName {
			return true, nil
		}
	}

	return false, nil
}

func (crs *ClientsRoomsServiceImpl) BroadcastMessageToRoom(shortCode string, message *Message) error {
	crs.mu.RLock()
	defer crs.mu.RUnlock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		return ErrRoomDoesNotExist
	}

	for _, client := range room.clients {
		if client.name != message.Sender {
			client.messageQueue <- message
		}
	}

	return nil
}

func (crs *ClientsRoomsServiceImpl) GetClientMessage(shortCode string, clientName string) (*Message, error) {
	crs.mu.RLock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		crs.mu.RUnlock()
		return nil, ErrRoomDoesNotExist
	}

	for _, client := range room.clients {
		if client.name == clientName {
			crs.mu.RUnlock()
			msg, ok := <-client.messageQueue
			if !ok {
				return nil, ErrClientClosed
			}

			return msg, nil
		}
	}

	crs.mu.RUnlock()
	return nil, ErrClientNotFound
}
