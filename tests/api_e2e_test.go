package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Dhinihan/chirpy/internal/application/admin"
	"github.com/Dhinihan/chirpy/internal/application/api"
	"github.com/Dhinihan/chirpy/internal/auth"
	"github.com/Dhinihan/chirpy/internal/database"
	"github.com/Dhinihan/chirpy/internal/model/chirp"
	"github.com/Dhinihan/chirpy/internal/model/user"
	"github.com/google/uuid"
	_ "github.com/lib/pq" // Driver necessário
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// APITestSuite agrupa tudo que precisamos para o teste E2E
type APITestSuite struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	db          *sql.DB
	ctx         context.Context
	queries     *database.Queries
	mux         *http.ServeMux
}

// Esta função é o ponto de entrada que o 'go test' reconhece
func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}

// SetupSuite roda uma única vez antes de todos os testes da suite
func (s *APITestSuite) SetupSuite() {
	s.ctx = context.Background()

	container, err := postgres.Run(s.ctx,
		"postgres:15-alpine",                 // Imagem desejada
		postgres.WithDatabase("chirpy_test"), // Nome do banco
		postgres.WithUsername("user"),        // Usuário
		postgres.WithPassword("pass"),        // Senha
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		s.T().Fatal("Falha ao iniciar o container:", err)
	}
	s.pgContainer = container

	// 2. Pega a URL de conexão (o Testcontainers escolhe uma porta livre no seu PC)
	connStr, err := container.ConnectionString(s.ctx, "sslmode=disable")
	if err != nil {
		s.FailNow("Falha ao obter string de conexão", err)
	}

	// 3. Abre a conexão real com o banco
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		s.FailNow("Falha ao conectar no banco", err)
	}
	s.db = dbConn

	// 4. Inicializa o sqlc com essa conexão
	s.queries = database.New(s.db)

	// 5. Rodar migrações do Goose
	err = goose.Up(s.db, "../sql/schema")
	if err != nil {
		s.T().Fatal("Falha ao rodar migrações do Goose:", err)
	}

	s.mux = http.NewServeMux()
	cfg := admin.NewApiConfig(s.queries, "segredo")
	api.RegisterHandlers(cfg, s.mux)

}

// TearDownSuite roda uma única vez após todos os testes terminarem
func (s *APITestSuite) TearDownSuite() {
	if s.pgContainer != nil {
		err := s.pgContainer.Terminate(s.ctx)
		if err != nil {
			s.T().Log("Erro ao finalizar o container:", err)
		}
	}
}

func (s *APITestSuite) SetupTest() {
	err := s.queries.ResetUsers(s.ctx)
	if err != nil {
		s.T().Fatalf("Falha ao limpar o banco de dados: %v", err)
	}
}

// ------------ HELPERS ------------------

// executeRequest é um helper para não repetir código de boilerpate
func (s *APITestSuite) executeRequest(method, url, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	res := httptest.NewRecorder()
	s.mux.ServeHTTP(res, req)
	return res
}
func (s *APITestSuite) executeAuthRequest(
	method, url, body string,
	uid uuid.UUID,
) *httptest.ResponseRecorder {
	token, _ := auth.MakeJWT(uid, "segredo", time.Minute)
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Authorization", "bearer "+token)
	res := httptest.NewRecorder()
	s.mux.ServeHTTP(res, req)
	return res
}

func (s *APITestSuite) createUser(email string) uuid.UUID {
	body := fmt.Sprintf(`{"email": "%s", "password": "senha" }`, email)
	res := s.executeRequest("POST", "/api/users", body)
	var user user.User
	json.Unmarshal(res.Body.Bytes(), &user)
	return user.ID
}

func (s *APITestSuite) createChirp(body string, userId uuid.UUID) uuid.UUID {
	res := s.executeAuthRequest(
		"POST",
		"/api/chirps",
		fmt.Sprintf(
			`{"body": "%s"}`,
			body,
		),
		userId,
	)
	var chirp chirp.Chirp
	json.Unmarshal(res.Body.Bytes(), &chirp)
	return chirp.ID
}

