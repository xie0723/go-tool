package email

import (
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"strings"

	"github.com/astaxie/beego"
	"github.com/go-gomail/gomail"
)

type email struct {
	ToEmails   string // 需要发送的email地址，分号分隔
	CcEmails   string // 需要抄送的email地址，分号分隔
	Sender     string // 邮件发送者地址
	SenderPwd  string // 邮件发送者密码
	SenderName string // 邮件发送这姓名
	Subject    string // 邮件标题
	Body       string // 邮件内容(html/text)
	SmtpServer string // 邮件服务器
	SmtpPort   int    // 邮件服务器端口
	AttachPath string // 附件的绝对路径
	AttachName string // 附件显示的名称
}

//var once sync.Once
//var ins *email

// 初始化实例
func New(senderName, subject, bodystring, toEmails string) *email {
	//once.Do(func() {
	ins := new(email)
	//})
	ins.SenderName = senderName
	ins.Subject = subject
	ins.Body = bodystring
	ins.ToEmails = toEmails
	return ins
}

// 设置邮件抄送地址
func (e *email) SetCcEmails(ccEmails string) *email {
	e.CcEmails = ccEmails
	return e
}

// 设置发送者
func (e *email) SetSender(sender string, senderpwd string) *email {
	e.Sender = sender
	e.SenderPwd = senderpwd
	return e
}

// 设置SMTP
func (e *email) SetSMTP(smtpServer string, smtpPort int) *email {
	e.SmtpServer = smtpServer
	e.SmtpPort = smtpPort
	return e
}

// 从配置文件中读取参数
func (e *email) LoadBeegoConf() *email {
	e.Sender = beego.AppConfig.String("Email::Addr")
	e.SmtpServer = beego.AppConfig.String("Email::SMTPServer")
	e.SmtpPort, _ = beego.AppConfig.Int("Email::SMTPPort")
	e.SenderPwd = beego.AppConfig.String("Email::Pwd")
	return e
}

// 设置附件信息
func (e *email) SetAttach(name, path string) *email {
	e.AttachName = name
	e.AttachPath = path
	return e
}

// 发送邮件
func (e *email) Send() (err error) {
	if strings.TrimSpace(e.ToEmails) == "" {
		return errors.New("to emails is nil")
	}
	m := gomail.NewMessage()
	fromemail := e.Sender
	toemail := strings.Split(e.ToEmails, ";")
	m.SetHeaders(map[string][]string{
		"From":    {m.FormatAddress(fromemail, e.SenderName)},
		"To":      toemail,
		"Subject": {e.Subject},
	})
	if strings.TrimSpace(e.CcEmails) != "" {
		ccemail := strings.Split(e.CcEmails, ";")
		m.SetHeader("Cc", ccemail...)
	}
	m.SetBody("text/html", e.Body)
	if len(strings.TrimSpace(e.AttachPath)) > 0 {
		m.Attach(
			e.AttachPath, gomail.SetHeader(map[string][]string{
				"Content-Disposition": {
					fmt.Sprintf(`attachment; filename="%s"`, mime.QEncoding.Encode("UTF-8", e.AttachName)),
				},
			}))
	}
	// qq企业邮箱登录需要后缀
	d := gomail.NewDialer(e.SmtpServer, e.SmtpPort, e.Sender, e.SenderPwd)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return d.DialAndSend(m)
}
