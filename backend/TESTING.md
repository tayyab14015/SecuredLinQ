# Testing Guide

This document provides instructions on how to run tests for the SecuredLinQ backend project.

## Prerequisites

- Go 1.21 or higher
- All project dependencies installed (`go mod download`)

## Test Structure

The test files are organized alongside the source code they test:

```
backend/
├── internal/
│   ├── service/
│   │   ├── auth_service.go
│   │   ├── auth_service_test.go
│   │   ├── driver_service.go
│   │   ├── driver_service_test.go
│   │   ├── meeting_service_test.go
│   │   └── ...
│   ├── handler/
│   │   ├── auth_handler.go
│   │   ├── auth_handler_test.go
│   │   └── ...
│   └── ...
```

## Running Tests

### Run All Tests

To run all tests in the project:

```bash
cd backend
go test ./...
```

### Run Tests with Verbose Output

To see detailed output for each test:

```bash
go test -v ./...
```

### Run Tests for a Specific Package

To run tests only for a specific package:

```bash
# Run auth service tests
go test ./internal/service -v

# Run handler tests
go test ./internal/handler -v
```

### Run a Specific Test

To run a specific test function:

```bash
go test ./internal/service -v -run TestValidateAdminCredentials
```

### Run Tests with Coverage

To generate a coverage report:

```bash
# Generate coverage report
go test ./... -cover

# Generate detailed coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Tests in Parallel

To run tests in parallel (faster execution):

```bash
go test ./... -parallel 4
```

## Test Categories

### Unit Tests

Unit tests test individual functions and methods in isolation using mocks:

- **Auth Service Tests** (`auth_service_test.go`):
  - `TestValidateAdminCredentials` - Tests admin credential validation
  - `TestValidateDriverCredentials` - Tests driver credential validation
  - `TestCreateSession` - Tests session creation
  - `TestValidateSession` - Tests session validation
  - `TestHashPassword` - Tests password hashing
  - `TestCheckPasswordHash` - Tests password verification

- **Driver Service Tests** (`driver_service_test.go`):
  - `TestRegisterDriver` - Tests driver registration
  - `TestValidateCredentials` - Tests credential validation
  - `TestGetDriverByID` - Tests driver retrieval
  - `TestGetAllDrivers` - Tests driver listing
  - `TestDeactivateDriver` - Tests driver deactivation
  - `TestActivateDriver` - Tests driver activation

- **Meeting Service Tests** (`meeting_service_test.go`):
  - `TestGetOrCreateMeetingRoom` - Tests meeting room creation/retrieval
  - `TestGetMeetingRoomByRoomID` - Tests meeting room lookup
  - `TestUpdateLastJoined` - Tests last joined timestamp update
  - `TestEndMeeting` - Tests meeting termination

### Handler Tests

Handler tests test HTTP endpoints using the Gin test framework:

- **Email Handler Tests** (`email_handler_test.go`):
  - `TestSendMeetingLink` - Tests meeting link email sending endpoint
  - `TestEmailContentGeneration` - Tests email content generation
  - `TestEmailSubjectGeneration` - Tests email subject line generation
  - `TestEmailHandlerInitialization` - Tests email handler creation
  - `TestEmailRequestValidation` - Tests request validation
  - `TestEmailContentContainsRequiredFields` - Tests email content includes all required fields
  - `TestVerifyEmailContentHelper` - Tests email content verification helper

## Writing New Tests

### Example: Service Test

```go
func TestMyFunction(t *testing.T) {
    // Setup
    mockRepo := new(MockRepository)
    service := NewMyService(mockRepo)
    
    // Test
    result, err := service.MyFunction()
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockRepo.AssertExpectations(t)
}
```

### Example: Handler Test

```go
func TestMyHandler(t *testing.T) {
    // Setup
    router := gin.New()
    handler := NewMyHandler(service)
    router.POST("/endpoint", handler.MyHandler)
    
    // Test
    body, _ := json.Marshal(requestData)
    req := httptest.NewRequest(http.MethodPost, "/endpoint", bytes.NewBuffer(body))
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
}
```

## Mock Objects

The tests use the `testify/mock` package for creating mock objects. Mocks are defined in each test file and implement the interfaces they mock.

## Best Practices

1. **Test Isolation**: Each test should be independent and not rely on other tests
2. **Clear Test Names**: Use descriptive test names that explain what is being tested
3. **Arrange-Act-Assert**: Structure tests with clear setup, execution, and assertion phases
4. **Mock Expectations**: Always verify that mock expectations were met using `AssertExpectations(t)`
5. **Error Cases**: Test both success and error scenarios
6. **Edge Cases**: Test boundary conditions and edge cases

## Continuous Integration

Tests should be run as part of CI/CD pipeline:

```yaml
# Example GitHub Actions workflow
- name: Run tests
  run: |
    cd backend
    go test ./... -v -cover
```

## Troubleshooting

### Tests Fail with "package not found"

Ensure all dependencies are installed:

```bash
go mod download
go mod tidy
```

### Tests Fail with Database Errors

Unit tests use mocks and should not require a database. If you see database errors, check that mocks are properly set up.

### Tests Timeout

If tests timeout, check for:
- Infinite loops in test code
- Blocking operations without timeouts
- Deadlocks in concurrent tests

## Test Coverage Goals

Aim for:
- **Service Layer**: 80%+ coverage
- **Handler Layer**: 70%+ coverage
- **Repository Layer**: 60%+ coverage (can be lower if using integration tests)

## Email Testing Notes

The email handler tests (`email_handler_test.go`) test the HTTP endpoint and request validation. Since the email functionality uses SMTP, full integration tests would require:

1. **Mock SMTP Server**: Use a library like `github.com/emersion/go-smtp` to create a test SMTP server
2. **Test SMTP Configuration**: Configure the handler to use the test SMTP server
3. **Verify Email Content**: Capture and verify the actual email content sent

Example for integration testing with mock SMTP:

```go
// This would require additional setup with a mock SMTP server
func TestSendEmailIntegration(t *testing.T) {
    // Setup mock SMTP server
    // Configure handler to use mock server
    // Send email
    // Verify email content was sent correctly
}
```

For now, the tests focus on:
- Request validation
- Configuration validation
- Error handling
- Email content structure verification

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Gin Testing Guide](https://gin-gonic.com/docs/testing/)
- [Go SMTP Package](https://pkg.go.dev/net/smtp)