func (s *APITestSuite) generateAuthUser(email string) user.User {
	body := fmt.Sprintf(`{"email": "%s", "password": "senha" }`, email)
	res := s.executeRequest("POST", "/api/users", body)
	var user user.User
	json.Unmarshal(res.Body.Bytes(), &user)
	res2 := s.executeRequest("POST", "/api/login",
		fmt.Sprintf(
			`{"email": "%s", "password": "%s"}`,
			email,
			"senha",
		))
	json.Unmarshal(res2.Body.Bytes(), &user)
	return user
}

// --------------- TESTES ---------------

func (s *APITestSuite) TestPostUsers() {
	email := "test@example.com"
	password := "senha segura"
	body := fmt.Sprintf(
		`{"email": "%s", "password": "%s"}`,
		email,
		password,
	)
	res := s.executeRequest("POST", "/api/users", body)

	s.Equal(201, res.Code)
	s.Contains(res.Body.String(), email)
	s.Contains(res.Body.String(), `"id"`)
	s.Contains(res.Body.String(), `"created_at"`)
	s.Contains(res.Body.String(), `"updated_at"`)
	s.NotContains(res.Body.String(), "Password")
	s.NotContains(res.Body.String(), "senha")
	s.NotContains(res.Body.String(), "segura")
}

func (s *APITestSuite) TestUpdateUser() {
	email := "novo@email.com"
	password := "senha"
	uid := s.createUser("velho@email")
	credsTemplate := `{"email": "%s", "password": "%s"}`
	creds := fmt.Sprintf(credsTemplate, email, password)
	res := s.executeAuthRequest("PUT", "/api/users", creds, uid)
	s.Equal(200, res.Code)
	s.Contains(res.Body.String(), email)
	s.Contains(res.Body.String(), `"id"`)
	s.Contains(res.Body.String(), `"created_at"`)
	s.Contains(res.Body.String(), `"updated_at"`)
	s.NotContains(res.Body.String(), "Password")
	s.NotContains(res.Body.String(), "senha")
	s.NotContains(res.Body.String(), "segura")
}

func (s *APITestSuite) TestPostDuplicatedUser() {
	email := "test@example.com"
	body := fmt.Sprintf(`{"email": "%s"}`, email)
	s.executeRequest("POST", "/api/users", body)
	res := s.executeRequest("POST", "/api/users", body)

	s.Equal(500, res.Code)
	s.Contains(res.Body.String(), `"Não foi possível criar o usuário"`)
}

func (s *APITestSuite) TestPostChirp() {
	userId := s.createUser("test@example.com")
	msg := "Valid message"
	body, _ := json.Marshal(struct {
		Body string `json:"body"`
	}{msg})
	res := s.executeAuthRequest(
		"POST",
		"/api/chirps",
		string(body),
		userId,
	)
	s.Equal(201, res.Code)

	var resp chirp.Chirp
	err := json.Unmarshal(res.Body.Bytes(), &resp)
	s.Require().NoError(err, "Erro ao ler resposta")
	s.Equal(resp.Body, msg)
	s.Equal(resp.UserID.String(), userId.String())
}

func (s *APITestSuite) TestPostInvalidChirp() {
	userId := s.createUser("test@example.com")
	msg := ""
	body, _ := json.Marshal(struct {
		Body string `json:"body"`
	}{msg})
	res := s.executeAuthRequest(
		"POST",
		"/api/chirps",
		string(body),
		userId,
	)
	s.Equal(400, res.Code)
	s.Contains(res.Body.String(), "Chirp not informed")
}

