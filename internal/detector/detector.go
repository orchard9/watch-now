package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/orchard9/watch-now/internal/config"
)

type ProjectDetector struct {
	projectPath string
}

type ProjectInfo struct {
	Type             string
	Services         []config.ServiceConfig
	QualityChecks    []config.CheckConfig
	HasMakefile      bool
	HasPackageJSON   bool
	HasGoMod         bool
	HasDockerCompose bool
	DetectedPorts    []int
}

func NewProjectDetector(path string) *ProjectDetector {
	return &ProjectDetector{
		projectPath: path,
	}
}

func (d *ProjectDetector) DetectProject() (*ProjectInfo, error) {
	info := &ProjectInfo{
		Services:      []config.ServiceConfig{},
		QualityChecks: []config.CheckConfig{},
	}

	// Detect project files
	info.HasMakefile = d.fileExists("Makefile")
	info.HasPackageJSON = d.fileExists("package.json")
	info.HasGoMod = d.fileExists("go.mod")
	info.HasDockerCompose = d.fileExists("docker-compose.yml") || d.fileExists("docker-compose.yaml")

	// Determine project type
	info.Type = d.determineProjectType(info)

	// Generate quality checks based on project type
	info.QualityChecks = d.generateQualityChecks(info)

	// Try to detect services (if it looks like a service-oriented project)
	if d.looksLikeServiceProject() {
		info.Services = d.detectServices()
	}

	return info, nil
}

func (d *ProjectDetector) fileExists(filename string) bool {
	_, err := os.Stat(filepath.Join(d.projectPath, filename))
	return !os.IsNotExist(err)
}

func (d *ProjectDetector) determineProjectType(info *ProjectInfo) string {
	// Check for monorepo patterns first
	if d.isMonorepo() {
		return "monorepo"
	}

	// Check for specific language indicators
	return d.detectLanguage(info)
}

func (d *ProjectDetector) isMonorepo() bool {
	return d.hasDirectories([]string{"backend", "frontend"}) ||
		d.hasDirectories([]string{"services"}) ||
		d.hasDirectories([]string{"apps", "packages"})
}

func (d *ProjectDetector) detectLanguage(info *ProjectInfo) string {
	if info.HasGoMod {
		return "go"
	}
	if info.HasPackageJSON {
		return "node"
	}
	if d.fileExists("requirements.txt") || d.fileExists("pyproject.toml") {
		return "python"
	}
	if d.fileExists("pom.xml") || d.fileExists("build.gradle") {
		return "java"
	}
	if d.fileExists("Cargo.toml") {
		return "rust"
	}
	return "unknown"
}

func (d *ProjectDetector) hasDirectories(dirs []string) bool {
	for _, dir := range dirs {
		if stat, err := os.Stat(filepath.Join(d.projectPath, dir)); err == nil && stat.IsDir() {
			return true
		}
	}
	return false
}

func (d *ProjectDetector) looksLikeServiceProject() bool {
	// Look for service directories
	servicesDir := filepath.Join(d.projectPath, "services")
	if stat, err := os.Stat(servicesDir); err == nil && stat.IsDir() {
		return true
	}

	// Look for backend/services
	backendServicesDir := filepath.Join(d.projectPath, "backend", "services")
	if stat, err := os.Stat(backendServicesDir); err == nil && stat.IsDir() {
		return true
	}

	return false
}

func (d *ProjectDetector) detectServices() []config.ServiceConfig {
	services := []config.ServiceConfig{}

	// Check backend/services directory (acecam style)
	backendServicesDir := filepath.Join(d.projectPath, "backend", "services")
	if stat, err := os.Stat(backendServicesDir); err == nil && stat.IsDir() {
		services = append(services, d.scanServicesDirectory(backendServicesDir, d.guessAcecamPorts)...)
	}

	// Check services directory
	servicesDir := filepath.Join(d.projectPath, "services")
	if stat, err := os.Stat(servicesDir); err == nil && stat.IsDir() {
		services = append(services, d.scanServicesDirectory(servicesDir, d.guessStandardPorts)...)
	}

	return services
}

func (d *ProjectDetector) scanServicesDirectory(servicesDir string, portGuesser func(string, int) int) []config.ServiceConfig {
	services := []config.ServiceConfig{}

	entries, err := os.ReadDir(servicesDir)
	if err != nil {
		return services
	}

	portOffset := 0
	for _, entry := range entries {
		if entry.IsDir() {
			serviceName := entry.Name()
			port := portGuesser(serviceName, portOffset)

			service := config.ServiceConfig{
				Name:    serviceName,
				Type:    "rest",
				URL:     fmt.Sprintf("http://localhost:%d", port),
				Health:  "/healthz", // Default to /healthz (acecam style)
				Timeout: 5 * time.Second,
			}

			services = append(services, service)
			portOffset++
		}
	}

	return services
}

