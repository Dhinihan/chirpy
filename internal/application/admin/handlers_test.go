package admin

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
)

func TestNewApiConfig(t *testing.T) {
	tests := []struct {
		name string
		want *apiConfig
	}{
		{"Cria uma configuração com sucesso", &apiConfig{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewApiConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewApiConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_apiConfig_MiddlewareMetricsInc(t *testing.T) {
	handler := func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Chamou")
	}
	cfg := NewApiConfig()
	newHandler := cfg.MiddlewareMetricsInc(http.HandlerFunc(handler))

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	newHandler.ServeHTTP(rr, req)
	if cfg.fileserverHits.Load() != 1 {
		t.Errorf("Middleware não adicionou ao server hits")
	}
	body, _ := io.ReadAll(rr.Body)
	if string(body) != "Chamou\n" {
		t.Errorf("%s não é o esperado", string(body))
	}
}

func Test_apiConfig_HandleMetrics(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	type fields struct {
		hits int32
	}
	type args struct {
		w   http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"começa com zero", fields{0}},
		{"mas pode ter qualquer valor", fields{9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &apiConfig{
				fileserverHits: atomic.Int32{},
			}
			cfg.fileserverHits.Store(tt.fields.hits)
			cfg.HandleMetrics(rr, req)
			expected := fmt.Sprintf("Chirpy has been visited %d times!", tt.fields.hits)
			if !strings.Contains(rr.Body.String(), expected) {
				t.Errorf("%s não contém %s\n", rr.Body.String(), expected)
			}
		})
	}
}

func Test_apiConfig_HandleReset(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	type fields struct {
		hits int32
	}
	type args struct {
		w   http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"é indepotente", fields{0}},
		{"e pode resetar qualquer valor", fields{9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &apiConfig{
				fileserverHits: atomic.Int32{},
			}
			cfg.fileserverHits.Store(tt.fields.hits)
			cfg.HandleReset(rr, req)
			expected := fmt.Sprintf("Hits: %d\n", 0)
			if !strings.Contains(rr.Body.String(), expected) {
				t.Errorf("%s não contém %s\n", rr.Body.String(), expected)
			}
			if cfg.fileserverHits.Load() != 0 {
				t.Errorf("Não resetou o número de hits\n")
			}
		})
	}
}
