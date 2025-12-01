// Package service 提供用户相关的业务逻辑服务
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/lin-snow/ech0/internal/event"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	model "github.com/lin-snow/ech0/internal/model/user"
	repository "github.com/lin-snow/ech0/internal/repository/user"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	"github.com/lin-snow/ech0/internal/transaction"
	cryptoUtil "github.com/lin-snow/ech0/internal/util/crypto"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
)

// UserService 用户服务结构体，提供用户相关的业务逻辑处理
type UserService struct {
	txManager      transaction.TransactionManager         // 事务管理器
	userRepository repository.UserRepositoryInterface     // 用户数据层接口
	settingService settingService.SettingServiceInterface // 系统设置数据层接口
	eventBus       event.IEventBus                        // 事件总线
}

// NewUserService 创建并返回新的用户服务实例
//
// 参数:
//   - userRepository: 用户数据层接口实现
//   - settingService: 系统设置数据层接口实现
//
// 返回:
//   - UserServiceInterface: 用户服务接口实现
func NewUserService(
	tm transaction.TransactionManager,
	userRepository repository.UserRepositoryInterface,
	settingService settingService.SettingServiceInterface,
	eventBusProvider func() event.IEventBus,
) UserServiceInterface {
	return &UserService{
		txManager:      tm,
		userRepository: userRepository,
		settingService: settingService,
		eventBus:       eventBusProvider(),
	}
}

// Login 用户登录验证
// 验证用户名和密码，成功后生成JWT token
//
// 参数:
//   - loginDto: 登录数据传输对象，包含用户名和密码
//
// 返回:
//   - string: 生成的JWT token
//   - error: 登录过程中的错误信息
func (userService *UserService) Login(loginDto *authModel.LoginDto) (string, error) {
	// 合法性校验
	if loginDto.Username == "" || loginDto.Password == "" {
		return "", errors.New(commonModel.USERNAME_OR_PASSWORD_NOT_BE_EMPTY)
	}

	// 将密码进行 MD5 加密
	loginDto.Password = cryptoUtil.MD5Encrypt(loginDto.Password)

	// 检查用户是否存在
	user, err := userService.userRepository.GetUserByUsername(loginDto.Username)
	if err != nil {
		return "", errors.New(commonModel.USER_NOTFOUND)
	}

	// 进行密码验证,查看外界传入的密码是否与数据库一致
	if user.Password != loginDto.Password {
		return "", errors.New(commonModel.PASSWORD_INCORRECT)
	}

	// 生成 Token
	token, err := jwtUtil.GenerateToken(jwtUtil.CreateClaims(user))
	if err != nil {
		return "", err
	}

	return token, nil
}