func (d *ProjectDetector) guessAcecamPorts(serviceName string, offset int) int {
	// acecam uses specific port patterns
	servicePortMap := map[string]int{
		"iam":          35050,
		"social":       35052,
		"analytics":    35054,
		"gaming":       35056,
		"notification": 35058,
		"logging":      35062,
	}

	if port, exists := servicePortMap[serviceName]; exists {
		return port
	}

	// Fallback to calculated port
	return 35000 + (offset * 2)
}

func (d *ProjectDetector) guessStandardPorts(serviceName string, offset int) int {
	// Standard microservice ports starting at 8080
	return 8080 + offset
}

func (d *ProjectDetector) generateQualityChecks(info *ProjectInfo) []config.CheckConfig {
	checks := []config.CheckConfig{}

	if info.HasMakefile {
		// Check what make targets are available
		makeTargets := d.detectMakeTargets()

		// Add common quality checks if targets exist
		commonChecks := []string{"fmt", "format", "lint", "test", "complexity", "deadcode", "docs"}
		for _, check := range commonChecks {
			if d.containsString(makeTargets, check) {
				checks = append(checks, config.CheckConfig{
					Name:    check,
					Command: "make",
					Args:    []string{check},
					Timeout: time.Duration(d.getTimeoutForCheck(check)),
				})
			}
		}
	} else {
		// Generate checks based on project type
		switch info.Type {
		case "go":
			checks = append(checks, d.generateGoChecks()...)
		case "node":
			checks = append(checks, d.generateNodeChecks()...)
		case "python":
			checks = append(checks, d.generatePythonChecks()...)
		}
	}

	return checks
}

func (d *ProjectDetector) detectMakeTargets() []string {
	// This is a simple implementation - in practice you'd parse the Makefile
	// For now, return common targets that are likely to exist
	return []string{"fmt", "lint", "test", "build", "clean", "complexity", "deadcode", "docs"}
}

func (d *ProjectDetector) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (d *ProjectDetector) getTimeoutForCheck(check string) int64 {
	timeouts := map[string]int64{
		"fmt":        30000000000,  // 30s
		"format":     30000000000,  // 30s
		"lint":       60000000000,  // 60s
		"test":       120000000000, // 120s
		"build":      180000000000, // 180s
		"complexity": 30000000000,  // 30s
		"deadcode":   30000000000,  // 30s
		"docs":       60000000000,  // 60s
	}

	if timeout, exists := timeouts[check]; exists {
		return timeout
	}
	return 30000000000 // Default 30s
}

func (d *ProjectDetector) generateGoChecks() []config.CheckConfig {
	return []config.CheckConfig{
		{Name: "format", Command: "gofmt", Args: []string{"-l", "."}, Timeout: 30 * time.Second},
		{Name: "test", Command: "go", Args: []string{"test", "./..."}, Timeout: 120 * time.Second},
		{Name: "build", Command: "go", Args: []string{"build", "./..."}, Timeout: 180 * time.Second},
	}
}

func (d *ProjectDetector) generateNodeChecks() []config.CheckConfig {
	checks := []config.CheckConfig{}

	// Check for common npm/yarn scripts
	if d.fileExists("package.json") {
		checks = append(checks,
			config.CheckConfig{Name: "lint", Command: "npm", Args: []string{"run", "lint"}, Timeout: 60 * time.Second},
			config.CheckConfig{Name: "test", Command: "npm", Args: []string{"test"}, Timeout: 120 * time.Second},
			config.CheckConfig{Name: "build", Command: "npm", Args: []string{"run", "build"}, Timeout: 180 * time.Second},
		)
	}

	return checks
}

func (d *ProjectDetector) generatePythonChecks() []config.CheckConfig {
	checks := []config.CheckConfig{}

	checks = append(checks,
		config.CheckConfig{Name: "format", Command: "black", Args: []string{"--check", "."}, Timeout: 30 * time.Second},
		config.CheckConfig{Name: "lint", Command: "flake8", Args: []string{"."}, Timeout: 60 * time.Second},
		config.CheckConfig{Name: "test", Command: "pytest", Args: []string{}, Timeout: 120 * time.Second},
	)

	return checks
}

func (d *ProjectDetector) GenerateConfig() *config.Config {
	info, _ := d.DetectProject()

	cfg := &config.Config{
		Services: info.Services,
		Checks:   info.QualityChecks,
		Interval: 30 * time.Second,
		API: config.APIConfig{
			Enabled: true,
			Port:    0, // Use ephemeral port
		},
	}

	return cfg
}
