package auth

type User struct {
	ID			 string
	Email		 string
	PasswordHash string
	Name 		 string
}

type RegisterRequest struct {
	Email	 string `json:"email"`
	Password string `json:"password"`
	Name	 string `json:"name"`
}

type LoginRequest struct {
	Email	   string `json:"email"`
	Password   string `json:"password"`
	DeviceName string `json:"device_name"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string	`json:"refresh_token"`
}