// Register 用户注册
// 注册新用户，包括用户数量限制检查、注册权限检查等
// 第一个注册的用户自动设置为系统管理员
//
// 参数:
//   - registerDto: 注册数据传输对象，包含用户名和密码
//
// 返回:
//   - error: 注册过程中的错误信息
func (userService *UserService) Register(registerDto *authModel.RegisterDto) error {
	// 检查用户数量是否超过限制
	users, err := userService.userRepository.GetAllUsers()
	if err != nil {
		return err
	}
	if len(users) > authModel.MAX_USER_COUNT {
		return errors.New(commonModel.USER_COUNT_EXCEED_LIMIT)
	}

	// 将密码进行 MD5 加密
	registerDto.Password = cryptoUtil.MD5Encrypt(registerDto.Password)

	newUser := model.User{
		Username: registerDto.Username,
		Password: registerDto.Password,
		IsAdmin:  false,
	}

	// 检查用户是否已经存在
	user, err := userService.userRepository.GetUserByUsername(newUser.Username)
	if err == nil && user.ID != model.USER_NOT_EXISTS_ID {
		return errors.New(commonModel.USERNAME_HAS_EXISTS)
	}

	// 检查是否该系统第一次注册用户
	if len(users) == 0 {
		// 第一个注册的用户为系统管理员
		newUser.IsAdmin = true
	}

	// 检查是否开放注册
	var setting settingModel.SystemSetting
	if err := userService.settingService.GetSetting(&setting); err != nil {
		return err
	}
	if len(users) != 0 && !setting.AllowRegister {
		return errors.New(commonModel.USER_REGISTER_NOT_ALLOW)
	}
	if err := userService.txManager.Run(func(ctx context.Context) error {
		if err := userService.userRepository.CreateUser(ctx, &newUser); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// 发布用户注册事件
	newUser.Password = "" // 不包含密码信息
	userService.eventBus.Publish(
		context.Background(),
		event.NewEvent(
			event.EventTypeUserCreated,
			event.EventPayload{
				event.EventPayloadUser: newUser,
			},
		),
	)

	return nil
}

// UpdateUser 更新用户信息
// 只有管理员可以更新用户信息，支持更新用户名、密码和头像
//
// 参数:
//   - userid: 执行更新操作的用户ID（必须为管理员）
//   - userdto: 用户信息数据传输对象，包含要更新的用户信息
//
// 返回:
//   - error: 更新过程中的错误信息
func (userService *UserService) UpdateUser(userid uint, userdto model.UserInfoDto) error {
	// 检查执行操作的用户是否为管理员
	user, err := userService.userRepository.GetUserByID(int(userid))
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 检查是否需要更新用户名
	if userdto.Username != "" && userdto.Username != user.Username {
		// 检查用户名是否已存在
		existingUser, _ := userService.userRepository.GetUserByUsername(userdto.Username)
		if existingUser.ID != model.USER_NOT_EXISTS_ID {
			return errors.New(commonModel.USERNAME_ALREADY_EXISTS)
		}
		user.Username = userdto.Username
	}

	// 检查是否需要更新密码
	if userdto.Password != "" && cryptoUtil.MD5Encrypt(userdto.Password) != user.Password {
		// 检查密码是否为空
		if userdto.Password == "" {
			return errors.New(commonModel.USERNAME_OR_PASSWORD_NOT_BE_EMPTY)
		}
		// 更新密码
		user.Password = cryptoUtil.MD5Encrypt(userdto.Password)
	}

	// 检查是否需要更新头像
	if userdto.Avatar != "" && userdto.Avatar != user.Avatar {
		// 更新头像
		user.Avatar = userdto.Avatar
	}
	if err := userService.txManager.Run(func(ctx context.Context) error {
		// 更新用户信息
		return userService.userRepository.UpdateUser(ctx, &user)
	}); err != nil {
		return err
	}

	// 发布用户更新事件
	user.Password = "" // 不包含密码信息
	userService.eventBus.Publish(
		context.Background(),
		event.NewEvent(
			event.EventTypeUserUpdated,
			event.EventPayload{
				event.EventPayloadUser: user,
			},
		),
	)

	return nil
}

// UpdateUserAdmin 更新用户的管理员权限
// 只有系统管理员、管理员可以修改其他用户的管理员权限，不能修改自己和系统管理员的权限
//
// 参数:
//   - userid: 执行操作的用户ID（必须为管理员）
//   - id: 要修改权限的用户ID
//
// 返回:
//   - error: 更新过程中的错误信息
func (userService *UserService) UpdateUserAdmin(userid uint, id uint) error {
	// 检查执行操作的用户是否为管理员
	user, err := userService.userRepository.GetUserByID(int(userid))
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 检查要修改权限的用户是否存在
	user, err = userService.userRepository.GetUserByID(int(id))
	if err != nil {
		return err
	}

	// 检查系统管理员信息
	sysadmin, err := userService.GetSysAdmin()
	if err != nil {
		return err
	}

	// 检查是否尝试修改自己或系统管理员的权限
	if userid == user.ID || id == sysadmin.ID {
		return errors.New(commonModel.INVALID_PARAMS_BODY)
	}

	user.IsAdmin = !user.IsAdmin

	if err := userService.txManager.Run(func(ctx context.Context) error {
		// 更新用户信息
		return userService.userRepository.UpdateUser(ctx, &user)
	}); err != nil {
		return err
	}

	// 发布用户更新事件
	user.Password = "" // 不包含密码信息
	userService.eventBus.Publish(
		context.Background(),
		event.NewEvent(
			event.EventTypeUserUpdated,
			event.EventPayload{
				event.EventPayloadUser: user,
			},
		),
	)

	return nil
}

// GetAllUsers 获取所有用户列表
// 返回除系统管理员外的所有用户，并移除密码信息
//
// 返回:
//   - []model.User: 用户列表（不包含密码信息）
//   - error: 获取过程中的错误信息
func (userService *UserService) GetAllUsers() ([]model.User, error) {
	allures, err := userService.userRepository.GetAllUsers()
	if err != nil {
		return nil, err
	}

	sysadmin, err := userService.GetSysAdmin()
	if err != nil {
		return nil, err
	}

	// 处理用户信息(去掉管理员用户)
	for i := range allures {
		if allures[i].ID == sysadmin.ID {
			allures = append(allures[:i], allures[i+1:]...)
			break
		}
	}

	// 处理用户信息(去掉密码)
	for i := range allures {
		allures[i].Password = ""
	}

	return allures, nil
}

// GetSysAdmin 获取系统管理员信息
//
// 返回:
//   - model.User: 系统管理员用户信息
//   - error: 获取过程中的错误信息
func (userService *UserService) GetSysAdmin() (model.User, error) {
	sysadmin, err := userService.userRepository.GetSysAdmin()
	if err != nil {
		return model.User{}, err
	}

	return sysadmin, nil
}

// DeleteUser 删除用户
// 只有管理员可以删除用户，不能删除自己和系统管理员
//
// 参数:
//   - userid: 执行删除操作的用户ID（必须为管理员）
//   - id: 要删除的用户ID
//
// 返回:
//   - error: 删除过程中的错误信息
func (userService *UserService) DeleteUser(userid, id uint) error {
	return userService.txManager.Run(func(ctx context.Context) error {
		// 检查执行操作的用户是否为管理员
		user, err := userService.userRepository.GetUserByID(int(userid))
		if err != nil {
			return err
		}
		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		// 检查要删除的用户是否存在
		user, err = userService.userRepository.GetUserByID(int(id))
		if err != nil {
			return err
		}

		sysadmin, err := userService.GetSysAdmin()
		if err != nil {
			return err
		}

		if userid == user.ID || id == sysadmin.ID {
			return errors.New(commonModel.INVALID_PARAMS_BODY)
		}

		if err := userService.userRepository.DeleteUser(ctx, id); err != nil {
			return err
		}

		return nil
	})
}

// GetUserByID 根据用户ID获取用户信息
//
// 参数:
//   - userId: 用户ID
//
// 返回:
//   - model.User: 用户信息
//   - error: 获取过程中的错误信息
func (userService *UserService) GetUserByID(userId int) (model.User, error) {
	return userService.userRepository.GetUserByID(userId)
}

// BindOAuth 绑定 OAuth2 账号
func (userService *UserService) BindOAuth(userID uint, provider string, redirectURI string) (string, error) {
	user, err := userService.userRepository.GetUserByID(int(userID))
	if err != nil {
		return "", err
	}

	if !user.IsAdmin {
		return "", bindingPermissionError(provider)
	}

	setting, err := userService.getOAuthSetting(provider)
	if err != nil {
		return "", err
	}

	state, err := jwtUtil.GenerateOAuthState(
		string(authModel.OAuth2ActionBind),
		userID,
		redirectURI,
		provider,
	)
	if err != nil {
		return "", err
	}

	authorizeURL := userService.buildOAuthAuthorizeURL(setting, provider, state)
	if authorizeURL == "" {
		return "", errors.New(commonModel.OAUTH2_NOT_CONFIGURED)
	}

	return authorizeURL, nil
}

// GetOAuthLoginURL 获取 OAuth2 登录 URL
func (userService *UserService) GetOAuthLoginURL(provider string, redirectURI string) (string, error) {
	setting, err := userService.getOAuthSetting(provider)
	if err != nil {
		return "", err
	}

	state, err := jwtUtil.GenerateOAuthState(
		string(authModel.OAuth2ActionLogin),
		authModel.NO_USER_LOGINED,
		redirectURI,
		provider,
	)
	if err != nil {
		return "", err
	}

	authorizeURL := userService.buildOAuthAuthorizeURL(setting, provider, state)
	if authorizeURL == "" {
		return "", errors.New(commonModel.OAUTH2_NOT_CONFIGURED)
	}

	return authorizeURL, nil
}

// HandleOAuthCallback 处理 OAuth2 回调
func (userService *UserService) HandleOAuthCallback(provider string, code string, state string) string {
	setting, err := userService.getOAuthSetting(provider)
	if err != nil {
		return ""
	}

	oauthState, err := jwtUtil.ParseOAuthState(state)
	if err != nil {
		return ""
	}

	if oauthState.Provider != provider {
		return ""
	}

	switch provider {
	case string(commonModel.OAuth2GITHUB):
		tokenResp, err := exchangeGithubCodeForToken(setting, code)
		if err != nil {
			fmt.Printf("Error exchanging %s code for token: %v\n", provider, err)
			return ""
		}

		githubUser, err := fetchGitHubUserInfo(setting, tokenResp.AccessToken)
		if err != nil {
			fmt.Printf("Error fetching %s user info: %v\n", provider, err)
			return ""
		}

		return userService.resolveOAuthCallback(oauthState, provider, fmt.Sprint(githubUser.ID))

	case string(commonModel.OAuth2GOOGLE):
		tokenResp, err := exchangeGoogleCodeForToken(setting, code)
		if err != nil {
			fmt.Printf("Error exchanging %s code for token: %v\n", provider, err)
			return ""
		}

		googleUser, err := fetchGoogleUserInfo(setting, tokenResp.AccessToken)
		if err != nil {
			fmt.Printf("Error fetching %s user info: %v\n", provider, err)
			return ""
		}

		return userService.resolveOAuthCallback(oauthState, provider, googleUser.Sub)

	case string(commonModel.OAuth2QQ):
		tokenResp, err := exchangeQQCodeForToken(setting, code)
		if err != nil {
			fmt.Printf("Error exchanging %s code for token: %v\n", provider, err)
			return ""
		}

		qqOpenIDResp, err := fetchQQUserInfo(tokenResp.AccessToken)
		if err != nil {
			fmt.Printf("Error fetching %s user info: %v\n", provider, err)
			return ""
		}

		return userService.resolveOAuthCallback(oauthState, provider, qqOpenIDResp.OpenID)

	case string(commonModel.OAuth2CUSTOM):
		accessToken, err := exchangeCustomCodeForToken(setting, code)
		if err != nil {
			fmt.Printf("Error exchanging %s code for token: %v\n", provider, err)
			return ""
		}

		customUserID, err := fetchCustomUserInfo(setting, accessToken)
		if err != nil {
			fmt.Printf("Error fetching %s user info: %v\n", provider, err)
			return ""
		}

		return userService.resolveOAuthCallback(oauthState, provider, customUserID)

	default:
		return ""
	}
}

func (userService *UserService) getOAuthSetting(provider string) (*settingModel.OAuth2Setting, error) {
	var setting settingModel.OAuth2Setting
	if err := userService.settingService.GetOAuth2Setting(0, &setting, true); err != nil {
		return nil, err
	}

	if setting.Provider != provider {
		return nil, errors.New(commonModel.OAUTH2_NOT_CONFIGURED)
	}

	if !setting.Enable {
		return nil, errors.New(commonModel.OAUTH2_NOT_ENABLED)
	}

	if setting.ClientID == "" || setting.RedirectURI == "" || setting.AuthURL == "" || setting.TokenURL == "" ||
		setting.UserInfoURL == "" || setting.ClientSecret == "" {
		return nil, errors.New(commonModel.OAUTH2_NOT_CONFIGURED)
	}

	return &setting, nil
}

func (userService *UserService) buildOAuthAuthorizeURL(
	setting *settingModel.OAuth2Setting,
	provider, state string,
) string {
	scope := ""
	if len(setting.Scopes) > 0 {
		scope = strings.Join(setting.Scopes, " ")
	}

	switch provider {
	case string(commonModel.OAuth2GITHUB):
		return fmt.Sprintf(
			"%s?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
			setting.AuthURL,
			url.QueryEscape(setting.ClientID),
			url.QueryEscape(setting.RedirectURI),
			url.QueryEscape(scope),
			url.QueryEscape(state),
		)
	case string(commonModel.OAuth2GOOGLE):
		params := url.Values{}
		params.Set("client_id", setting.ClientID)
		params.Set("redirect_uri", setting.RedirectURI)
		params.Set("response_type", "code")
		params.Set("state", state)
		params.Set("access_type", "offline")
		params.Set("prompt", "consent")
		if scope != "" {
			params.Set("scope", scope)
		}

		return fmt.Sprintf("%s?%s", setting.AuthURL, params.Encode())

	case string(commonModel.OAuth2QQ):
		params := url.Values{}
		params.Set("response_type", "code")
		params.Set("client_id", setting.ClientID)
		params.Set("redirect_uri", setting.RedirectURI)
		params.Set("state", state)
		params.Set("display", "pc")
		if scope != "" {
			params.Set("scope", scope)
		}
		return fmt.Sprintf("%s?%s", setting.AuthURL, params.Encode())

	case string(commonModel.OAuth2CUSTOM):
		params := url.Values{}
		params.Set("client_id", setting.ClientID)
		params.Set("redirect_uri", setting.RedirectURI)
		params.Set("response_type", "code")
		params.Set("state", state)
		if scope != "" {
			params.Set("scope", scope)
		}

		return fmt.Sprintf("%s?%s", setting.AuthURL, params.Encode())
	default:
		return ""
	}
}

func bindingPermissionError(provider string) error {
	switch provider {
	case string(commonModel.OAuth2GITHUB):
		return errors.New(commonModel.NO_PERMISSION_BINDING_GITHUB)
	case string(commonModel.OAuth2GOOGLE):
		return errors.New(commonModel.NO_PERMISSION_BINDING_GOOGLE)
	case string(commonModel.OAuth2QQ):
		return errors.New(commonModel.NO_PERMISSION_BINDING_QQ)
	case string(commonModel.OAuth2CUSTOM):
		return errors.New(commonModel.NO_PERMISSION_BINDING_CUSTOM)
	default:
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}
}

func (userService *UserService) resolveOAuthCallback(
	oauthState *authModel.OAuthState,
	provider, externalID string,
) string {
	switch oauthState.Action {
	case string(authModel.OAuth2ActionLogin):
		if oauthState.UserID != authModel.NO_USER_LOGINED {
			return ""
		}

		user, err := userService.userRepository.GetUserByOAuthID(
			context.Background(),
			provider,
			externalID,
		)
		if err != nil {
			fmt.Printf("Error fetching user by %s OAuth ID: %v\n", provider, err)
			return ""
		}

		token, err := jwtUtil.GenerateToken(jwtUtil.CreateClaims(user))
		if err != nil {
			fmt.Printf("Error generating token: %v\n", err)
			return ""
		}

		redirectURL, err := url.Parse(oauthState.Redirect)
		if err != nil {
			return ""
		}
		query := redirectURL.Query()
		query.Set("token", token)
		redirectURL.RawQuery = query.Encode()

		return redirectURL.String()

	case string(authModel.OAuth2ActionBind):
		if oauthState.UserID == authModel.NO_USER_LOGINED {
			return ""
		}

		userService.txManager.Run(func(ctx context.Context) error {
			return userService.userRepository.BindOAuth(ctx, oauthState.UserID, provider, externalID)
		})

		return oauthState.Redirect + "?bind=success"

	default:
		return ""
	}
}

// 用 code 换取 access_token
func exchangeGithubCodeForToken(
	setting *settingModel.OAuth2Setting,
	code string,
) (*authModel.GitHubTokenResponse, error) {
	data := map[string]string{
		"client_id":     setting.ClientID,
		"client_secret": setting.ClientSecret,
		"code":          code,
		"redirect_uri":  setting.RedirectURI,
	}
	jsonData, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", setting.TokenURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New("GitHub token 响应错误: " + string(body))
	}

	var tokenResp authModel.GitHubTokenResponse
	_ = json.Unmarshal(body, &tokenResp)
	return &tokenResp, nil
}

// 获取 GitHub 用户信息
func fetchGitHubUserInfo(setting *settingModel.OAuth2Setting, accessToken string) (*authModel.GitHubUser, error) {
	req, _ := http.NewRequest("GET", setting.UserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New("GitHub 用户信息请求失败: " + string(body))
	}

	var user authModel.GitHubUser
	_ = json.Unmarshal(body, &user)
	return &user, nil
}

// 用 code 换取 Google access_token
func exchangeGoogleCodeForToken(
	setting *settingModel.OAuth2Setting,
	code string,
) (*authModel.GoogleTokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", setting.ClientID)
	data.Set("client_secret", setting.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", setting.RedirectURI)
	data.Set("grant_type", "authorization_code")

	req, _ := http.NewRequest("POST", setting.TokenURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Google token 响应错误: " + string(body))
	}

	var tokenResp authModel.GoogleTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// 获取 Google 用户信息
func fetchGoogleUserInfo(setting *settingModel.OAuth2Setting, accessToken string) (*authModel.GoogleUser, error) {
	req, _ := http.NewRequest("GET", setting.UserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Google 用户信息请求失败: " + string(body))
	}

	var user authModel.GoogleUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// exchangeQQCodeForToken 用 code 换取 QQ access_token
func exchangeQQCodeForToken(setting *settingModel.OAuth2Setting, code string) (*authModel.QQTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", setting.ClientID)
	data.Set("client_secret", setting.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", setting.RedirectURI)
	data.Set("fmt", "json")
	data.Set("need_openid", "1")

	req, _ := http.NewRequest("GET", setting.TokenURL+"?"+data.Encode(), nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("QQ token 响应错误: " + string(body))
	}

	raw := strings.TrimSpace(string(body))

	// 去掉 callback(...) 包裹
	if strings.HasPrefix(raw, "callback(") && strings.HasSuffix(raw, ");") {
		raw = strings.TrimPrefix(raw, "callback(")
		raw = strings.TrimSuffix(raw, ");")
		raw = strings.TrimSpace(raw)
	}

	var tokenResp authModel.QQTokenResponse

	// 优先尝试 JSON 解析
	if err := json.Unmarshal([]byte(raw), &tokenResp); err == nil {
		if tokenResp.AccessToken != "" {
			return &tokenResp, nil
		}
	}

	// 尝试解析为 query 格式
	vals, err := url.ParseQuery(raw)
	if err == nil && vals.Get("access_token") != "" {
		tokenResp.AccessToken = vals.Get("access_token")
		tokenResp.RefreshToken = vals.Get("refresh_token")
		tokenResp.ExpiresIn, _ = strconv.ParseInt(vals.Get("expires_in"), 10, 64)
		tokenResp.OpenID = vals.Get("openid")
		return &tokenResp, nil
	}

	// 如果都失败，返回错误
	return nil, errors.New("无法解析 QQ token 响应: " + string(body))
}

// fetchQQUserInfo 获取 QQ 用户信息
func fetchQQUserInfo(accessToken string) (*authModel.QQOpenIDResponse, error) {
	// 先获取 openid
	openIDURL := "https://graph.qq.com/oauth2.0/me" + "?access_token=" + url.QueryEscape(accessToken) + "&fmt=json"
	req, _ := http.NewRequest("GET", openIDURL, nil)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("QQ openid 请求失败: " + string(body))
	}

	var openIDResp authModel.QQOpenIDResponse
	if err := json.Unmarshal(body, &openIDResp); err != nil {
		return nil, err
	}

	return &openIDResp, nil
}

// exchangeCustomCodeForToken 通用 OAuth2 令牌交换
func exchangeCustomCodeForToken(setting *settingModel.OAuth2Setting, code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", setting.ClientID)
	data.Set("client_secret", setting.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", setting.RedirectURI)
	data.Set("grant_type", "authorization_code")

	req, _ := http.NewRequest("POST", setting.TokenURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", errors.New("Custom token 响应错误: " + string(body))
	}

	var tokenResp map[string]any
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	if accessToken, ok := tokenResp["access_token"]; ok {
		if tokenStr := fmt.Sprint(accessToken); tokenStr != "" && tokenStr != "<nil>" {
			return tokenStr, nil
		}
	}

	return "", errors.New("Custom token 响应缺少 access_token")
}

// fetchCustomUserInfo 获取自定义 OAuth2 用户信息
func fetchCustomUserInfo(setting *settingModel.OAuth2Setting, accessToken string) (string, error) {
	req, _ := http.NewRequest("GET", setting.UserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("Custom 用户信息请求失败: " + string(body))
	}

	var userData map[string]any
	if err := json.Unmarshal(body, &userData); err != nil {
		return "", err
	}

	for _, key := range []string{"id", "sub", "user_id", "uid", "openid"} {
		if val, ok := userData[key]; ok {
			if id := fmt.Sprint(val); id != "" && id != "<nil>" {
				return id, nil
			}
		}
	}

	return "", errors.New("Custom 用户信息缺少唯一标识字段 (id/sub/user_id/uid)")
}

// GetOAuthInfo 获取 OAuth2 信息
func (userService *UserService) GetOAuthInfo(userId uint, provider string) (model.OAuthInfoDto, error) {
	var oauthInfo model.OAuthInfoDto

	// 检查当前用户是否存在
	user, err := userService.userRepository.GetUserByID(int(userId))
	if err != nil {
		return oauthInfo, err
	}

	// 检查用户是否为管理员
	if !user.IsAdmin {
		return oauthInfo, bindingPermissionError(provider)
	}

	oauthInfoBinding, err := userService.userRepository.GetOAuthInfo(userId, provider)
	if err != nil {
		return oauthInfo, err
	}

	oauthInfo = model.OAuthInfoDto{
		Provider: oauthInfoBinding.Provider,
		UserID:   oauthInfoBinding.UserID,
		OAuthID:  oauthInfoBinding.OAuthID,
	}

	return oauthInfo, nil
}
