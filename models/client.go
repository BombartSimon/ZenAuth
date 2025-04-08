package models

type Client struct {
	ID           string
	Secret       string
	Name         string
	RedirectURIs []string
}
