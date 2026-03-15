# Golang Gin Framework - Development Rules

## Core Development Principles

### Test-Driven Development (TDD)
- Write tests BEFORE implementing functionality
- Follow Red-Green-Refactor cycle: write failing test → make it pass → refactor
- Aim for minimum 80% code coverage
- Use table-driven tests for multiple test cases
- Test files must be named `*_test.go` and placed alongside implementation files
- Use `testify` package for assertions and mocking
- Mock external dependencies using interfaces

### Test Naming Conventions (Assertive Style)
Tests should clearly state what behavior they verify using assertive language:

```go
// BAD: Vague test names
func TestCreateUser(t *testing.T)
func TestUserValidation(t *testing.T)
func TestGetUser(t *testing.T)

// GOOD: Assertive, behavior-focused names
func TestCreateUser_ShouldReturnCreatedUser_WhenValidDataProvided(t *testing.T)
func TestCreateUser_ShouldReturnValidationError_WhenEmailIsInvalid(t *testing.T)
func TestGetUser_ShouldReturnUser_WhenUserExists(t *testing.T)
func TestGetUser_ShouldReturnNotFoundError_WhenUserDoesNotExist(t *testing.T)
func TestUpdateUser_ShouldFailToUpdate_WhenUserIsUnauthorized(t *testing.T)
func TestDeleteUser_ShouldDeleteSuccessfully_WhenUserIsAdmin(t *testing.T)
```

**Pattern**: `Test[Function]_Should[ExpectedBehavior]_When[Condition]`

For table-driven tests:
```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name           string  // Use assertive description
        input          *User
        expectedError  error
        shouldSucceed  bool
    }{
        {
            name:          "should create user successfully when all required fields are provided",
            input:         validUser,
            shouldSucceed: true,
        },
        {
            name:          "should fail to create user when email is missing",
            input:         userWithoutEmail,
            expectedError: ErrEmailRequired,
            shouldSucceed: false,
        },
        {
            name:          "should fail to create user when email already exists",
            input:         duplicateEmailUser,
            expectedError: ErrEmailAlreadyExists,
            shouldSucceed: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Quality Over Coverage
- **Write meaningful tests that verify actual behavior**
- **DO NOT write tests just to increase coverage percentage**
- Focus on:
  - Business logic correctness
  - Error handling and edge cases
  - API contract validation
  - Database operations
  - Authentication and authorization
- Avoid:
  - Testing framework code
  - Testing getters/setters without logic
  - Duplicate tests that verify the same behavior
  - Tests that mock everything and test nothing
- Good test indicators:
  - Tests fail when behavior changes
  - Tests document expected behavior clearly
  - Tests catch real bugs during refactoring
  - Tests are readable and maintainable

### Self-Documenting Code (No Comments)
- **Code should be self-explanatory through clear naming**
- Use descriptive variable, function, and method names
- Extract complex logic into well-named functions
- Prefer explicit code over clever tricks
- Comments indicate unclear code that needs refactoring
- Only exception: Package documentation, complex algorithms, or critical business rules requiring context

```go
// BAD: Needs comments to understand
func ProcessData(d []byte) error {
    // Parse the data
    var u User
    json.Unmarshal(d, &u)
    
    // Check if valid
    if u.E == "" || u.P == "" {
        return errors.New("invalid")
    }
    
    // Hash password
    h := md5.Sum([]byte(u.P))
    u.P = string(h[:])
    
    return nil
}

// GOOD: Self-explanatory through naming
func CreateUserFromJSON(jsonData []byte) error {
    user, err := parseUserFromJSON(jsonData)
    if err != nil {
        return fmt.Errorf("failed to parse user data: %w", err)
    }
    
    if err := validateUserCredentials(user); err != nil {
        return err
    }
    
    hashedPassword := hashPassword(user.Password)
    user.Password = hashedPassword
    
    return nil
}

func parseUserFromJSON(jsonData []byte) (*User, error) {
    var user User
    if err := json.Unmarshal(jsonData, &user); err != nil {
        return nil, err
    }
    return &user, nil
}

func validateUserCredentials(user *User) error {
    if user.Email == "" {
        return ErrEmailRequired
    }
    if user.Password == "" {
        return ErrPasswordRequired
    }
    return nil
}

