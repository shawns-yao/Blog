package adminuser

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/shawns-yao/shawn-blog/server/internal/domain/identity"
)

var ErrLastAdminMutation = errors.New("至少保留一个可用管理员账号")
var ErrSelfAdminMutation = errors.New("当前登录用户不能修改自己的状态")

type InputError struct{ Message string }

func (e *InputError) Error() string { return e.Message }

type Service struct {
	repo identity.Repository
}

func NewService(repo identity.Repository) *Service {
	return &Service{repo: repo}
}

type ListUsersCmd struct {
	Keyword    string
	OnlyAdmin  *bool
	OnlyActive *bool
	Page       int
	PageSize   int
}

type UpdateUserCmd struct {
	OperatorID int64
	UserID     int64
	Nickname   string
	Email      string
	IsActive   bool
	IsAdmin    bool
}

type CreateUserCmd struct {
	Username string
	Nickname string
	Email    string
	Password string
	IsAdmin  bool
}

func (s *Service) CreateUser(ctx context.Context, cmd CreateUserCmd) (*identity.User, error) {
	username := strings.TrimSpace(cmd.Username)
	if username == "" || utf8.RuneCountInString(username) > 45 || strings.ContainsFunc(username, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) {
		return nil, &InputError{"账号不能为空，最多45个字符，不能包含空格"}
	}
	if utf8.RuneCountInString(cmd.Password) < 8 || len(cmd.Password) > 72 {
		return nil, &InputError{"密码至少8个字符，过长的密码请缩短"}
	}
	nickname := strings.TrimSpace(cmd.Nickname)
	if nickname == "" {
		nickname = username
	}
	if utf8.RuneCountInString(nickname) > 45 {
		return nil, &InputError{"昵称最多45个字符"}
	}
	email := strings.TrimSpace(cmd.Email)
	if email != "" {
		address, err := mail.ParseAddress(email)
		if err != nil || address.Address != email || len(email) > 255 {
			return nil, &InputError{"请输入有效的邮箱地址"}
		}
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &identity.User{
		Username: username, Nickname: nickname, Email: email,
		Password: string(hashed), IsActive: true, IsAdmin: cmd.IsAdmin,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

func (s *Service) ListUsers(ctx context.Context, cmd ListUsersCmd) ([]identity.User, int64, error) {
	return s.repo.ListUsers(ctx, identity.UserListOptions{
		Keyword:    strings.TrimSpace(cmd.Keyword),
		OnlyAdmin:  cmd.OnlyAdmin,
		OnlyActive: cmd.OnlyActive,
		Page:       cmd.Page,
		PageSize:   cmd.PageSize,
	})
}

func (s *Service) UpdateUser(ctx context.Context, cmd UpdateUserCmd) (*identity.User, error) {
	if cmd.UserID <= 0 {
		return nil, identity.ErrUserNotFound
	}

	current, err := s.repo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	nickname := strings.TrimSpace(cmd.Nickname)
	if nickname == "" {
		nickname = current.Nickname
	}
	email := strings.TrimSpace(cmd.Email)

	if current.IsAdmin && (!cmd.IsActive || !cmd.IsAdmin) {
		activeAdmins, err := s.repo.CountActiveAdmins(ctx)
		if err != nil {
			return nil, err
		}
		if activeAdmins <= 1 {
			return nil, ErrLastAdminMutation
		}
	}

	if cmd.UserID == cmd.OperatorID && (cmd.IsActive != current.IsActive || cmd.IsAdmin != current.IsAdmin) {
		return nil, ErrSelfAdminMutation
	}

	updated, err := s.repo.UpdateAdminUser(ctx, cmd.UserID, nickname, email, cmd.IsActive, cmd.IsAdmin)
	if err != nil {
		return nil, err
	}
	updated.Password = ""
	return updated, nil
}
