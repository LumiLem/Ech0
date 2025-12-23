package repository

import (
	"context"
	"time"

	"github.com/lin-snow/ech0/internal/cache"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	model "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type UserRepository struct {
	db    func() *gorm.DB
	cache cache.ICache[string, any]
}

func NewUserRepository(
	dbProvider func() *gorm.DB,
	cache cache.ICache[string, any],
) UserRepositoryInterface {
	return &UserRepository{
		db:    dbProvider,
		cache: cache,
	}
}

// getDB 从上下文中获取事务
func (userRepository *UserRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(transaction.TxKey).(*gorm.DB); ok {
		return tx
	}
	return userRepository.db()
}

// GetUserByUsername 根据用户名获取用户
func (userRepository *UserRepository) GetUserByUsername(username string) (model.User, error) {
	// 先查缓存
	cacheKey := GetUsernameKey(username)
	if cachedUser, err := userRepository.cache.Get(cacheKey); err == nil {
		// 缓存命中，类型断言
		if user, ok := cachedUser.(model.User); ok {
			return user, nil
		}
	}

	// 缓存未命中，查询数据库
	user := model.User{}
	err := userRepository.db().Where("username = ?", username).First(&user).Error
	if err != nil {
		return model.User{}, err
	}

	// 写缓存， cost 设为1
	userRepository.cache.Set(cacheKey, &user, 1)

	return user, nil
}

