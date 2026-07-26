package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/consts"
	"github.com/OptLTD/member/internal/model"
	"github.com/OptLTD/member/internal/support"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	conn  *gorm.DB
	userS *UserService
}

type TokenInfo struct {
	UserID  uint64
	Expiry  time.Time
	IsAdmin bool
}

func NewAuthService() *AuthService {
	service := &AuthService{
		conn:  config.GetDB(),
		userS: NewUserService(),
	}

	go service.cleanupExpiredTokens()
	return service
}

func (s *AuthService) sessionToTokenInfo(session *model.SessionModel) *TokenInfo {
	return &TokenInfo{
		UserID:  session.UserID,
		Expiry:  session.ExpireAt,
		IsAdmin: session.IsAdmin,
	}
}

// SignIn 用户登录
func (s *AuthService) SignIn(username, password string) (string, *model.UserModel, error) {
	user, err := s.userS.GetUserByUsername(username)
	if err != nil {
		return "", nil, consts.ErrInvalidCredentials
	}

	if !user.IsActive {
		return "", nil, consts.ErrUserAccountDisabled
	}

	if !user.Verified {
		return "", nil, consts.ErrUserAccountNotVerified
	}

	if !s.userS.ValidatePassword(user.Password, password) {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}

	token, err := s.generateToken()
	if err != nil {
		return "", nil, err
	}

	tokenInfo := &TokenInfo{
		UserID:  user.ID,
		IsAdmin: user.UserRole == model.RoleAdmin,
		Expiry:  time.Now().Add(24 * time.Hour),
	}

	if err = s.SetToken(token, tokenInfo); err != nil {
		return "", nil, fmt.Errorf("%w: %v", consts.ErrTokenStorageFailed, err)
	}

	user.Password = ""
	return token, user, nil
}

// SignOut 用户登出
func (s *AuthService) SignOut(token string) {
	s.DeleteToken(token)
}

// SignUp 用户注册
func (s *AuthService) SignUp(req *model.SignUpRequest) (*model.UserModel, error) {
	existingUser, err := s.userS.GetUserByUsername(req.Username)
	if err == nil && existingUser != nil {
		return nil, consts.ErrUsernameAlreadyExists
	}

	existingUser, err = s.userS.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, consts.ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrPasswordEncryptionFailed, err)
	}

	apiKey, err := support.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("API Key 生成失败: %v", err)
	}

	setup := NewSetupService()
	limit, err := setup.GetPlanLimit(model.PlanBasic)
	if err != nil || limit == nil {
		limit = DefaultBasicLimit()
	}

	today := time.Now().Format("2006-01-02")
	user := &model.UserModel{
		Email: req.Email, Username: req.Username,
		APIKey: support.HashAPIKey(apiKey), Password: string(hashedPassword),
		UserRole: model.RoleUser, UserPlan: model.PlanBasic,
		IsActive: true, ApiLimit: limit,
		ApiUsage: &model.ApiUsage{DayStamp: today},
	}
	if err := s.userS.CreateUser(user); err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrUserCreationFailed, err)
	}
	if config.HasModule(config.ModuleReferral) && req.RefCode != "" {
		_ = NewReferralService().BindReferral(user.ID, req.RefCode)
	}
	if config.HasModule(config.ModuleAnalytics) {
		_ = NewAnalyticsService().TrackInternal(user.ID, "user.register", map[string]any{
			"refCode": req.RefCode,
		})
	}
	user.APIKey = apiKey
	return user, nil
}

func (s *AuthService) DeleteSignup(user *model.UserModel) error {
	return s.userS.DeleteUser(user.ID)
}

// VerifySignupCode 验证注册邮箱
func (s *AuthService) VerifySignupCode(code string) error {
	var reset model.VerifyModel
	if err := s.conn.Where("token = ?", code).First(&reset).Error; err != nil {
		return consts.ErrInvalidVerificationCode
	}
	if reset.ExpireAt.Before(time.Now()) {
		return consts.ErrVerificationCodeExpired
	}

	user, err := s.userS.GetUserByEmail(reset.Email)
	if err != nil || user == nil {
		return fmt.Errorf("用户不存在")
	}
	user.Verified = true
	if err := s.conn.Save(&user).Error; err != nil {
		return fmt.Errorf("%w: %v", consts.ErrEmailVerificationFailed, err)
	}
	if err := s.conn.Delete(&reset).Error; err != nil {
		return fmt.Errorf("%w: %v", consts.ErrVerificationCodeDeletionFailed, err)
	}
	return nil
}

func (s *AuthService) GenerateSignupCode(email string) (*model.VerifyModel, error) {
	reset := &model.VerifyModel{Email: email}
	token, err := s.generateToken()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrVerificationCodeGenerationFailed, err)
	}
	reset.Token = token
	reset.ExpireAt = time.Now().Add(10 * time.Minute)

	if res := s.conn.Create(reset); res.Error != nil {
		return nil, fmt.Errorf("验证码生成失败: %v", res.Error)
	}
	return reset, nil
}

// ValidateSigninToken 验证登录 token
func (s *AuthService) ValidateSigninToken(token string) (*TokenInfo, bool) {
	return s.GetToken(token)
}

