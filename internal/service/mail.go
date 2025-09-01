package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"llm-member/internal/config"
	"llm-member/internal/consts"
	"llm-member/internal/model"
	"log"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type MailService struct {
	cfg  *config.Config
	mail *config.MailConfig
}

func NewMailService() *MailService {
	return &MailService{
		cfg:  config.Load(),
		mail: config.GetMailConfig(),
	}
}

// SendSigninCodeEmail 发送登录验证码邮件
func (s *MailService) SendSigninCodeEmail(token *model.VerifyModel) error {
	subject := "您的登录验证码"

	// 使用模板渲染邮件内容
	templateData := map[string]any{
		"name": config.Load().AppName,
		"code": token.Token,
	}

	body, err := s.GetTemplate("mail.signin", templateData)
	if err != nil {
		return fmt.Errorf("%w: %w", consts.ErrEmailTemplateLoadFailed, err)
	}

	return s.Send([]string{token.Email}, subject, body.String())
}

// SendSignupCodeEmail 发送注册验证邮件
func (s *MailService) SendSignupCodeEmail(token *model.VerifyModel) error {
	subject := "请验证您的邮箱地址"

	// 生成验证链接
	verifyURL := fmt.Sprintf(
		"http://%s:%s/verify?code=%s",
		s.cfg.AppHost, s.cfg.AppPort, token.Token,
	)

	// 使用模板渲染邮件内容
	templateData := map[string]any{
		"name":       s.cfg.AppName,
		"code":       token.Token,
		"verify_url": verifyURL,
	}

	body, err := s.GetTemplate("mail.signup", templateData)
	if err != nil {
		return fmt.Errorf("%w: %w", consts.ErrEmailTemplateLoadFailed, err)
	}

	return s.Send([]string{token.Email}, subject, body.String())
}

// SendResetCodeEmail 发送密码重置邮件
func (s *MailService) SendResetCodeEmail(token *model.VerifyModel) error {
	link := fmt.Sprintf("http://localhost:8080/reset?code=%s", token.Token)
	subject := "重置您的密码"

	// 使用模板渲染邮件内容
	templateData := map[string]any{
		"name": config.Load().AppName,
		"link": link,
	}

	body, err := s.GetTemplate("mail.reset", templateData)
	if err != nil {
		return fmt.Errorf("%w: %w", consts.ErrEmailTemplateLoadFailed, err)
	}

	return s.Send([]string{token.Email}, subject, body.String())
}

// 参考net/smtp的func SendMail()
// 使用net.Dial连接tls（SSL）端口时，smtp.NewClient()会卡住且不提示err
// len(to)>1时，to[1]开始提示是密送
func (self *MailService) Send(to []string, title string, body string) (err error) {
	client, err := self.Dial(self.mail.HostAddr)
	if err != nil {
		log.Println("Create smpt client error:", err)
		return err
	}
	defer client.Close()

	host := strings.Split(self.mail.HostAddr, ":")[0]
	auth := smtp.PlainAuth("", self.mail.UserName, self.mail.Password, host)
	if auth == nil {
		log.Println("smpt auth error:", err)
		return fmt.Errorf("%w", consts.ErrSMTPAuthError)
	}
	if ok, _ := client.Extension("AUTH"); ok {
		if err = client.Auth(auth); err != nil {
			log.Println("Error during AUTH", err)
			return err
		}
	}

	if err = client.Mail(self.mail.FromAddr); err != nil {
		return err
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	var writer io.WriteCloser
	if writer, err = client.Data(); err != nil {
		return err
	}

	// send
	msg := self.Text(to, title, body)
	if _, err = writer.Write(msg); err != nil {
		return err
	}

	// close
	if err = writer.Close(); err != nil {
		return err
	}

	return client.Quit()
}

// return a smtp client
func (self *MailService) Dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		return nil, err
	}
	//分解主机端口字符串
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

func (self *MailService) Text(receiver []string, title string, body string) []byte {
	header := []string{
		fmt.Sprintf("To: %s\r\n", strings.Join(receiver, ",")),
		fmt.Sprintf("Subject: %s\r\n", title),
		fmt.Sprintf("From: %s<%s>\r\n", self.mail.FromName, self.mail.FromAddr),
		fmt.Sprintf("Content-Type: %s\r\n", "text/html; charset=UTF-8"),
	}

	message := strings.Join(header, "") + "\r\n" + body
	return []byte(message)
}

func (self *MailService) GetTemplate(path string, data map[string]any) (*bytes.Buffer, error) {
	body := new(bytes.Buffer)

	// 处理相对路径，如果以../开头，则从storage目录读取
	var filePath string
	if strings.HasPrefix(path, "../") {
		filePath = filepath.Join("storage", strings.TrimPrefix(path, "../")+".html")
	} else {
		filePath = filepath.Join("template", path+".html")
	}

	// 检查模板文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return body, fmt.Errorf("%w: %s", consts.ErrTemplateFileNotFound, filePath)
	}

	// 解析模板文件
	if tmpl, err := template.ParseFiles(filePath); err != nil {
		return body, err
	} else if err = tmpl.Execute(body, data); err != nil {
		return body, err
	}
	return body, nil
}
