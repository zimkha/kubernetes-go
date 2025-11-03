package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main()  {
   // call endpoint validate
   http.HandleFunc("/validate", handleValidate)
   fmt.Println("🚀 Webhook server running on :8443")
   	if err := http.ListenAndServeTLS(":8443", "/tls/tls.crt", "/tls/tls.key", nil); err != nil {
		panic(err)
	}
}

func validateLabels(namespace string, labels map[string] string) (bool , string) {
	env, ok := labels["env"]
	if !ok {
		return false, fmt.Sprintf("Missing label 'env' for namespace '%s'", namespace)
	}
	if env != namespace {
		return false, fmt.Sprintf("Invalid env label  expected '%s' got '%s'", namespace,env)
	}
	return true, "OK"
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not read request", http.StatusBadRequest)
		return
	}

	var review admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &review); err != nil {
		http.Error(w, "Could not decode admission review", http.StatusBadRequest)
		return
	}

	reviewResponse := admissionv1.AdmissionResponse{
		UID: review.Request.UID,
	}

	allowed := true
	message := "Validation passed"

	namespace := review.Request.Namespace

	switch review.Request.Kind.Kind {
	case "Pod":
		var pod corev1.Pod
		if err := json.Unmarshal(review.Request.Object.Raw, &pod); err != nil {
			allowed, message = logUnmarshalError("Pod", err)
		} else {
			log.Printf("📋 Pod name: %s", pod.Name)
			log.Printf("🏷️  Pod labels: %+v", pod.ObjectMeta.Labels)
			
			// ✅ CORRECTION: Utiliser pod.ObjectMeta.Labels
			allowed, message = validateLabels(namespace, pod.ObjectMeta.Labels)
		}
	case "Deployment":
		var deploy appsv1.Deployment
		if err := json.Unmarshal(review.Request.Object.Raw, &deploy); err != nil {
			allowed, message = logUnmarshalError("Deployment", err)
		} else {
			log.Printf("📋 Deployment name: %s", deploy.Name)
			log.Printf("🏷️  Deployment.metadata.labels: %+v", deploy.ObjectMeta.Labels)
			log.Printf("🎯 Deployment.spec.selector.matchLabels: %+v", deploy.Spec.Selector.MatchLabels)
			log.Printf("📦 Deployment.spec.template.metadata.labels: %+v", deploy.Spec.Template.ObjectMeta.Labels)
			
			// ✅ CORRECTION: Utiliser deploy.ObjectMeta.Labels (pas deploy.Labels)
			allowed, message = validateLabels(namespace, deploy.ObjectMeta.Labels)
		}
	case "Service": 
		var service corev1.Service
		if err :=json.Unmarshal(review.Request.Object.Raw, service); err != nil {
			allowed, message = logUnmarshalError("Service", err)
		} else {
			allowed, message = validateLabels(namespace, service.ObjectMeta.Labels)
		}

	default:
		log.Printf("⏭️  Unsupported kind: %s - allowing by default", review.Request.Kind.Kind)
		message = fmt.Sprintf("Unsupported kind: %s", review.Request.Kind.Kind)
		allowed = true // Autoriser les types non supportés
	}

	reviewResponse.Allowed = allowed
	if !allowed {
		reviewResponse.Result = &metav1.Status{
			Message: message,
			Code:    http.StatusForbidden,
		}
	}

	response := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &reviewResponse,
	}

	// Envoyer la réponse
	respBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("❌ Error marshaling response: %v", err)
		http.Error(w, "Could not encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}

func logUnmarshalError(kind string, err error) (allowed bool, message string) {
    log.Printf("❌ Error unmarshaling %s: %v", kind, err)
    allowed = false
    message = fmt.Sprintf("Error parsing %s: %v", kind, err)
    return
}