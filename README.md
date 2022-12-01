# SMTP-HTTP

This is a docker container for expose http port for use self smtp service.

## Deploy

### configure

Support two ways to transmit configuration parameters, `docker env` or `http request body`.

#### docker env

- `HOST` the mail server smtp service host, eg: smtp.gmail.com
- `PORT` the mail server smtp service port, eg: 465
- `FROM_EMAIL` the sender email address, eg: no-reply@example.com
- `FROM_NAME` the sender name, eg: Order-Notify
- `PASSWORD` the sender password
- `CONTENT_TYPE` the email content type. eg: text/plain; charset=UTF-8
- `SUBJECT` the email subject. eg: Order Notify

#### http request body

- `Host` the mail server smtp service host, eg: smtp.gmail.com
- `Port` the mail server smtp service port, eg: 465
- `FromEmail` the sender email address, eg: no-reply@example.com
- `FromName` the sender name, eg: Order-Notify
- `Password` the sender password
- `ContentType` the email content type. eg: text/plain; charset=UTF-8
- `Subject` the email subject. eg: Order Notify
- `ToEmail` the receiver email address, eg: receiver@gmail.com

### use docker

``` shell
docker run -d --restart=unless-stopped --name=smtp-http -p 18081:80 -e HOST=mail.example.com -e PORT=465 xu42/smtp-http
```

### use docker-compose

``` shell
# edit the .env file
docker-compose up -d
```

## Usage

```shell
curl -H 'content-type: application/json' -X POST "http://127.0.0.1:18081/send" -d '{"fromName":"Order Notify","toEmail":"receiver@gmail.com","subject":"Order Notify: New Order","body":"this is a test email"}'
```

## Reference
- https://www.cnblogs.com/aaronhoo/p/16364492.html
- https://cloud.tencent.com/document/product/1288/65752

## License

The MIT License (MIT). Please see [License File](LICENSE) for more information.
