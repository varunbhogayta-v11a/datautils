package db

import (
	"errors"
	"sync"

	"github.com/varunbhogayta-v11a/datautils/pkg/models"
)

type MockUserRepository struct {
	mu     sync.RWMutex
	users  []models.User
	nextID uint
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  []models.User{},
		nextID: 1,
	}
}

func (r *MockUserRepository) Create(user *models.User) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if u.Email == user.Email {
			return nil, errors.New("user already exists")
		}
		if u.Username == user.Username {
			return nil, errors.New("user already exists")
		}
	}

	user.ID = r.nextID
	r.nextID++
	r.users = append(r.users, *user)
	return user, nil
}

func (r *MockUserRepository) GetByID(id uint) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Username == username {
			return &u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *MockUserRepository) Update(user *models.User) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, u := range r.users {
		if u.ID == user.ID {
			r.users[i] = *user
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *MockUserRepository) Delete(id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, u := range r.users {
		if u.ID == id {
			r.users = append(r.users[:i], r.users[i+1:]...)
			return nil
		}
	}
	return errors.New("user not found")
}

func (r *MockUserRepository) List() ([]models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]models.User, len(r.users))
	copy(users, r.users)
	return users, nil
}

func (r *MockUserRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users = []models.User{}
	r.nextID = 1
}

type MockOperationLogRepository struct {
	mu     sync.RWMutex
	logs   []models.OperationLog
	nextID uint
}

func NewMockOperationLogRepository() *MockOperationLogRepository {
	return &MockOperationLogRepository{
		logs:   []models.OperationLog{},
		nextID: 1,
	}
}

func (r *MockOperationLogRepository) Create(log *models.OperationLog) (*models.OperationLog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.ID = r.nextID
	r.nextID++
	r.logs = append(r.logs, *log)
	return log, nil
}

func (r *MockOperationLogRepository) GetByUserID(userID uint) ([]models.OperationLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []models.OperationLog
	for _, log := range r.logs {
		if log.UserID == userID {
			result = append(result, log)
		}
	}
	return result, nil
}

func (r *MockOperationLogRepository) List() ([]models.OperationLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	logs := make([]models.OperationLog, len(r.logs))
	copy(logs, r.logs)
	return logs, nil
}

func (r *MockOperationLogRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = []models.OperationLog{}
	r.nextID = 1
}

type MockRouteRepository struct {
	mu     sync.RWMutex
	routes []models.RouteConfig
	nextID uint
}

func NewMockRouteRepository() *MockRouteRepository {
	return &MockRouteRepository{
		routes: []models.RouteConfig{},
		nextID: 1,
	}
}

func (r *MockRouteRepository) Create(route *models.RouteConfig) (*models.RouteConfig, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	route.ID = r.nextID
	r.nextID++
	r.routes = append(r.routes, *route)
	return route, nil
}

func (r *MockRouteRepository) GetByID(id uint) (*models.RouteConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, route := range r.routes {
		if route.ID == id {
			return &route, nil
		}
	}
	return nil, errors.New("route not found")
}

func (r *MockRouteRepository) GetByPath(path string) (*models.RouteConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, route := range r.routes {
		if route.Path == path && route.Active {
			return &route, nil
		}
	}
	return nil, errors.New("route not found")
}

func (r *MockRouteRepository) Update(route *models.RouteConfig) (*models.RouteConfig, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, rt := range r.routes {
		if rt.ID == route.ID {
			r.routes[i] = *route
			return route, nil
		}
	}
	return nil, errors.New("route not found")
}

func (r *MockRouteRepository) Delete(id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, route := range r.routes {
		if route.ID == id {
			r.routes = append(r.routes[:i], r.routes[i+1:]...)
			return nil
		}
	}
	return errors.New("route not found")
}

func (r *MockRouteRepository) List() ([]models.RouteConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var active []models.RouteConfig
	for _, route := range r.routes {
		if route.Active {
			active = append(active, route)
		}
	}
	return active, nil
}

func (r *MockRouteRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes = []models.RouteConfig{}
	r.nextID = 1
}
