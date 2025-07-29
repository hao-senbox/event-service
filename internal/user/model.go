package user

type UserInfor struct {
	UserID     string     `json:"user_id"`
	UserName   string     `json:"user_name"`
	FullName   string     `json:"full_name"`
	Role       string     `json:"role"`
	Avartar    string     `json:"avatar"`
}