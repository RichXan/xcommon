package xutil

import (
	"crypto/tls"
	"fmt"
	"strings"

	"gopkg.in/gomail.v2"
)

const (
	HTML  = "html"
	PLAIN = "plain"
)

// EmailParams 发送邮件需要的参数
type EmailParams struct {
	From     string   // 发件邮箱
	To       []string // 收件人
	Cc       []string // 抄送
	Bcc      []string // 密件抄送
	Subject  string   // 标题
	Body     string   // 内容
	BodyType string   // 内容类型
	Attach   string   // 附件
}

type SMTPConfig struct {
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	UserName string   `yaml:"username"`
	PassWord string   `yaml:"password"`
	From     string   `yaml:"from"`
	To       []string `yaml:"to_emails"`
	Cc       []string `yaml:"cc_emails"`
	Bcc      []string `yaml:"bcc_emails"`
}

type SMTPClient struct {
	Config SMTPConfig
}

func NewSMTPClient(config SMTPConfig) *SMTPClient {
	return &SMTPClient{Config: config}
}

func (s *SMTPClient) SendEmail(config EmailParams) error {
	m := gomail.NewMessage()

	// 设置发件人
	from := strings.TrimSpace(s.Config.From)
	if config.From != "" {
		from = strings.TrimSpace(config.From)
	}
	if from == "" {
		return fmt.Errorf("sender email address is required")
	}
	m.SetHeader("From", from)

	// 处理收件人
	var to []string
	if len(config.To) > 0 {
		to = make([]string, 0, len(config.To))
		for _, addr := range config.To {
			if addr = strings.TrimSpace(addr); addr != "" {
				to = append(to, addr)
			}
		}
	} else if len(s.Config.To) > 0 {
		to = make([]string, 0, len(s.Config.To))
		for _, addr := range s.Config.To {
			if addr = strings.TrimSpace(addr); addr != "" {
				to = append(to, addr)
			}
		}
	}

	if len(to) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	m.SetHeader("To", to...)

	// 处理抄送
	if len(config.Cc) > 0 {
		cc := make([]string, 0, len(config.Cc))
		for _, addr := range config.Cc {
			if addr = strings.TrimSpace(addr); addr != "" {
				cc = append(cc, addr)
			}
		}
		if len(cc) > 0 {
			m.SetHeader("Cc", cc...)
		}
	} else if len(s.Config.Cc) > 0 {
		cc := make([]string, 0, len(s.Config.Cc))
		for _, addr := range s.Config.Cc {
			if addr = strings.TrimSpace(addr); addr != "" {
				cc = append(cc, addr)
			}
		}
		if len(cc) > 0 {
			m.SetHeader("Cc", cc...)
		}
	}

	// 处理密送
	if len(config.Bcc) > 0 {
		bcc := make([]string, 0, len(config.Bcc))
		for _, addr := range config.Bcc {
			if addr = strings.TrimSpace(addr); addr != "" {
				bcc = append(bcc, addr)
			}
		}
		if len(bcc) > 0 {
			m.SetHeader("Bcc", bcc...)
		}
	} else if len(s.Config.Bcc) > 0 {
		bcc := make([]string, 0, len(s.Config.Bcc))
		for _, addr := range s.Config.Bcc {
			if addr = strings.TrimSpace(addr); addr != "" {
				bcc = append(bcc, addr)
			}
		}
		if len(bcc) > 0 {
			m.SetHeader("Bcc", bcc...)
		}
	}

	// 设置主题
	if config.Subject == "" {
		return fmt.Errorf("email subject is required")
	}
	m.SetHeader("Subject", config.Subject)

	// 设置正文
	if config.Body == "" {
		return fmt.Errorf("email body is required")
	}
	switch config.BodyType {
	case HTML:
		m.SetBody("text/html", config.Body)
	case PLAIN:
		m.SetBody("text/plain", config.Body)
	default:
		m.SetBody("text/plain", config.Body)
	}
	if config.Attach != "" {
		m.Attach(config.Attach)
	}

	d := gomail.NewDialer(
		s.Config.Host,
		s.Config.Port,
		s.Config.UserName,
		s.Config.PassWord,
	)
	// 关闭SSL协议认证
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// 发送邮件
	return d.DialAndSend(m)
}

func (s *SMTPClient) SetTo(to []string) {
	s.Config.To = to
}

func (s *SMTPClient) SetCc(cc []string) {
	s.Config.Cc = cc
}

func (s *SMTPClient) SetBcc(bcc []string) {
	s.Config.Bcc = bcc
}

func (s *SMTPClient) AddTo(to string) {
	if to = strings.TrimSpace(to); to != "" {
		s.Config.To = append(s.Config.To, to)
	}
}

func (s *SMTPClient) AddCc(cc string) {
	if cc = strings.TrimSpace(cc); cc != "" {
		s.Config.Cc = append(s.Config.Cc, cc)
	}
}

func (s *SMTPClient) AddBcc(bcc string) {
	if bcc = strings.TrimSpace(bcc); bcc != "" {
		s.Config.Bcc = append(s.Config.Bcc, bcc)
	}
}
