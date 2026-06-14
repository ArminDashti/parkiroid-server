package auth

// Single admin account for the web app. Change the hash after generating a new
// bcrypt digest for your production password (golang.org/x/crypto/bcrypt).
const (
	AdminUsername = "admin"

	// bcrypt hash of "parkiroid-dev-password"
	AdminPasswordHash = "$2a$10$71UUnN8dTXlP3T3i/ni0Ve3fK8kukrXCciZufunqcBAqWyOvUXLDO"
)
