package services

import (
	"errors"
	"sync"

	"github.com/MSSkowron/BookRESTAPI/pkg/crypto"
)

var (
	// ErrRoomAlreadyExist is returned when a room with the provided short code already exists.
	ErrRoomAlreadyExist = errors.New("room with the provided short code already exists")
	// ErrRoomDoesNotExist is returned when a requested room is not found.
	ErrRoomDoesNotExist = errors.New("room not found")
	// ErrInvalidPassword is returned when a provided password is invalid.
	ErrInvalidPassword = errors.New("invalid password")
	// ErrUserAlreadyExists is returned when a user with the provided user name already exists in the chat room.
	ErrUserAlreadyExists = errors.New("user with the provided user name already exists in the chat room")
	// ErrUserNotFound is returned when a requested user is not found.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserMessageQueueClosed is returned when a user's message queue is closed.
	ErrUserMessageQueueClosed = errors.New("user message queue is closed")
)

// Message represents a chat message with a sender and body.
type Message struct {
	// Sender is the name of the user who sent the message.
	Sender string

	// Body is the content of the message.
	Body string
}

// RoomService is an interface that defines the methods required for users and rooms management.
type RoomService interface {
	// RoomExists checks if a room with the given short code exists.
	RoomExists(shortCode string) bool

	// CheckPassword checks if the provided password matches the room's password.
	CheckPassword(shortCode, password string) error

	// CreateRoom creates a new chat room with the given short code, name, and password.
	CreateRoom(shortCode, name, password string) error

	// DeleteRoom deletes a chat room with the given short code.
	DeleteRoom(shortCode string) error

	// AddUserToRoom adds a user to a chat room with the given short code and user name.
	AddUserToRoom(shortCode string, userName string) error

	// RemoveUserFromRoom removes a user from a chat room with the given short code and user name.
	RemoveUserFromRoom(shortCode string, userName string) error

	// GetRoomUsers retrieves the list of user names currently in a chat room with the provided short code.
	GetRoomUsers(shortCode string) ([]string, error)

	// IsUserInRoom checks if a user with the given user name is in the chat room with the provided short code.
	IsUserInRoom(shortCode string, userName string) (bool, error)

	// BroadcastMessageToRoom broadcasts a message to all users in a chat room with the given short code.
	BroadcastMessageToRoom(shortCode string, message *Message) error

	// GetUserMessage retrieves a message from a user's message queue in a chat room.
	GetUserMessage(shortCode string, userName string) (*Message, error)
}

// RoomServiceImpl implements the RoomService interface.
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

// NewRoomService creates a new RoomServiceImpl instance with the specified maximum message queue size.
func NewRoomService(maxMessageQueueSize int) *RoomServiceImpl {
	return &RoomServiceImpl{
		maxMessageQueueSize: maxMessageQueueSize,
		rooms:               make(map[string]*room),
	}
}

// RoomExists checks if a room with the given short code exists.
func (crs *RoomServiceImpl) RoomExists(shortCode string) bool {
	crs.mu.RLock()
	defer crs.mu.RUnlock()

	_, ok := crs.rooms[shortCode]
	return ok
}

// CheckPassword checks if the provided password matches the room's password.
func (crs *RoomServiceImpl) CheckPassword(shortCode, password string) error {
	crs.mu.RLock()
	defer crs.mu.RUnlock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		return ErrRoomDoesNotExist
	}

	if err := crypto.CheckPassword(password, room.password); err != nil {
		if errors.Is(err, crypto.ErrInvalidCredentials) {
			return ErrInvalidPassword
		}

		return err
	}

	return nil
}

// CreateRoom creates a new chat room with the given short code, name, and password.
func (crs *RoomServiceImpl) CreateRoom(shortCode, name, password string) error {
	crs.mu.Lock()
	defer crs.mu.Unlock()

	if _, ok := crs.rooms[shortCode]; ok {
		return ErrRoomAlreadyExist
	}

	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}

	crs.rooms[shortCode] = &room{
		shortCode: shortCode,
		name:      name,
		password:  hashedPassword,
		users:     make(map[string]*user),
	}

	return nil
}

// DeleteRoom deletes a chat room with the given short code.
func (crs *RoomServiceImpl) DeleteRoom(shortCode string) error {
	crs.mu.Lock()
	defer crs.mu.Unlock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		return ErrRoomDoesNotExist
	}

	for _, user := range room.users {
		close(user.messageQueue)
	}

	delete(crs.rooms, shortCode)

	return nil
}

// AddUserToRoom adds a user to a chat room with the given short code and user name.
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

// RemoveUserFromRoom removes a user from a chat room with the given short code and user name.
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

// GetRoomUsers retrieves the list of user names currently in a chat room with the provided short code.
func (crs *RoomServiceImpl) GetRoomUsers(shortCode string) ([]string, error) {
	crs.mu.RLock()
	defer crs.mu.RUnlock()

	room, ok := crs.rooms[shortCode]
	if !ok {
		return nil, ErrRoomDoesNotExist
	}

	users := make([]string, 0, len(room.users))
	for userName := range room.users {
		users = append(users, userName)
	}

	return users, nil
}

// IsUserInRoom checks if a user with the given user name is in the chat room with the provided short code.
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

// BroadcastMessageToRoom broadcasts a message to all users in a chat room with the given short code.
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

// GetUserMessage retrieves a message from a user's message queue in a chat room.
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