func hashPassword(plainPassword string) string {
    hashedBytes := md5.Sum([]byte(plainPassword))
    return string(hashedBytes[:])
}
```

### Dependency Management
- Wrap ALL third-party libraries in adapters/wrappers
- Define interfaces for external dependencies (databases, HTTP clients, cache, etc.)
- Never import third-party packages directly in business logic
- Place wrappers in dedicated packages (e.g., `internal/adapters/`, `internal/infrastructure/`)
- Example structure:
  ```
  internal/
    adapters/
      database/
        postgres_adapter.go
      cache/
        redis_adapter.go
      http/
        client_adapter.go
  ```

## Clean Code Principles

### Functions and Methods
- **Small and Focused**: Functions should do ONE thing well (Single Responsibility)
- **Maximum 20-30 lines**: If longer, extract into smaller functions
- **Maximum 3-4 parameters**: Use structs for more parameters
- **One level of abstraction**: Don't mix high-level and low-level operations
- **Avoid side effects**: Functions should be predictable

```go
// BAD: Too long, multiple responsibilities
func HandleUserRegistration(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    if req.Email == "" || !strings.Contains(req.Email, "@") {
        c.JSON(400, gin.H{"error": "invalid email"})
        return
    }
    
    if len(req.Password) < 8 {
        c.JSON(400, gin.H{"error": "password too short"})
        return
    }
    
    hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
    
    user := User{
        Email:    req.Email,
        Password: string(hash),
    }
    
    db.Create(&user)
    
    c.JSON(201, user)
}

// GOOD: Small, focused, clear responsibilities
func HandleUserRegistration(c *gin.Context) {
    registrationRequest, err := parseRegistrationRequest(c)
    if err != nil {
        respondWithBadRequest(c, err)
        return
    }
    
    if err := validateRegistrationRequest(registrationRequest); err != nil {
        respondWithValidationError(c, err)
        return
    }
    
    createdUser, err := h.userService.RegisterUser(registrationRequest)
    if err != nil {
        respondWithServiceError(c, err)
        return
    }
    
    respondWithCreatedUser(c, createdUser)
}
```

### Variable Naming
- **Descriptive names**: `userRepository` not `ur`, `totalAmount` not `ta`
- **Boolean prefixes**: Use `is`, `has`, `should`, `can` (`isActive`, `hasPermission`)
- **Avoid abbreviations**: Unless universally known (`ID`, `URL`, `HTTP`)
- **Context matters**: Short names OK in small scopes
- **Constants**: Use UPPER_SNAKE_CASE for exported, camelCase for unexported

```go
// BAD: Unclear naming
func Process(u *U, d int) error {
    t := time.Now().Unix()
    if t - u.T > d {
        return errors.New("expired")
    }
    return nil
}

// GOOD: Clear, self-documenting
func ValidateUserSessionTimeout(user *User, maxDurationSeconds int) error {
    currentTimestamp := time.Now().Unix()
    sessionDuration := currentTimestamp - user.LastActivityTimestamp
    
    if sessionDuration > int64(maxDurationSeconds) {
        return ErrSessionExpired
    }
    
    return nil
}
```

### Error Handling
- **Always check errors**: Never ignore returned errors
- **Wrap errors with context**: Use `fmt.Errorf("context: %w", err)`
- **Create custom error types**: For domain-specific errors
- **Use sentinel errors**: For expected error conditions
- **Descriptive error messages**: Include relevant context

```go
// Domain errors
var (
    ErrUserNotFound       = errors.New("user not found")
    ErrEmailAlreadyExists = errors.New("email already exists")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrUnauthorized       = errors.New("unauthorized access")
)

// Custom error types
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// Wrapping errors with context
func (s *UserService) GetUserByID(userID string) (*User, error) {
    user, err := s.repository.FindByID(userID)
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve user %s: %w", userID, err)
    }
    return user, nil
}
```

### Code Organization
- **Package by feature**: Not by layer (avoid `models/`, `controllers/` packages)
- **Small files**: 200-300 lines max per file
- **Clear separation**: Domain, use case, delivery, infrastructure
- **Minimize dependencies**: Avoid circular dependencies

## RESTful API Conventions

### HTTP Methods and Semantics
```go
// GET - Retrieve resources (Safe, Idempotent)
router.GET("/users", userHandler.ListUsers)           // List all users
router.GET("/users/:id", userHandler.GetUser)         // Get single user

// POST - Create new resources (Not idempotent)
router.POST("/users", userHandler.CreateUser)         // Create new user

// PUT - Update/Replace entire resource (Idempotent)
router.PUT("/users/:id", userHandler.UpdateUser)      // Replace entire user

// PATCH - Partial update (Idempotent)
router.PATCH("/users/:id", userHandler.PatchUser)     // Update specific fields

// DELETE - Remove resource (Idempotent)
router.DELETE("/users/:id", userHandler.DeleteUser)   // Delete user
```

### URL Naming Conventions
- **Use nouns, not verbs**: `/users` not `/getUsers`
- **Plural resource names**: `/users` not `/user`
- **Kebab-case for multi-word**: `/user-profiles` not `/userProfiles`
- **Nested resources**: `/users/:id/orders` for related resources
- **No trailing slashes**: `/users` not `/users/`
- **Lowercase only**: `/users` not `/Users`

```go
// BAD: Non-RESTful URLs
router.POST("/createUser", handler.CreateUser)
router.GET("/getUser/:id", handler.GetUser)
router.POST("/user/delete/:id", handler.DeleteUser)
router.GET("/UserOrders/:userId", handler.GetOrders)

