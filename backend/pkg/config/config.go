package config

// Config - Configuration
type Config struct {
	ChannelSecret      string `config:"CHANNEL_SECRET"`       // Channel Secret
	ChannelAccessToken string `config:"CHANNEL_ACCESS_TOKEN"` // Channel Access Token
}
