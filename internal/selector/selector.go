package selector

import (
	"math/rand"
	"time"
)

var topics = [][]string{
	{"Kubernetes", "K8s", "controller runtime", "Ingress", "CNI", "eBPF"},
	{"CI/CD", "GitHub Actions", "GitLab CI", "Tekton", "Drone", "Argo Workflows"},
	{"SRE", "SLI/SLO", "error budgets", "incident response", "postmortems"},
	{"IaC", "Terraform", "Pulumi", "OpenTofu", "drift", "policy as code"},
	{"Observability", "OpenTelemetry", "tracing", "metrics", "logs", "profiling"},
	{"Service Mesh", "Istio", "Linkerd", "mTLS", "zero-trust"},
	{"Security", "SBOM", "supply chain", "SLSA", "cosign", "OPA", "gitleaks"},
	{"Cloud", "GKE", "EKS", "AKS", "serverless", "FinOps"},
}

var styles = []string{
	"punchy, 1-2 lines, no hashtags, use a rhetorical hook",
	"curious, conversational, one emoji allowed, avoid buzzwords",
	"mini-tip with a quick example, newline for readability",
	"myth-busting tone, cite a common misconception and a fix",
	"micro-story: problem, constraint, clever workaround",
}

func init() { rand.Seed(time.Now().UnixNano()) }

func RandomTopicSet() []string { return topics[rand.Intn(len(topics))] }
func RandomStyle() string      { return styles[rand.Intn(len(styles))] }
