package jwt

import (
	"testing"
	"user-management-service/internal/config"
	"user-management-service/internal/models"
)

func TestNewAccessToken(t *testing.T) {
	type args struct {
		user *models.User
		cfg  config.Token
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAccessToken(tt.args.user, tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NewAccessToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
