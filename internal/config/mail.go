package config

type MailConfig struct {
	FromName string // 发件人名称
	HostAddr string // SMTP服务器地址
	FromAddr string // 发件人邮箱
	Password string // 邮箱密码或授权码
	UserName string // 用户名
}

// GetMailConfig 获取邮件配置
func GetMailConfig() *MailConfig {
	return &MailConfig{
		FromName: getEnv("MAIL_FROM_NAME", ""),
		HostAddr: getEnv("MAIL_HOST_ADDR", ""),
		FromAddr: getEnv("MAIL_FROM_ADDR", ""),
		Password: getEnv("MAIL_PASS_WORD", ""),
		UserName: getEnv("MAIL_USER_NAME", ""),
	}
}
