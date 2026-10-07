package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"testing"

	"github.com/stretchr/testify/require"
)

// Client es un cliente HTTP para los tests que guarda el access token y la
// cookie de refresh, como haria el frontend.
type Client struct {
	t           *testing.T
	base        string
	http        *http.Client
	AccessToken string
}

func NewClient(t *testing.T, app *TestApp) *Client {
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	return &Client{
		t:    t,
		base: app.Address + "/api/v1",
		http: &http.Client{Jar: jar},
	}
}

type Response struct {
	Status int
	Body   []byte
	Header http.Header
}

func (r Response) JSON(t *testing.T, v any) {
	require.NoError(t, json.Unmarshal(r.Body, v), string(r.Body))
}

func (c *Client) Do(method, path string, body io.Reader, contentType string) Response {
	req, err := http.NewRequest(method, c.base+path, body)
	require.NoError(c.t, err)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}

	resp, err := c.http.Do(req)
	require.NoError(c.t, err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	require.NoError(c.t, err)

	return Response{Status: resp.StatusCode, Body: data, Header: resp.Header}
}

func (c *Client) JSON(method, path string, payload any) Response {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		require.NoError(c.t, err)
		body = bytes.NewReader(data)
	}
	return c.Do(method, path, body, "application/json")
}

type Archivo struct {
	Nombre    string
	Contenido []byte
}

func (c *Client) Multipart(method, path string, campos map[string]string, archivo *Archivo) Response {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	for k, v := range campos {
		require.NoError(c.t, w.WriteField(k, v))
	}

	if archivo != nil {
		part, err := w.CreateFormFile("archivo", archivo.Nombre)
		require.NoError(c.t, err)
		_, err = part.Write(archivo.Contenido)
		require.NoError(c.t, err)
	}

	require.NoError(c.t, w.Close())
	return c.Do(method, path, &buf, w.FormDataContentType())
}

// Registrar crea una cuenta e inicia sesion. Devuelve el id del alumno.
func (c *Client) Registrar(nombre, email string) int64 {
	resp := c.JSON("POST", "/signup", map[string]any{
		"nombre":     nombre,
		"email":      email,
		"carrera_id": 1,
		"password":   "Chedul-123",
	})
	require.Equal(c.t, http.StatusCreated, resp.Status, string(resp.Body))

	login := c.JSON("POST", "/login", map[string]any{
		"email":    email,
		"password": "Chedul-123",
	})
	require.Equal(c.t, http.StatusOK, login.Status, string(login.Body))

	var data struct {
		AccessToken string `json:"accessToken"`
		User        struct {
			ID int64 `json:"id"`
		} `json:"user"`
	}
	login.JSON(c.t, &data)
	require.NotEmpty(c.t, data.AccessToken)
	c.AccessToken = data.AccessToken

	return data.User.ID
}

func (c *Client) MateriaID(codigo string) int64 {
	resp := c.JSON("GET", "/materias", nil)
	require.Equal(c.t, http.StatusOK, resp.Status)

	var materias []struct {
		ID     int64  `json:"id"`
		Codigo string `json:"codigo"`
	}
	resp.JSON(c.t, &materias)

	for _, m := range materias {
		if m.Codigo == codigo {
			return m.ID
		}
	}

	c.t.Fatalf("no existe la materia %s", codigo)
	return 0
}

func (c *Client) CondicionID(nombre string) int64 {
	resp := c.JSON("GET", "/condicion", nil)
	require.Equal(c.t, http.StatusOK, resp.Status)

	var condiciones []struct {
		ID        int64  `json:"id"`
		Condicion string `json:"condicion"`
	}
	resp.JSON(c.t, &condiciones)

	for _, cond := range condiciones {
		if cond.Condicion == nombre {
			return cond.ID
		}
	}

	c.t.Fatalf("no existe la condicion %s", nombre)
	return 0
}

func ruta(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
