package models

import "time"

type User struct {
	UUID        string     `json:"uuid,omitempty"`
	Name        string     `json:"name,omitempty"`
	Surname     string     `json:"surname,omitempty"`
	Password    string     `json:"password,omitempty"`
	Username    string     `json:"username,omitempty"`
	PassHash    []byte     `json:"pass_hash,omitempty"`
	PhoneNumber string     `json:"phone_number,omitempty"`
	Email       string     `json:"email,omitempty"`
	Role        string     `json:"role,omitempty"`
	GroupID     int64      `json:"group_id ,omitempty"`
	ImageS3Path string     `json:"image_s3_path,omitempty"`
	IsBlocked   bool       `json:"is_blocked,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	ModifiedAt  *time.Time `json:"modified_at,omitempty"`
}
