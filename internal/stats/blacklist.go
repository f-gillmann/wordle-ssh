package stats

// BlacklistedUsernames contains usernames that should never record stats
var BlacklistedUsernames = map[string]bool{
	"anonymous": true,
	"root":      true,
	"toor":      true,
	"admin":     true,
	"user":      true,
	"guest":     true,
	"test":      true,
	"demo":      true,
	"ubuntu":    true,
	"debian":    true,
	"centos":    true,
	"fedora":    true,
	"oracle":    true,
	"pi":        true,
	"vagrant":   true,
	"default":   true,
	"1234":      true,
	"ftp":       true,
}

// IsBlacklisted checks if a username is blacklisted
func IsBlacklisted(username string) bool {
	return BlacklistedUsernames[username]
}
