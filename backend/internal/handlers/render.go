package handlers

import (
	"bytes"
	"html/template"
	"os"

	"github.com/labstack/echo/v4"
)

var verifyTpl = template.Must(template.New("verifyPage").Parse(`
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>{{.Title}}</title>
  <style>
    body {
      font-family: system-ui, Arial, sans-serif;
      background: linear-gradient(135deg, #0f1221, #1b2b4b);
      color: #f0f0f0;
      text-align: center;
      padding-top: 10%;
    }
    .card {
      display: inline-block;
      padding: 32px 48px;
      border-radius: 16px;
      background: rgba(255,255,255,0.05);
      border: 1px solid rgba(255,255,255,0.15);
      box-shadow: 0 12px 24px rgba(0,0,0,0.3);
    }
    h2 { margin-bottom: 12px; color: {{.Color}}; }
    p { margin: 0 0 8px; }
    a {
      display:inline-block;
      margin-top: 16px;
      padding: 10px 20px;
      border-radius: 8px;
      background: linear-gradient(135deg,#6c8bff,#55e7ff);
      color: #0b1120;
      text-decoration: none;
      font-weight: 600;
    }
    a:hover { box-shadow: 0 4px 12px rgba(110,160,255,0.5); }
  </style>
</head>
<body>
  <div class="card">
    <h2>{{.Title}}</h2>
    <p>{{.Message}}</p>
    <a href="{{.LoginURL}}">Go to Sign In</a>
  </div>
</body>
</html>
`))

type verifyPageData struct {
	Title    string
	Message  string
	Color    string
	LoginURL string
}

func RenderVerificationPage(c echo.Context, code int, ok bool, title, message string) error {
	color := "#3ce37a"
	if !ok {
		color = "#ff6b6b"
	}
	login := os.Getenv("FRONTEND_LOGIN_URL")
	if login == "" {
		login = "http://localhost:3000/login"
	}
	var buf bytes.Buffer
	_ = verifyTpl.Execute(&buf, verifyPageData{
		Title:    title,
		Message:  message,
		Color:    color,
		LoginURL: login,
	})
	return c.HTML(code, buf.String())
}