// GOOD: RESTful URLs
router.POST("/users", handler.CreateUser)
router.GET("/users/:id", handler.GetUser)
router.DELETE("/users/:id", handler.DeleteUser)
router.GET("/users/:id/orders", handler.GetUserOrders)
```

### HTTP Status Codes
```go
// Success codes
200 OK              // GET, PUT, PATCH successful
201 Created         // POST successful (include Location header)
204 No Content      // DELETE successful, PATCH with no body

// Client error codes
400 Bad Request     // Invalid request body/parameters
401 Unauthorized    // Missing or invalid authentication
403 Forbidden       // Authenticated but not authorized
404 Not Found       // Resource doesn't exist
409 Conflict        // Resource conflict (duplicate email, etc.)
422 Unprocessable Entity // Validation errors

// Server error codes
500 Internal Server Error // Unexpected server error
503 Service Unavailable   // Temporary server issue
```

### Response Structure
```go
// Success response with data
type SuccessResponse struct {
    Data    interface{} `json:"data"`
    Message string      `json:"message,omitempty"`
}

// Error response
type ErrorResponse struct {
    Error   string                 `json:"error"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// Pagination metadata
type PaginationMetadata struct {
    CurrentPage int `json:"current_page"`
    PerPage     int `json:"per_page"`
    TotalPages  int `json:"total_pages"`
    TotalCount  int `json:"total_count"`
}

// List response with pagination
type ListResponse struct {
    Data       interface{}        `json:"data"`
    Pagination PaginationMetadata `json:"pagination"`
}

// Example handler responses
func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    user, err := h.userService.GetUserByID(userID)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            c.JSON(http.StatusNotFound, ErrorResponse{
                Error:   "not_found",
                Message: "User not found",
            })
            return
        }
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error:   "internal_error",
            Message: "Failed to retrieve user",
        })
        return
    }
    
    c.JSON(http.StatusOK, SuccessResponse{
        Data: user,
    })
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    var createRequest CreateUserRequest
    
    if err := c.ShouldBindJSON(&createRequest); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Error:   "invalid_request",
            Message: "Invalid request body",
            Details: map[string]interface{}{"validation_error": err.Error()},
        })
        return
    }
    
    createdUser, err := h.userService.CreateUser(createRequest)
    if err != nil {
        if errors.Is(err, ErrEmailAlreadyExists) {
            c.JSON(http.StatusConflict, ErrorResponse{
                Error:   "conflict",
                Message: "User with this email already exists",
            })
            return
        }
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error:   "internal_error",
            Message: "Failed to create user",
        })
        return
    }
    
    c.Header("Location", fmt.Sprintf("/users/%s", createdUser.ID))
    c.JSON(http.StatusCreated, SuccessResponse{
        Data:    createdUser,
        Message: "User created successfully",
    })
}
```

### API Versioning
```go
// URL path versioning (recommended)
v1 := router.Group("/api/v1")
{
    v1.GET("/users", userHandler.ListUsers)
    v1.POST("/users", userHandler.CreateUser)
}

v2 := router.Group("/api/v2")
{
    v2.GET("/users", userHandlerV2.ListUsers)
    v2.POST("/users", userHandlerV2.CreateUser)
}
```

### Request Validation
```go
type CreateUserRequest struct {
    Email     string `json:"email" binding:"required,email"`
    Password  string `json:"password" binding:"required,min=8"`
    FirstName string `json:"first_name" binding:"required"`
    LastName  string `json:"last_name" binding:"required"`
    Age       int    `json:"age" binding:"required,min=18,max=120"`
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    var createRequest CreateUserRequest
    
    if err := c.ShouldBindJSON(&createRequest); err != nil {
        respondWithValidationError(c, err)
        return
    }
    
    // Additional business validation
    if err := h.validateBusinessRules(createRequest); err != nil {
        respondWithValidationError(c, err)
        return
    }
    
    // Continue processing...
}
```

### Query Parameters and Filtering
```go
// Pagination
router.GET("/users?page=1&per_page=20", handler.ListUsers)

// Filtering
router.GET("/users?status=active&role=admin", handler.ListUsers)

// Sorting
router.GET("/users?sort=created_at&order=desc", handler.ListUsers)

// Search
router.GET("/users?search=john", handler.ListUsers)

// Implementation
func (h *UserHandler) ListUsers(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
    status := c.Query("status")
    sortBy := c.DefaultQuery("sort", "created_at")
    order := c.DefaultQuery("order", "desc")
    
    filters := domain.UserFilters{
        Page:    page,
        PerPage: perPage,
        Status:  status,
        SortBy:  sortBy,
        Order:   order,
    }
    
    users, pagination, err := h.userService.ListUsers(filters)
    // Handle response...
}
```

## Design Patterns & Architecture

### Clean Architecture Layers
```
myapp/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── domain/           # Entities + interfaces (no external dependencies)
│   ├── usecase/          # Business logic
│   ├── repository/       # Data access interfaces
│   ├── delivery/
│   │   └── http/         # Gin handlers
│   ├── infrastructure/   # External implementations
│   └── middleware/
├── pkg/                  # Shared utilities
├── test/                 # Integration tests
├── go.mod
└── go.sum
```

### SOLID Principles
- **Single Responsibility**: Each struct/function has one reason to change
- **Open/Closed**: Extend behavior via interfaces, not modification
- **Liskov Substitution**: Implementations must be interchangeable
- **Interface Segregation**: Small, focused interfaces (prefer many small over one large)
- **Dependency Inversion**: Depend on abstractions (interfaces), not concrete implementations

### Design Patterns to Apply
- **Repository Pattern**: For data access abstraction
- **Factory Pattern**: For complex object creation
- **Strategy Pattern**: For interchangeable algorithms
- **Dependency Injection**: Constructor injection via interfaces
- **Adapter Pattern**: For wrapping third-party libraries

## Code Quality Standards

### DRY (Don't Repeat Yourself)
- Extract repeated logic into functions/methods
- Use generic functions where appropriate (Go 1.18+)
- Create reusable utilities in `pkg/` or `internal/utils/`
- Avoid copy-paste code; refactor instead

### KISS (Keep It Simple, Stupid)
- Prefer simple, straightforward solutions
- Avoid premature optimization
- Use clear variable and function names
- Keep functions small (ideally < 50 lines)
- One level of abstraction per function

### YAGNI (You Aren't Gonna Need It)
- Implement only current requirements
- No speculative features or abstractions
- Add complexity only when needed
- Refactor when patterns emerge, not before

## Golang-Specific Guidelines

### Code Style
- Follow official Go conventions and `gofmt` formatting
- Use `golangci-lint` with strict configuration
- Run `go vet` before committing
- Error handling: always check errors, never ignore
- Use context.Context for cancellation and deadlines

### Naming Conventions
- Use camelCase for unexported, PascalCase for exported
- Interfaces: `Reader`, `Writer`, `Handler` (noun or verb+er)
- Avoid stuttering: `user.UserService` → `user.Service`
- Package names: short, lowercase, no underscores

### Handler Structure
```go
// handlers/user_handler.go
type UserHandler struct {
    userUseCase domain.UserUseCase // Interface, not concrete
}

func NewUserHandler(uc domain.UserUseCase) *UserHandler {
    return &UserHandler{userUseCase: uc}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    // Handle request
}
```

### Middleware Usage
- Create reusable middleware in `internal/middleware/`
- Authentication, logging, error handling as middleware
- Keep middleware focused and composable

## Incremental Development

### Workflow
1. Write interface and domain models first
2. Write tests for the interface contract
3. Implement minimal working solution
4. Refactor while tests remain green
5. Add more tests for edge cases
6. Iterate in small, deployable increments

### Git Practices
- Commit after each green test cycle
- Small, focused commits with clear messages
- Feature branches with descriptive names
- Code review before merging

## Checklist Before Committing
- [ ] All tests pass (`go test ./...`)
- [ ] Test names are assertive and behavior-focused
- [ ] Code coverage meets threshold (meaningful tests)
- [ ] No linting errors (`golangci-lint run`)
- [ ] No unnecessary comments (code is self-explanatory)
- [ ] Dependencies properly wrapped
- [ ] Interfaces used for dependencies
- [ ] No duplicated code
- [ ] Functions are small and focused (<30 lines)
- [ ] Proper error handling with context
- [ ] RESTful conventions followed
- [ ] HTTP status codes used correctly
- [ ] Clean code principles applied
- [ ] Documentation for exported functions
- [ ] Tests follow table-driven approach

## Dependencies to Consider
```go
// Testing
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/mock"

// Validation
"github.com/go-playground/validator/v10"

// Configuration
"github.com/spf13/viper"

// Database (wrapped)
"gorm.io/gorm" // Behind repository interface
```

## Remember
- **Tests first, always (with assertive naming)**
- **Self-documenting code, no comments**
- **Interfaces over concrete types**
- **Small, incremental changes**
- **Simplicity over cleverness**
- **RESTful conventions strictly**
- **Clean code principles always**
- **Refactor continuously**
- **Keep the complete application abstract everything looks like pulg and play**