// GetAllUsers 获取所有用户
func (userRepository *UserRepository) GetAllUsers() ([]model.User, error) {
	var users []model.User
	err := userRepository.db().Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// CreateUser 创建一个新的用户
func (userRepository *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	err := userRepository.getDB(ctx).Create(user).Error
	if err != nil {
		return err
	}

	// 加入缓存
	userRepository.cache.Set(GetUserIDKey(user.ID), user, 1)
	userRepository.cache.Set(GetUsernameKey(user.Username), *user, 1)

	return nil
}

// GetUserByID 根据用户ID获取用户
func (userRepository *UserRepository) GetUserByID(id int) (model.User, error) {
	cacheKey := GetUserIDKey(uint(id))
	if cachedUser, err := userRepository.cache.Get(cacheKey); err == nil {
		// 缓存命中，类型断言
		if user, ok := cachedUser.(model.User); ok {
			return user, nil
		}
	}

	var user model.User
	if err := userRepository.db().First(&user, id).Error; err != nil {
		return user, err
	}

	userRepository.cache.Set(cacheKey, user, 1)

	return user, nil
}

// GetSysAdmin 获取系统管理员
func (userRepository *UserRepository) GetSysAdmin() (model.User, error) {
	cacheKey := GetSysAdminKey()
	if cachedUser, err := userRepository.cache.Get(cacheKey); err == nil {
		// 缓存命中，类型断言
		if user, ok := cachedUser.(model.User); ok {
			return user, nil
		}
	}

	// 获取系统管理员（首个注册的用户）
	user := model.User{}
	err := userRepository.db().Where("is_admin = ?", true).First(&user).Error
	if err != nil {
		return model.User{}, err
	}

	userRepository.cache.Set(cacheKey, user, 1)

	return user, nil
}

// UpdateUser 更新用户信息
func (userRepository *UserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	err := userRepository.getDB(ctx).Save(user).Error
	if err != nil {
		return err
	}

	userRepository.cache.Set(GetUserIDKey(user.ID), user, 1)
	userRepository.cache.Set(GetUsernameKey(user.Username), *user, 1)
	if user.IsAdmin {
		userRepository.cache.Set(GetAdminKey(user.ID), *user, 1)
	}

	return nil
}

// DeleteUser 删除用户
func (userRepository *UserRepository) DeleteUser(ctx context.Context, id uint) error {
	// 先查找待删除的用户
	userToDel, err := userRepository.GetUserByID(int(id))
	if err != nil {
		return err
	}

	err = userRepository.getDB(ctx).Delete(&model.User{}, id).Error
	if err != nil {
		return err
	}

	// 清空缓存
	userRepository.cache.Delete(GetUserIDKey(userToDel.ID))
	userRepository.cache.Delete(GetUsernameKey(userToDel.Username))
	if userToDel.IsAdmin {
		userRepository.cache.Delete(GetAdminKey(userToDel.ID))
	}

	return nil
}

// BindOAuth 绑定 OAuth 或 OIDC 账号
func (userRepository *UserRepository) BindOAuth(
	ctx context.Context,
	userID uint,
	provider, oauthID, issuer, authType string,
) error {
	// 检查是否已绑定(可能是 OAuth2 或 OIDC)
	var existing model.OAuthBinding
	if authType == string(authModel.AuthTypeOIDC) {
		// 查出 OIDC 绑定 (auth_type 为 oidc)
		err := userRepository.getDB(ctx).
			Where("user_id = ? AND provider = ? AND issuer = ? AND auth_type = ?", userID, provider, issuer, string(authModel.AuthTypeOIDC)).
			First(&existing).
			Error
		if err == nil {
			// 已绑定，更新 oauth_id
			existing.OAuthID = oauthID
			return userRepository.getDB(ctx).Save(&existing).Error
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
	} else {
		// 查出 OAuth2 绑定 (auth_type 为空或 oauth2) && issuer 为空或 issuer 为 ""
		err := userRepository.getDB(ctx).Where("user_id = ? AND provider = ? AND (issuer IS NULL OR issuer = ?) AND (auth_type = ? OR auth_type IS NULL)", userID, provider, "", string(authModel.AuthTypeOAuth2)).First(&existing).Error
		if err == nil {
			// 已绑定，更新 oauth_id
			existing.OAuthID = oauthID
			existing.AuthType = string(authModel.AuthTypeOAuth2)
			return userRepository.getDB(ctx).Save(&existing).Error
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
	}

	// 未绑定，创建新记录
	newBinding := model.OAuthBinding{
		UserID:   userID,
		Provider: provider,
		OAuthID:  oauthID,
		Issuer:   issuer,
		AuthType: authType,
	}
	if err := userRepository.getDB(ctx).Create(&newBinding).Error; err != nil {
		return err
	}

	return nil
}

// GetUserByOAuthID 根据OAuth提供商和OAuth ID获取用户
// 通过查询OAuth绑定表找到对应的用户ID，然后返回完整的用户信息
//
// 参数:
//   - ctx: 上下文对象，用于事务管理
//   - provider: OAuth提供商名称（如"github", "google", "qq", "custom"）
//   - oauthID: 第三方平台的用户唯一标识（如GitHub的user ID、QQ的OpenID）
//
// 返回:
//   - model.User: 用户信息
//   - error: 如果未找到绑定关系或用户不存在，返回错误
func (userRepository *UserRepository) GetUserByOAuthID(
	ctx context.Context,
	provider, oauthID string,
) (model.User, error) {
	var binding model.OAuthBinding
	err := userRepository.getDB(ctx).
		Where("provider = ? AND o_auth_id = ?", provider, oauthID).
		First(&binding).
		Error
	if err != nil {
		return model.User{}, err
	}

	return userRepository.GetUserByID(int(binding.UserID))
}

// GetUserByOIDC 根据 OIDC 提供商、issuer 与 sub 获取用户
func (userRepository *UserRepository) GetUserByOIDC(
	ctx context.Context,
	provider, oauthID, issuer string,
) (model.User, error) {
	var binding model.OAuthBinding
	err := userRepository.getDB(ctx).
		Where(
			"provider = ? AND o_auth_id = ? AND issuer = ? AND auth_type = ?",
			provider,
			oauthID,
			issuer,
			string(authModel.AuthTypeOIDC),
		).
		First(&binding).Error
	if err != nil {
		return model.User{}, err
	}

	return userRepository.GetUserByID(int(binding.UserID))
}

// GetOAuthInfo 获取 OAuth2 信息
func (userRepository *UserRepository) GetOAuthInfo(
	userId uint,
	provider string,
) (model.OAuthBinding, error) {
	var oauthInfo model.OAuthBinding
	err := userRepository.db().
		Where("user_id = ? AND provider = ? AND (auth_type = ? OR auth_type IS NULL)", userId, provider, string(authModel.AuthTypeOAuth2)).
		First(&oauthInfo).Error
	if err != nil {
		return model.OAuthBinding{}, err
	}

	return oauthInfo, nil
}

// GetOAuthOIDCInfo 获取 OIDC 信息
func (userRepository *UserRepository) GetOAuthOIDCInfo(
	userId uint,
	provider string,
	issuer string,
) (model.OAuthBinding, error) {
	var oauthInfo model.OAuthBinding
	err := userRepository.db().
		Where("user_id = ? AND provider = ? AND issuer = ? AND auth_type = ?", userId, provider, issuer, string(authModel.AuthTypeOIDC)).
		First(&oauthInfo).
		Error
	if err != nil {
		return model.OAuthBinding{}, err
	}
	return oauthInfo, nil
}

// -----------------------
// Passkey / WebAuthn
// -----------------------

func (userRepository *UserRepository) CreatePasskey(
	ctx context.Context,
	passkey *authModel.Passkey,
) error {
	return userRepository.getDB(ctx).Create(passkey).Error
}

func (userRepository *UserRepository) ListPasskeysByUserID(
	userID uint,
) ([]authModel.Passkey, error) {
	var passkeys []authModel.Passkey
	if err := userRepository.db().
		Where("user_id = ?", userID).
		Order("id desc").
		Find(&passkeys).Error; err != nil {
		return nil, err
	}
	return passkeys, nil
}

func (userRepository *UserRepository) GetPasskeyByCredentialID(
	credentialID string,
) (authModel.Passkey, error) {
	var passkey authModel.Passkey
	if err := userRepository.db().
		Where("credential_id = ?", credentialID).
		First(&passkey).Error; err != nil {
		return authModel.Passkey{}, err
	}
	return passkey, nil
}

func (userRepository *UserRepository) UpdatePasskeyUsage(
	ctx context.Context,
	passkeyID uint,
	signCount uint32,
	lastUsedAt time.Time,
) error {
	return userRepository.getDB(ctx).
		Model(&authModel.Passkey{}).
		Where("id = ?", passkeyID).
		Updates(map[string]any{
			"sign_count":   signCount,
			"last_used_at": lastUsedAt,
		}).Error
}

func (userRepository *UserRepository) UpdatePasskeyDeviceName(
	ctx context.Context,
	userID, passkeyID uint,
	deviceName string,
) error {
	return userRepository.getDB(ctx).
		Model(&authModel.Passkey{}).
		Where("id = ? AND user_id = ?", passkeyID, userID).
		Update("device_name", deviceName).Error
}

func (userRepository *UserRepository) DeletePasskeyByID(
	ctx context.Context,
	userID, passkeyID uint,
) error {
	return userRepository.getDB(ctx).
		Where("id = ? AND user_id = ?", passkeyID, userID).
		Delete(&authModel.Passkey{}).Error
}

func (userRepository *UserRepository) CacheSetPasskeySession(
	key string,
	val any,
	ttl time.Duration,
) {
	_ = userRepository.cache.SetWithTTL(key, val, 1, ttl)
}

func (userRepository *UserRepository) CacheGetPasskeySession(key string) (any, error) {
	return userRepository.cache.Get(key)
}

func (userRepository *UserRepository) CacheDeletePasskeySession(key string) {
	userRepository.cache.Delete(key)
}
