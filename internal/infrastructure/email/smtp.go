package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/google/uuid"

	"srp-biess/internal/domain"
)

// ConfigSMTP sale de SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASSWORD, SMTP_FROM, SMTP_FROM_NAME y
// SMTP_TLS (SDD 16.4.4). En desarrollo apunta a Mailpit (localhost:1025, sin TLS ni credenciales).
type ConfigSMTP struct {
	Host     string
	Puerto   int
	Usuario  string
	Clave    string
	De       string
	NombreDe string
	TLS      bool // STARTTLS obligatorio (producción)
	Espera   time.Duration
}

// SMTPSender implementa domain.EmailSender con net/smtp.
type SMTPSender struct {
	cfg ConfigSMTP
}

var _ domain.EmailSender = (*SMTPSender)(nil)

func NewSMTPSender(cfg ConfigSMTP) *SMTPSender {
	if cfg.Espera <= 0 {
		cfg.Espera = 15 * time.Second
	}
	return &SMTPSender{cfg: cfg}
}

var errDestinatario = errors.New("destinatario de correo no válido")

func (s *SMTPSender) Enviar(ctx context.Context, c domain.Correo) error {
	para, err := mail.ParseAddress(c.Para)
	if err != nil {
		return fmt.Errorf("%w: %v", errDestinatario, err)
	}
	mensaje, err := ConstruirMensaje(mail.Address{Name: s.cfg.NombreDe, Address: s.cfg.De}, *para, c, time.Now())
	if err != nil {
		return err
	}

	direccion := net.JoinHostPort(s.cfg.Host, fmt.Sprint(s.cfg.Puerto))
	conn, err := (&net.Dialer{Timeout: s.cfg.Espera}).DialContext(ctx, "tcp", direccion)
	if err != nil {
		return fmt.Errorf("no se pudo conectar al SMTP %s: %w", direccion, err)
	}
	limite := time.Now().Add(s.cfg.Espera)
	if d, ok := ctx.Deadline(); ok && d.Before(limite) {
		limite = d
	}
	_ = conn.SetDeadline(limite)
	cliente, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("saludo SMTP: %w", err)
	}
	defer cliente.Close()

	if s.cfg.TLS {
		if err := cliente.StartTLS(&tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return fmt.Errorf("STARTTLS: %w", err)
		}
	}
	if s.cfg.Usuario != "" {
		if err := cliente.Auth(smtp.PlainAuth("", s.cfg.Usuario, s.cfg.Clave, s.cfg.Host)); err != nil {
			return fmt.Errorf("autenticación SMTP: %w", err)
		}
	}
	if err := cliente.Mail(s.cfg.De); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}
	// Destinatario principal
	if err := cliente.Rcpt(para.Address); err != nil {
		return fmt.Errorf("RCPT TO (principal): %w", err)
	}
	// Destinatarios en copia (CC)
	for _, cc := range c.CC {
		addr, err := mail.ParseAddress(cc)
		if err != nil {
			return fmt.Errorf("CC inválido %s: %w", cc, err)
		}
		if err := cliente.Rcpt(addr.Address); err != nil {
			return fmt.Errorf("RCPT TO (CC %s): %w", cc, err)
		}
	}
	w, err := cliente.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}
	if _, err := w.Write(mensaje); err != nil {
		return fmt.Errorf("escribiendo el mensaje: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("cerrando el mensaje: %w", err)
	}
	return cliente.Quit()
}

// ConstruirMensaje arma un MIME multipart/mixed: texto y HTML (multipart/alternative) más los adjuntos
// en base64. El asunto y los nombres van codificados (RFC 2047); se rechazan saltos de línea en las
// cabeceras para evitar inyección.
func ConstruirMensaje(de, para mail.Address, c domain.Correo, fecha time.Time) ([]byte, error) {
	for _, v := range []string{c.Asunto, de.Name, para.Name} {
		if strings.ContainsAny(v, "\r\n") {
			return nil, errors.New("cabecera de correo con salto de línea")
		}
	}
	var buf bytes.Buffer
	mezcla := multipart.NewWriter(&buf)
	cabeceras := []string{
		"From: " + de.String(),
		"To: " + para.String(),
	}
	// Agregar cabecera CC si hay destinatarios en copia
	if len(c.CC) > 0 {
		ccAddrs := make([]string, len(c.CC))
		for i, cc := range c.CC {
			if addr, err := mail.ParseAddress(cc); err == nil {
				ccAddrs[i] = addr.String()
			} else {
				ccAddrs[i] = cc
			}
		}
		cabeceras = append(cabeceras, "Cc: "+strings.Join(ccAddrs, ", "))
	}
	cabeceras = append(cabeceras,
		"Subject: "+mime.QEncoding.Encode("UTF-8", c.Asunto),
		"Date: "+fecha.Format(time.RFC1123Z),
		"Message-ID: <"+uuid.NewString()+"@srp-biess>",
		"MIME-Version: 1.0",
		"Content-Type: multipart/mixed; boundary="+mezcla.Boundary(),
	)
	cuerpo := strings.Join(cabeceras, "\r\n") + "\r\n\r\n"

	var alt bytes.Buffer
	alternativa := multipart.NewWriter(&alt)
	for _, parte := range []struct{ tipo, contenido string }{
		{"text/plain; charset=UTF-8", c.Texto},
		{"text/html; charset=UTF-8", c.HTML},
	} {
		w, err := alternativa.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {parte.tipo},
			"Content-Transfer-Encoding": {"quoted-printable"},
		})
		if err != nil {
			return nil, err
		}
		qp := quotedprintable.NewWriter(w)
		if _, err := qp.Write([]byte(parte.contenido)); err != nil {
			return nil, err
		}
		if err := qp.Close(); err != nil {
			return nil, err
		}
	}
	if err := alternativa.Close(); err != nil {
		return nil, err
	}
	w, err := mezcla.CreatePart(textproto.MIMEHeader{
		"Content-Type": {"multipart/alternative; boundary=" + alternativa.Boundary()},
	})
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(alt.Bytes()); err != nil {
		return nil, err
	}

	for _, a := range c.Adjuntos {
		nombre := mime.QEncoding.Encode("UTF-8", strings.NewReplacer("\r", "", "\n", "", `"`, "").Replace(a.Nombre))
		w, err := mezcla.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {a.Tipo + `; name="` + nombre + `"`},
			"Content-Disposition":       {`attachment; filename="` + nombre + `"`},
			"Content-Transfer-Encoding": {"base64"},
		})
		if err != nil {
			return nil, err
		}
		codificado := base64.StdEncoding.EncodeToString(a.Contenido)
		for len(codificado) > 76 { // líneas de 76 caracteres (RFC 2045)
			if _, err := w.Write([]byte(codificado[:76] + "\r\n")); err != nil {
				return nil, err
			}
			codificado = codificado[76:]
		}
		if _, err := w.Write([]byte(codificado + "\r\n")); err != nil {
			return nil, err
		}
	}
	if err := mezcla.Close(); err != nil {
		return nil, err
	}
	return append([]byte(cuerpo), buf.Bytes()...), nil
}
