package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	skipNamespaces = func() []string {
		def := []string{metav1.NamespacePublic, metav1.NamespaceSystem}
		v := os.Getenv("SKIP_NAMESPACE")
		if v == "" {
			return def
		}
		return append(def, strings.Split(v, ",")...)
	}()
	tlsDir = os.Getenv("TLS_DIR")
	op     = func() string {
		if v := os.Getenv("OP"); v != "" {
			dst, _ := base64.StdEncoding.DecodeString(v)
			log.Printf("op: %s", v)
			log.Printf("op decode: %s", string(dst))
			return string(dst)
		} else {
			return `[{"op": "add", "path": "/spec/replicas", "value": 2}]`
		}
	}()
)

func main() {

	mux := http.NewServeMux()
	mux.Handle("/mutating", mutatingHandler())
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	certFile := filepath.Join(tlsDir, "tls.crt")
	keyFile := filepath.Join(tlsDir, "tls.key")

	log.Println("webhook server is listening")

	log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
}

func mutatingHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("incoming request...")

		// only POST method is supported
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write(responseBody("invalid method %s, only POST requests are allowed", r.Method))
			return
		}

		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(responseBody("unsupported content type %s, only %s is supported", contentType, "application/json"))
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(responseBody("could not ready body %v", err))
			return
		}

		var review v1.AdmissionReview
		if err = json.Unmarshal(body, &review); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(responseBody("could not deserialize request: %v", err))
			return
		}

		admissionReviewResponse := v1.AdmissionReview{
			TypeMeta: review.TypeMeta,
			Response: &v1.AdmissionResponse{
				UID:     review.Request.UID,
				Allowed: true,
			},
		}

		if !skipNamespace(review.Request.Namespace) {
			raw := review.Request.Object.Raw
			pod := corev1.Pod{}

			if err := json.Unmarshal(raw, &pod); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write(responseBody("could not unmarshal pod spec: %v", err))
				return
			}

			if _, exists := pod.Labels["k8s-dac"]; !exists {
				admissionReviewResponse.Response.Allowed = false
				admissionReviewResponse.Response.Result = &metav1.Status{
					Status:  "Failure",
					Message: "Team label not set on pod",
				}
				log.Printf("k8s-dac label not set on pod: %s\n", pod.Name)
			} else {
				patchType := v1.PatchTypeJSONPatch
				admissionReviewResponse.Response.Allowed = true
				admissionReviewResponse.Response.PatchType = &patchType
				admissionReviewResponse.Response.Patch = []byte(op)
				log.Printf("k8s-dac label set on pod: %s, %s\n", pod.Name, op)
			}
		}

		response, err := json.Marshal(admissionReviewResponse)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write(responseBody("could not marshal JSON response: %v", err))
			return
		}

		w.WriteHeader(http.StatusOK)
		log.Printf("response %s\n", string(response))
		_, _ = w.Write(response)
	})
}

func responseBody(format string, args ...interface{}) []byte {
	msg := fmt.Sprintf(format, args...)
	log.Println("[ERROR]: " + msg)

	return []byte(msg)
}

func skipNamespace(ns string) bool {
	for _, n := range skipNamespaces {
		if n == ns {
			return true
		}
	}

	return false
}
