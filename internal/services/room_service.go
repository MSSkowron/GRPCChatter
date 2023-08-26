package services

import (
	"errors"
	"sync"
)

var (
	ErrRoomAlreadyExist       = errors.New("room with the provided short code already exists")
	ErrRoomDoesNotExist       = errors.New("room not found")
	ErrInvalidPassword        = errors.New("invalid password")
	ErrUserAlreadyExists      = errors.New("user with the provided user name already exists in the chat room")
	ErrUserNotFound           = errors.New("user not found")
	ErrUserMessageQueueClosed = errors.New("user message queue is closed")
)

type Message struct {
	Sender string
	Body   string
}

// RoomService is an interface that defines the methods required for users and rooms management.
type RoomService interface {
	RoomExists(shortCode string) bool
	CheckPassword(shortCode string, password string) error
	CreateRoom(shortCode string, name string, password string) error
	DeleteRoom(shortCode string) error
	AddUserToRoom(shortCode string, userName string) error
	RemoveUserFromRoom(shortCode string, userName string) error
	IsUserInRoom(shortCode string, userName string) (bool, error)
	BroadcastMessageToRoom(shortCode string, message *Message) error
	GetUserMessage(shortCode string, userName string) (*Message, error)
}

// RoomServiceImpl implements the UserRoomsService interface.
type RoomServiceImpl struct {
	mu                  sync.RWMutex
	rooms               map[string]*room
	maxMessageQueueSize int
}

type room struct {
	shortCode string
	name      string
	password  string
	users     map[string]*user
}

type user struct {
	name         string
	messageQueue chan *Message
}

// NewUserRoomsServiceImpl creates a new UserRoomsServiceImpl instance with the specified short code length.
func NewRoomService(maxMessageQueueSize int) *RoomServiceImpl {
	return &RoomServiceImpl{
		maxMessageQueueSize: maxMessageQueueSize,
		rooms:               make(map[string]*room),
	}
}

func (crs *RoomServiceImpl) RoomExists(shortCode string) bool {
	crs.mu.RLock()
	defer crs.mu.RUnlock()

	_, ok := crs.rooms[shortCode]
	return ok
}

func (crs *RoomServiceImpl) CheckPassword(shortCode, password string) error {
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

func (crs *RoomServiceImpl) CreateRoom(shortCode, name, password string) error {
	crs.mu.Lock()
	defer crs.mu.Unlock()

	if _, ok := crs.rooms[shortCode]; ok {
		return ErrRoomAlreadyExist
	}

	crs.rooms[shortCode] = &room{
		shortCode: shortCode,
		name:      name,
		password:  password,
		users:     make(map[string]*user),
	}

	return nil
}

func (crs *RoomServiceImpl) DeleteRoom(shortCode string) error {
	crs.mu.Lock()
	defer crs.mu.Unlock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		return ErrRoomDoesNotExist
	}

	for _, User := range room.users {
		close(User.messageQueue)
	}

	delete(crs.rooms, shortCode)

	return nil
}

func (crs *RoomServiceImpl) AddUserToRoom(shortCode string, userName string) error {
	crs.mu.Lock()
	defer crs.mu.Unlock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		return ErrRoomDoesNotExist
	}

	if _, ok := room.users[userName]; ok {
		return ErrUserAlreadyExists
	}

	room.users[userName] = &user{
		name:         userName,
		messageQueue: make(chan *Message, crs.maxMessageQueueSize),
	}

	return nil
}

func (crs *RoomServiceImpl) RemoveUserFromRoom(shortCode string, userName string) error {
	crs.mu.Lock()
	defer crs.mu.Unlock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		return ErrRoomDoesNotExist
	}

	client, ok := room.users[userName]
	if !ok {
		return ErrUserNotFound
	}

	close(client.messageQueue)
	delete(room.users, userName)

	return nil
}

func (crs *RoomServiceImpl) IsUserInRoom(shortCode string, userName string) (bool, error) {
	crs.mu.RLock()
	defer crs.mu.RUnlock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		return false, ErrRoomDoesNotExist
	}

	_, ok = room.users[userName]
	return ok, nil
}

func (crs *RoomServiceImpl) BroadcastMessageToRoom(shortCode string, message *Message) error {
	crs.mu.RLock()
	defer crs.mu.RUnlock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		return ErrRoomDoesNotExist
	}

	for _, user := range room.users {
		if user.name != message.Sender {
			user.messageQueue <- message
		}
	}

	return nil
}

func (crs *RoomServiceImpl) GetUserMessage(shortCode string, userName string) (*Message, error) {
	crs.mu.RLock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		crs.mu.RUnlock()
		return nil, ErrRoomDoesNotExist
	}

	user, ok := room.users[userName]
	if !ok {
		crs.mu.RUnlock()
		return nil, ErrUserNotFound
	}

	crs.mu.RUnlock()
	msg, ok := <-user.messageQueue
	if !ok {
		return nil, ErrUserMessageQueueClosed
	}

	return msg, nil
}