// ValidateAPIKey 验证 API Key 并检查账户状态
func (s *AuthService) ValidateAPIKey(apiKey string) (*model.UserModel, error) {
	user, err := s.userS.GetUserByAPIKey(apiKey)
	if err != nil {
		return nil, consts.ErrInvalidAPIKey
	}
	if !user.IsActive {
		return nil, consts.ErrUserAccountDisabled
	}
	if !user.Verified {
		return nil, consts.ErrUserAccountNotVerified
	}
	if user.ExpireAt != nil && user.ExpireAt.Before(time.Now()) {
		return nil, consts.ErrAPIKeyAccountExpired
	}
	return user, nil
}

// GetUserFromSigninToken 从 token 获取用户信息
func (s *AuthService) GetUserFromSigninToken(token string) (*model.UserModel, error) {
	tokenInfo, valid := s.ValidateSigninToken(token)
	if !valid {
		return nil, consts.ErrInvalidToken
	}

	user, err := s.userS.GetUserByID(tokenInfo.UserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	user.Password = ""
	return user, nil
}

func (s *AuthService) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ForgotPassword 忘记密码
func (s *AuthService) ForgotPassword(email string) (*model.VerifyModel, error) {
	user, err := s.userS.GetUserByEmail(email)
	if err != nil {
		return nil, consts.ErrEmailNotRegistered
	}

	if !user.IsActive {
		return nil, fmt.Errorf("用户账户已被禁用")
	}

	reset := &model.VerifyModel{Email: email}
	token, err := s.generateToken()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrResetCodeGenerationFailed, err)
	}
	reset.Token = token
	reset.ExpireAt = time.Now().Add(30 * time.Minute)

	s.conn.Where("email = ?", email).Delete(&model.VerifyModel{})

	if res := s.conn.Create(reset); res.Error != nil {
		return nil, fmt.Errorf("重置码生成失败: %v", res.Error)
	}

	return reset, nil
}

// VerifyResetCode 验证重置码
func (s *AuthService) VerifyResetCode(code string) (*model.VerifyModel, error) {
	var reset model.VerifyModel
	if err := s.conn.Where("token = ?", code).First(&reset).Error; err != nil {
		return nil, consts.ErrInvalidResetCode
	}

	if reset.ExpireAt.Before(time.Now()) {
		return nil, consts.ErrResetCodeExpired
	}

	return &reset, nil
}

// ResetPassword 重置密码
func (s *AuthService) ResetPassword(code, newPassword string) error {
	reset, err := s.VerifyResetCode(code)
	if err != nil {
		return err
	}

	user, err := s.userS.GetUserByEmail(reset.Email)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%w: %v", consts.ErrPasswordEncryptionFailed, err)
	}

	user.Password = string(hashedPassword)
	if err := s.conn.Save(user).Error; err != nil {
		return fmt.Errorf("%w: %v", consts.ErrPasswordUpdateFailed, err)
	}

	s.conn.Delete(reset)
	s.deleteSessionsByUserID(user.ID)
	return nil
}

func (s *AuthService) deleteSessionsByUserID(userID uint64) {
	s.conn.Where("user_id = ?", userID).Delete(&model.SessionModel{})
}

func (s *AuthService) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.conn.Where("expire_at < ?", time.Now()).Delete(&model.SessionModel{})
	}
}

// SetToken 写入 session 到数据库
func (s *AuthService) SetToken(token string, tokenInfo *TokenInfo) error {
	if time.Now().After(tokenInfo.Expiry) {
		return consts.ErrTokenExpired
	}

	session := &model.SessionModel{
		Token:    token,
		UserID:   tokenInfo.UserID,
		IsAdmin:  tokenInfo.IsAdmin,
		ExpireAt: tokenInfo.Expiry,
	}
	if err := s.conn.Create(session).Error; err != nil {
		return fmt.Errorf("%w: %v", consts.ErrTokenStorageFailed, err)
	}
	return nil
}

// GetToken 从数据库读取 session
func (s *AuthService) GetToken(token string) (*TokenInfo, bool) {
	var session model.SessionModel
	if err := s.conn.Where("token = ?", token).First(&session).Error; err != nil {
		return nil, false
	}

	if time.Now().After(session.ExpireAt) {
		s.conn.Delete(&session)
		return nil, false
	}

	return s.sessionToTokenInfo(&session), true
}

// DeleteToken 删除 session
func (s *AuthService) DeleteToken(token string) {
	s.conn.Where("token = ?", token).Delete(&model.SessionModel{})
}

// GenerateCallbackToken 生成回调 token
func (s *AuthService) GenerateCallbackToken(user *model.UserModel) (string, error) {
	token, err := s.generateToken()
	if err != nil {
		return "", err
	}

	tokenInfo := &TokenInfo{
		UserID:  user.ID,
		IsAdmin: user.UserRole == model.RoleAdmin,
		Expiry:  time.Now().Add(5 * time.Minute),
	}

	if err = s.SetToken(token, tokenInfo); err != nil {
		return "", fmt.Errorf("%w: %v", consts.ErrTokenGenerationFailed, err)
	}

	return token, nil
}

// GenerateCallbackSign 使用 APP_SECRET 做 HMAC-SHA256 签名
func (s *AuthService) GenerateCallbackSign(token, email, suff string) (string, error) {
	secret := os.Getenv("APP_SECRET")
	if secret == "" {
		return "", consts.ErrAppSecretNotConfigured
	}
	data := fmt.Sprintf("%s:%s:%s", token, email, suff)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return fmt.Sprintf("sn-%x", mac.Sum(nil)[:16]), nil
}

// ValidateCallbackToken 验证回调 token
func (s *AuthService) ValidateCallbackToken(token string) (*TokenInfo, bool) {
	return s.GetToken(token)
}
