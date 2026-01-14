# Test Commands Reference

This document provides commands to run each individual test case in the project.

## Prerequisites

Make sure you're in the `backend` directory:
```bash
cd backend
```

## Auth Service Tests

Location: `internal/service/auth_service_test.go`

### Run all auth service tests:
```bash
go test ./internal/service -v -run TestValidateAdminCredentials|TestHashPassword|TestCheckPasswordHash
```

### Run individual tests:

#### TestValidateAdminCredentials
```bash
go test ./internal/service -v -run TestValidateAdminCredentials
```

#### TestHashPassword
```bash
go test ./internal/service -v -run TestHashPassword
```

#### TestCheckPasswordHash
```bash
go test ./internal/service -v -run TestCheckPasswordHash
```

### Run specific sub-tests (table-driven test cases):

#### TestValidateAdminCredentials - Valid credentials
```bash
go test ./internal/service -v -run "TestValidateAdminCredentials/Valid credentials"
```

#### TestValidateAdminCredentials - Invalid username
```bash
go test ./internal/service -v -run "TestValidateAdminCredentials/Invalid username"
```

#### TestValidateAdminCredentials - Invalid password
```bash
go test ./internal/service -v -run "TestValidateAdminCredentials/Invalid password"
```

#### TestCheckPasswordHash - Correct password
```bash
go test ./internal/service -v -run "TestCheckPasswordHash/Correct password"
```

#### TestCheckPasswordHash - Wrong password
```bash
go test ./internal/service -v -run "TestCheckPasswordHash/Wrong password"
```

---

## Meeting Service Tests

Location: `internal/service/meeting_service_test.go`

### Run all meeting service tests:
```bash
go test ./internal/service -v -run TestGenerateRoomID
```

### Run individual test:

#### TestGenerateRoomID
```bash
go test ./internal/service -v -run TestGenerateRoomID
```

---

## Email Handler Tests

Location: `internal/handler/email_handler_test.go`

### Run all email handler tests:
```bash
go test ./internal/handler -v -run TestSendMeetingLink|TestEmailContentGeneration|TestEmailSubjectGeneration|TestEmailHandlerInitialization|TestVerifyEmailContentHelper
```

### Run individual tests:

#### TestSendMeetingLink
```bash
go test ./internal/handler -v -run TestSendMeetingLink
```

#### TestEmailContentGeneration
```bash
go test ./internal/handler -v -run TestEmailContentGeneration
```

#### TestEmailSubjectGeneration
```bash
go test ./internal/handler -v -run TestEmailSubjectGeneration
```

#### TestEmailHandlerInitialization
```bash
go test ./internal/handler -v -run TestEmailHandlerInitialization
```

#### TestVerifyEmailContentHelper
```bash
go test ./internal/handler -v -run TestVerifyEmailContentHelper
```

### Run specific sub-tests (table-driven test cases):

#### TestSendMeetingLink - Successful email send
```bash
go test ./internal/handler -v -run "TestSendMeetingLink/Successful email send"
```

#### TestSendMeetingLink - Missing driver email
```bash
go test ./internal/handler -v -run "TestSendMeetingLink/Missing driver email"
```

#### TestSendMeetingLink - Missing driver name
```bash
go test ./internal/handler -v -run "TestSendMeetingLink/Missing driver name"
```

#### TestSendMeetingLink - Missing meeting link
```bash
go test ./internal/handler -v -run "TestSendMeetingLink/Missing meeting link"
```

#### TestSendMeetingLink - Missing load number
```bash
go test ./internal/handler -v -run "TestSendMeetingLink/Missing load number"
```

#### TestSendMeetingLink - Email not configured (missing sender email)
```bash
go test ./internal/handler -v -run "TestSendMeetingLink/Email not configured - missing sender email"
```

#### TestSendMeetingLink - Email not configured (missing app password)
```bash
go test ./internal/handler -v -run "TestSendMeetingLink/Email not configured - missing app password"
```

#### TestEmailSubjectGeneration - Standard load number
```bash
go test ./internal/handler -v -run "TestEmailSubjectGeneration/Standard load number"
```

#### TestEmailSubjectGeneration - Long load number
```bash
go test ./internal/handler -v -run "TestEmailSubjectGeneration/Long load number"
```

---

## Run All Tests

### Run all tests in the project:
```bash
go test ./... -v
```

### Run all tests in a specific package:
```bash
# All service tests
go test ./internal/service -v

# All handler tests
go test ./internal/handler -v
```

### Run all tests with coverage:
```bash
go test ./... -v -cover
```

### Run all tests and generate coverage report:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Quick Reference

### Pattern for running tests:
```bash
go test <package_path> -v -run <TestFunctionName>
```

### Pattern for running sub-tests:
```bash
go test <package_path> -v -run "<TestFunctionName>/<SubTestName>"
```
Note: Use quotes around the pattern when sub-test names contain spaces.

### Examples:
```bash
# Run a specific test function
go test ./internal/service -v -run TestHashPassword

# Run a specific sub-test
go test ./internal/service -v -run TestValidateAdminCredentials/Valid_credentials

# Run multiple test functions (use | to separate)
go test ./internal/service -v -run "TestHashPassword|TestCheckPasswordHash"

# Run all tests matching a pattern
go test ./internal/handler -v -run TestEmail
```

---

## Notes

1. **Sub-test names**: When running sub-tests from table-driven tests, use the exact name from the test struct's `name` field. Use quotes around the pattern if the name contains spaces.

2. **Case sensitivity**: Test function names are case-sensitive. Use exact capitalization.

3. **Verbose output**: The `-v` flag provides detailed output showing which tests pass or fail.

4. **Test patterns**: The `-run` flag accepts a regular expression pattern, so you can use patterns like `TestEmail.*` to run all email-related tests.

5. **Windows PowerShell**: If using PowerShell on Windows, you may need to quote the test pattern:
   ```powershell
   go test ./internal/service -v -run "TestValidateAdminCredentials"
   ```

---

## Troubleshooting

### Test not found?
- Make sure you're in the `backend` directory
- Check that the test file exists and the function name is correct
- Verify the package path is correct (e.g., `./internal/service`)

### Sub-test not found?
- Check the exact `name` field in the test struct
- Replace spaces with underscores in the sub-test name
- Use quotes around the pattern if it contains special characters

### Tests fail with import errors?
- Run `go mod download` to ensure all dependencies are installed
- Run `go mod tidy` to clean up dependencies
