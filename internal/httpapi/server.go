package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/example/genome-variant-annotation/internal/job"
	"github.com/example/genome-variant-annotation/internal/normalization"
	"github.com/example/genome-variant-annotation/internal/platform"
	"github.com/example/genome-variant-annotation/internal/reference"
	"github.com/example/genome-variant-annotation/internal/variant"
)

type Server struct {
	references *reference.Service
	jobs       *job.Service
	normalizer *normalization.Service
	token      string
	started    time.Time
	requests   atomic.Uint64
}

func NewServer(references *reference.Service, jobs *job.Service, normalizer *normalization.Service, token string) *Server {
	return &Server{references: references, jobs: jobs, normalizer: normalizer, token: token, started: time.Now()}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/readyz", s.ready)
	mux.HandleFunc("/metrics", s.metrics)
	mux.HandleFunc("/v1/reference-datasets", s.datasets)
	mux.HandleFunc("/v1/reference-datasets/", s.datasetAction)
	mux.HandleFunc("/v1/annotation-jobs", s.annotationJobs)
	mux.HandleFunc("/v1/annotation-jobs/", s.annotationJob)
	mux.HandleFunc("/v1/normalization/preview", s.normalizationPreview)
	mux.HandleFunc("/v1/variants/", s.variantQuery)
	mux.HandleFunc("/v1/regions/", s.regionQuery)
	return s.middleware(mux)
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.requests.Add(1)
		w.Header().Set("X-Request-ID", platform.NewID("req"))
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/healthz" && r.URL.Path != "/readyz" && r.URL.Path != "/metrics" {
			got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if subtle.ConstantTimeCompare([]byte(got), []byte(s.token)) != 1 {
				writeError(w, platform.ErrInvalid, http.StatusUnauthorized)
				return
			}
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				writeError(w, fmt.Errorf("internal failure"), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func decode(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, err error, status int) {
	writeJSON(w, status, map[string]any{"error": map[string]any{"message": err.Error(), "status": status}})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "uptime": time.Since(s.started).String()})
}
func (s *Server) ready(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
func (s *Server) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "annotation_http_requests_total %d\n", s.requests.Load())
}

type createDatasetRequest struct {
	Name     string              `json:"name"`
	Build    string              `json:"build"`
	Features []reference.Feature `json:"features"`
}

func (s *Server) datasets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, platform.ErrInvalid, http.StatusMethodNotAllowed)
		return
	}
	var request createDatasetRequest
	if err := decode(w, r, &request); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	dataset, err := s.references.Create(r.Context(), request.Name, request.Build)
	if err == nil && len(request.Features) > 0 {
		err = s.references.AddFeatures(r.Context(), dataset.ID, request.Features)
		if err == nil {
			dataset, _ = s.references.Get(r.Context(), dataset.ID)
		}
	}
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, dataset)
}

func (s *Server) datasetAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		writeError(w, platform.ErrInvalid, http.StatusBadRequest)
		return
	}
	id := parts[2]
	if len(parts) == 4 && parts[3] == "publish" && r.Method == http.MethodPost {
		dataset, err := s.references.Publish(r.Context(), id)
		if err != nil {
			writeError(w, err, statusFor(err))
			return
		}
		writeJSON(w, http.StatusOK, dataset)
		return
	}
	if r.Method == http.MethodGet {
		dataset, err := s.references.Get(r.Context(), id)
		if err != nil {
			writeError(w, err, statusFor(err))
			return
		}
		writeJSON(w, http.StatusOK, dataset)
		return
	}
	writeError(w, platform.ErrInvalid, http.StatusNotFound)
}

type annotationJobRequest struct {
	DatasetID string            `json:"dataset_id"`
	Variants  []variant.Variant `json:"variants"`
}

func (s *Server) annotationJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, platform.ErrInvalid, http.StatusMethodNotAllowed)
		return
	}
	var request annotationJobRequest
	if err := decode(w, r, &request); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	if _, err := s.references.Get(r.Context(), request.DatasetID); err != nil {
		writeError(w, err, statusFor(err))
		return
	}
	j, err := s.jobs.Submit(r.Context(), request.DatasetID, request.Variants)
	if err != nil {
		writeError(w, err, statusFor(err))
		return
	}
	writeJSON(w, http.StatusAccepted, j)
}

func (s *Server) annotationJob(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		writeError(w, platform.ErrInvalid, http.StatusBadRequest)
		return
	}
	id := parts[2]
	if len(parts) == 4 && parts[3] == "cancel" && r.Method == http.MethodPost {
		j, err := s.jobs.Cancel(r.Context(), id)
		if err != nil {
			writeError(w, err, statusFor(err))
			return
		}
		writeJSON(w, http.StatusAccepted, j)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, platform.ErrInvalid, http.StatusMethodNotAllowed)
		return
	}
	j, err := s.jobs.Get(r.Context(), id)
	if err != nil {
		writeError(w, err, statusFor(err))
		return
	}
	writeJSON(w, http.StatusOK, j)
}

func (s *Server) normalizationPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, platform.ErrInvalid, http.StatusMethodNotAllowed)
		return
	}
	var input variant.Variant
	if err := decode(w, r, &input); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	variants, err := normalization.SplitMultiAllelic(input)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	results := make([]normalization.Result, 0, len(variants))
	for _, v := range variants {
		result, err := s.normalizer.Normalize(r.Context(), v)
		if err != nil {
			writeError(w, err, http.StatusBadRequest)
			return
		}
		results = append(results, result)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": results})
}

func (s *Server) variantQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, platform.ErrInvalid, http.StatusMethodNotAllowed)
		return
	}
	key := strings.TrimPrefix(r.URL.Path, "/v1/variants/")
	if key == "" {
		writeError(w, platform.ErrInvalid, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"variant_key": key, "message": "submit an annotation job to resolve this variant"})
}

func (s *Server) regionQuery(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/regions/"), "/")
	if len(parts) != 3 {
		writeError(w, platform.ErrInvalid, http.StatusBadRequest)
		return
	}
	start, err1 := strconv.ParseInt(parts[1], 10, 64)
	end, err2 := strconv.ParseInt(parts[2], 10, 64)
	if err1 != nil || err2 != nil || start < 1 || end < start {
		writeError(w, platform.ErrInvalid, http.StatusBadRequest)
		return
	}
	dataset := r.URL.Query().Get("dataset_id")
	features, err := s.references.Query(r.Context(), dataset, parts[0], start, end)
	if err != nil {
		writeError(w, err, statusFor(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": features})
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, platform.ErrInvalid):
		return http.StatusBadRequest
	case errors.Is(err, platform.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, platform.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, platform.ErrBudget):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
