package main

import (
	"testing"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		expectErr bool
	}{
		{
			name: "valid configuration mainnet",
			config: Config{
				PeerAddrs: []Addr{
					{IP: "192.168.1.1", Port: 8333},
					{IP: "10.0.0.1", Port: 18333},
				},
				Network: "mainnet",
			},
			expectErr: false,
		},
		{
			name: "valid configuration simnet",
			config: Config{
				PeerAddrs: []Addr{
					{IP: "192.168.1.1", Port: 8333},
				},
				Network: "simnet",
			},
			expectErr: false,
		},
		{
			name: "invalid IP address",
			config: Config{
				PeerAddrs: []Addr{
					{IP: "invalid_ip", Port: 8333},
				},
				Network: "mainnet",
			},
			expectErr: true,
		},
		{
			name: "invalid port number (too low)",
			config: Config{
				PeerAddrs: []Addr{
					{IP: "192.168.1.1", Port: -1},
				},
				Network: "mainnet",
			},
			expectErr: true,
		},
		{
			name: "invalid port number (too high)",
			config: Config{
				PeerAddrs: []Addr{
					{IP: "192.168.1.1", Port: 70000},
				},
				Network: "mainnet",
			},
			expectErr: true,
		},
		{
			name: "invalid network",
			config: Config{
				PeerAddrs: []Addr{
					{IP: "192.168.1.1", Port: 8333},
				},
				Network: "invalidnet",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("Validate() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}
