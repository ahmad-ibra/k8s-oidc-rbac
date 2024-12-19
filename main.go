package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Server struct {
	defaultClientset *kubernetes.Clientset
	clientset        *kubernetes.Clientset
	podCounter       int
	mu               sync.Mutex
}

func main() {
	// Set up in-cluster Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to create in-cluster config: %v", err)
	}

	defaultClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes clientset: %v", err)
	}

	server := &Server{
		defaultClientset: defaultClientset,
	}

	app := fiber.New()

	// Middleware to handle impersonation
	app.Use(server.impersonationMiddleware(config))

	// Define API routes
	app.Post("/api/v1/pods", server.createPod)
	app.Get("/api/v1/pods", server.listPods)
	app.Get("/api/v1/pods/:name", server.getPod)

	// Start the server
	log.Println("Starting server on :8000")
	if err := app.Listen(":8000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// impersonationMiddleware sets impersonation headers in the Kubernetes client.
func (s *Server) impersonationMiddleware(baseConfig *rest.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// No impersonation; continue request handling
			s.clientset = s.defaultClientset
			return c.Next()
		}

		// Extract the token from the Authorization header
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Authorization header format",
			})
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Decode the JWT to extract user information
		sub, groups, err := decodeToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to decode token: %v", err),
			})
		}

		// Create a new config with the impersonation settings
		impersonationConfig := rest.CopyConfig(baseConfig)
		impersonationConfig.BearerToken = token
		impersonationConfig.Impersonate = rest.ImpersonationConfig{
			UserName: sub,
			Groups:   groups,
		}

		// Update the clientset with the new impersonation config
		impersonatedClientset, err := kubernetes.NewForConfig(impersonationConfig)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create impersonation clientset: %v", err),
			})
		}

		// Replace the clientset in the server with the impersonation client
		s.clientset = impersonatedClientset

		return c.Next()
	}
}

// decodeToken parses the JWT token to extract the email and groups.
func decodeToken(tokenString string) (string, []string, error) {
	// Parse the token without verifying the signature (Okta should handle validation).
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", nil, fmt.Errorf("invalid token claims format")
	}

	// Extract email (or username)
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", nil, fmt.Errorf("sub claim missing or invalid")
	}

	// Extract groups
	var groups []string
	if rawGroups, ok := claims["groups"].([]interface{}); ok {
		for _, g := range rawGroups {
			if group, ok := g.(string); ok {
				groups = append(groups, group)
			}
		}
	}

	return sub, groups, nil
}

// createPod handles POST /api/v1/pods
func (s *Server) createPod(c *fiber.Ctx) error {
	s.mu.Lock()
	s.podCounter++
	podName := fmt.Sprintf("pod%d", s.podCounter)
	s.mu.Unlock()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "noop",
					Image: "busybox",
					Command: []string{
						"/bin/sh",
						"-c",
						"while true; do echo 'noop'; sleep 10; done",
					},
				},
			},
		},
	}

	_, err := s.clientset.CoreV1().Pods("default").Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create pod: %v", err),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": fmt.Sprintf("Pod %s created", podName),
	})
}

// listPods handles GET /api/v1/pods
func (s *Server) listPods(c *fiber.Ctx) error {
	pods, err := s.clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to list pods: %v", err),
		})
	}

	podNames := []string{}
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)
	}

	return c.JSON(fiber.Map{
		"pods": podNames,
	})
}

// getPod handles GET /api/v1/pods/:name
func (s *Server) getPod(c *fiber.Ctx) error {
	podName := c.Params("name")

	pod, err := s.clientset.CoreV1().Pods("default").Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Pod %s not found: %v", podName, err),
		})
	}

	return c.JSON(fiber.Map{
		"name":      pod.Name,
		"namespace": pod.Namespace,
		"status":    pod.Status.Phase,
	})
}
