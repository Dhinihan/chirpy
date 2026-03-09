# Chirpy 🐦

[English](#english) | [Português](#português)

---

## English

*Disclaimer: This documentation was written by AI, but the code was not.*

This project is a web server that implements a Twitter-like RESTful API called "Chirpy". It was developed as a guided study project from the **Boot.dev** "Build a Web Server" course.

### Features
- User authentication with JWT and Argon2id password hashing.
- RESTful endpoints to create, retrieve, and delete "chirps".
- Webhooks integration (simulating the "Polka" payment gateway to upgrade users to Chirpy Red).
- Database integration with PostgreSQL, using `sqlc` for type-safe queries and `goose` for schema migrations.
- E2E testing using `testcontainers-go`.

### Installation & Setup

1. **Prerequisites:**
   - Go 1.25+
   - PostgreSQL (running locally or via Docker)
   - Docker (required for running E2E tests via `testcontainers-go`)

2. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd chirpy
   ```

3. **Configure the environment:**
   Create a `.env` file in the root directory and add the following variables:
   ```env
   DB_URL="postgres://username:password@localhost:5432/chirpy?sslmode=disable"
   JWT_SECRET="your_super_secret_jwt_key"
   POLKA_KEY="your_polka_webhook_key"
   ```

4. **Run the server:**
   Ensure you have goose installed or use `go run` if configured, and start your app:
   ```bash
   go run main.go
   ```
   The server will start on `http://localhost:8080`.

### Testing
To run the automated tests (requires Docker to spin up the Postgres container):
```bash
go test ./...
```

---

## Português

*Aviso: Esta documentação foi escrita por IA, mas o código não.*

Este projeto é um servidor web que implementa uma API RESTful semelhante ao Twitter chamada "Chirpy". Ele foi desenvolvido como um projeto de estudo guiado através do curso "Build a Web Server" da plataforma **Boot.dev**.

### Funcionalidades
- Autenticação de usuários com JWT e hash de senhas utilizando Argon2id.
- Endpoints RESTful para criar, listar e deletar "chirps".
- Integração de webhooks (simulando um gateway de pagamento "Polka" para dar upgrade de usuários para a versão Chirpy Red).
- Integração de banco de dados com PostgreSQL, utilizando `sqlc` para queries tipadas e `goose` para migrações de esquema.
- Testes End-to-End (E2E) utilizando `testcontainers-go`.

### Instalação e Configuração

1. **Pré-requisitos:**
   - Go 1.25+
   - PostgreSQL (rodando localmente ou via Docker)
   - Docker (necessário para rodar os testes E2E via `testcontainers-go`)

2. **Clone o repositório:**
   ```bash
   git clone <url-do-repositorio>
   cd chirpy
   ```

3. **Configure o ambiente:**
   Crie um arquivo `.env` no diretório raiz e adicione as seguintes variáveis:
   ```env
   DB_URL="postgres://usuario:senha@localhost:5432/chirpy?sslmode=disable"
   JWT_SECRET="sua_chave_secreta_jwt"
   POLKA_KEY="sua_chave_do_webhook_polka"
   ```

4. **Rode o servidor:**
   ```bash
   go run main.go
   ```
   O servidor iniciará em `http://localhost:8080`.

### Testes
Para rodar os testes automatizados (necessita do Docker em execução para levantar o contêiner do Postgres):
```bash
go test ./...
```