func (s *APITestSuite) TestPostDirtyChirp() {
	userId := s.createUser("test@example.com")
	msg := "It is a kerfuffle"
	body, _ := json.Marshal(struct {
		Body string `json:"body"`
	}{msg})
	res := s.executeAuthRequest(
		"POST",
		"/api/chirps",
		string(body),
		userId,
	)
	s.Equal(201, res.Code)
	s.Contains(res.Body.String(), "It is a ****")
}

func (s *APITestSuite) TestGetAllChirps() {
	userA := s.createUser("testA@email.com")
	userB := s.createUser("testB@email.com")
	bodyA := "Body A"
	bodyB := "Body B"
	bodyC := "Body C"
	s.createChirp(bodyB, userA)
	s.createChirp(bodyC, userA)
	s.createChirp(bodyA, userB)

	res := s.executeRequest("GET", "/api/chirps", "")
	s.Equal(200, res.Code)
	s.Regexp(
		fmt.Sprintf("%s.+%s.+%s", bodyB, bodyC, bodyA),
		res.Body.String(),
	)
}

func (s *APITestSuite) TestGetChirp() {
	userId := s.createUser("test@email.com")
	body := "Corpo esperado"
	chirpId := s.createChirp(body, userId)

	res := s.executeRequest("GET", "/api/chirps/"+chirpId.String(), "")
	s.Equal(200, res.Code)
	s.Contains(res.Body.String(), body)
	s.Contains(res.Body.String(), userId.String())
}

func (s *APITestSuite) TestValidLogin() {
	email := "email@valido.com"
	uid := s.createUser(email)
	req := fmt.Sprintf(
		`{"email": "%s", "password": "%s"}`,
		email,
		"senha",
	)
	res := s.executeRequest("POST", "/api/login", req)

	s.Equal(res.Code, 200)
	s.Contains(res.Body.String(), uid.String())
	s.Contains(res.Body.String(), `"token":`)
	s.Contains(res.Body.String(), `"refresh_token":`)
}

func (s *APITestSuite) TestInvalidLoginWrongPassword() {
	email := "email@valido.com"
	s.createUser(email)
	req := fmt.Sprintf(
		`{"email": "%s", "password": "%s"}`,
		email,
		"senha errada",
	)
	res := s.executeRequest("POST", "/api/login", req)

	s.Equal(res.Code, 401)
}

func (s *APITestSuite) TestInvalidLoginWrongEmail() {
	email := "email@valido.com"
	s.createUser(email)
	req := fmt.Sprintf(
		`{"email": "email@errado.com", "password": "%s"}`,
		"senha",
	)
	res := s.executeRequest("POST", "/api/login", req)

	s.Equal(res.Code, 401)
}

func (s *APITestSuite) TestRefreshToken() {
	user := s.generateAuthUser("logado@email.com")
	req := httptest.NewRequest(
		"POST",
		"/api/refresh",
		strings.NewReader(""),
	)
	req.Header.Set("Authorization", "bearer "+user.RefreshToken)
	res := httptest.NewRecorder()
	s.mux.ServeHTTP(res, req)
	s.Equal(200, res.Code)
	s.Contains(res.Body.String(), `"token":`)
}

func (s *APITestSuite) TestRevokeRefreshToken() {
	user := s.generateAuthUser("logado@email.com")
	req1 := httptest.NewRequest(
		"POST",
		"/api/revoke",
		strings.NewReader(""),
	)
	req1.Header.Set("Authorization", "bearer "+user.RefreshToken)
	res1 := httptest.NewRecorder()
	s.mux.ServeHTTP(res1, req1)
	s.Equal(204, res1.Code)
	req2 := httptest.NewRequest(
		"POST",
		"/api/refresh",
		strings.NewReader(""),
	)
	req2.Header.Set("Authorization", "bearer "+user.RefreshToken)
	res2 := httptest.NewRecorder()
	s.mux.ServeHTTP(res2, req2)
	s.Equal(401, res2.Code)
	s.NotContains(res2.Body.String(), `"token":`)
}
