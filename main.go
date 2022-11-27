package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/send", send)

	err := http.ListenAndServe(":8081", mux)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server closed")
	} else if err != nil {
		fmt.Println("error starting server: ", err)
		os.Exit(1)
	}
}

func send(w http.ResponseWriter, r *http.Request) {

	var emailBody EmailBody

	jsonBody, errBody := io.ReadAll(r.Body)
	if errBody != nil {
		fmt.Println("http post body err: ", errBody)
		return
	}
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
	toEmail := getEnvStr("TO_EMAIL", emailBody.ToEmail)

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

	auth := smtp.PlainAuth("", fromEmail, password, host)
	err := SendMailWithTLS(fmt.Sprintf("%s:%d", host, port), auth, fromEmail, []string{toEmail}, []byte(message))
	if err != nil {
		fmt.Println("send email err: ", err)
	} else {
		fmt.Println("send mail success", toEmail)
	}
}

func Dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		log.Println("tls.Dial Error:", err)
		return nil, err
	}

	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

func SendMailWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) (err error) {
	c, err := Dial(addr)
	if err != nil {
		log.Println("Create smtp client error:", err)
		return err
	}
	defer c.Close()
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				log.Println("Error during AUTH", err)
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
