package contract

type AdminUserListResp struct {
	Items []UserResp `json:"items"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

type UpdateAdminUserReq struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	IsActive bool   `json:"isActive"`
	IsAdmin  bool   `json:"isAdmin"`
}

type CreateAdminUserReq struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"isAdmin"`
}
