package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
)

var contextChan chan *Context

func main() {

	numWorkers := 32
	contextChan = make(chan *Context, 10)
	CreateWorkers(numWorkers, contextChan)

	http.HandleFunc("/send", HandleAllRequest)
	http.ListenAndServe(":80", nil)
}

func CreateWorkers(numWorkers int, contextChan chan *Context) {
	for i := 0; i < numWorkers; i++ {
		go func(contextChan chan *Context) {
			for context := range contextChan {
				sendHandler(context.RequestBody)
			}
		}(contextChan)
	}
}

func HandleAllRequest(w http.ResponseWriter, r *http.Request) {
	jsonBody, _ := io.ReadAll(r.Body)
	contextChan <- &Context{Response: w, RequestBody: jsonBody}
}

func sendHandler(jsonBody []byte) {

	var emailBody EmailBody

	errJson := json.Unmarshal(jsonBody, &emailBody)
	if errJson != nil {
		fmt.Println("body to json err: ", errJson, jsonBody)
		return
	}

	host := getEnvStr("HOST", emailBody.Host)
	port := getEnvInt("PORT", emailBody.Port)
	password := getEnvStr("PASSWORD", emailBody.Password)
	fromEmail := getEnvStr("FROM_EMAIL", emailBody.FromEmail)
	fromName := getEnvStr("FROM_NAME", emailBody.FromName)
	toEmail := emailBody.ToEmail

	headers := make(map[string]string)
	headers["From"] = fromName + " <" + fromEmail + ">"
	headers["To"] = toEmail
	headers["Subject"] = getEnvStr("SUBJECT", emailBody.Subject)
	headers["Content-Type"] = getEnvStr("CONTENT_TYPE", emailBody.ContentType)

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + emailBody.Body

	errSend := SendMailWithTLS(
		fmt.Sprintf("%s:%d", host, port),
		smtp.PlainAuth("", fromEmail, password, host),
		fromEmail, []string{toEmail}, []byte(message),
	)
	if errSend != nil {
		fmt.Println("send email err: ", errSend)
	} else {
		fmt.Println("send mail success", toEmail)
	}
}

func SendMailWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) (err error) {
	c, err := Dial(addr)
	if err != nil {
		fmt.Println("Create smtp client error:", err)
		return err
	}
	defer c.Close()
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				fmt.Println("Error during AUTH", err)
				return err
			}
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

func Dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		fmt.Println("tls.Dial Error:", err)
		return nil, err
	}

	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

func getEnvStr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		ret, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		} else {
			return ret
		}
	}
	return fallback
}

type Context struct {
	Response    http.ResponseWriter
	RequestBody []byte
}

type EmailBody struct {
	ContentType string
	FromEmail   string
	FromName    string
	Password    string
	ToEmail     string
	Subject     string
	Body        string
	Host        string
	Port        int
}
