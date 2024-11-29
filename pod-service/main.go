package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Server struct {
	clientset  *kubernetes.Clientset
	podCounter int
	mu         sync.Mutex
}

func main() {
	// Set up in-cluster Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to create in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes clientset: %v", err)
	}

	server := &Server{
		clientset: clientset,
	}

	app := fiber.New()

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
