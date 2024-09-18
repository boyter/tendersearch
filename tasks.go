// Tasks is a collection of tasks that would normally be done through shell scripts, but since it can be a real
// pain a lot of the time they have been put in here. Usually it does things like running smoke tests or validating
// unique codes

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

func main() {
	var smokeTest, checkUnique, confirmUnique bool

	flag.BoolVar(&smokeTest, "smoke-test", false, "builds a copy of the api locally and smoke tests it")
	flag.BoolVar(&checkUnique, "check-unique", false, "check if unique codes are actually unique")
	flag.BoolVar(&confirmUnique, "confirm-unique", false, "check if unique codes are present for all slog calls")
	flag.Parse()

	ran := false
	if smokeTest {
		runSmokeTest()
		ran = true
	}
	if checkUnique || confirmUnique {
		checkUniqueCodes(checkUnique, confirmUnique)
		ran = true
	}

	// if nothing specified print the usage
	if !ran {
		flag.PrintDefaults()
	}
}

func checkUniqueCodes(checkUnique, confirmUnique bool) {
	fileExclusions := []string{"tasks.go", "assets/smoke-tests/smoke_test.go"}
	denyListPrefix := []string{"vendor/", ".git/"}
	missingUnique := []uniqueCodeMatch{}
	uniqueCodes := map[string][]uniqueCodeMatch{}

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		for _, prefix := range denyListPrefix {
			if strings.HasPrefix(path, prefix) {
				return nil
			}
		}

		for _, exclusion := range fileExclusions {
			if path == exclusion {
				return nil
			}
		}

		file, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		uc, mu, err := identifyUnique(info.Name(), path, string(file))
		if err != nil {
			return err
		}
		missingUnique = append(missingUnique, mu...)
		for k, v := range uc {
			uniqueCodes[k] = append(uniqueCodes[k], v...)
		}

		return nil
	})
	if err != nil {
		return
	}

	osExit := false
	if confirmUnique {
		if len(missingUnique) != 0 {
			fmt.Println("slog entries missing unique codes found")
			osExit = true
		}
		for _, v := range missingUnique {
			fmt.Printf("\tline %5d. %v \n", v.Line, v.Path)
		}
	}

	if checkUnique {
		for k, v := range uniqueCodes {
			if len(v) > 1 {
				fmt.Printf("unique code: %v found %v times\n", k, len(v))
				for _, m := range v {
					fmt.Printf("\tline %5d. in %v\n", m.Line, m.Path)
				}
				osExit = true
			}
		}
	}

	if osExit {
		os.Exit(1)
	}
}

type uniqueCodeMatch struct {
	Path string
	Line int
}

func identifyUnique(filename, path, code string) (map[string][]uniqueCodeMatch, []uniqueCodeMatch, error) {
	uniqueCodes := map[string][]uniqueCodeMatch{}
	missingUnique := []uniqueCodeMatch{}
	c := regexp.MustCompile("\"[a-z0-9]+\"")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, code, parser.AllErrors)
	if err != nil {
		return uniqueCodes, missingUnique, err
	}
	matches := []string{"Info", "Warn", "Error", "Debug"}

	// Walk the AST and pull out the unique codes
	ast.Inspect(f, func(n ast.Node) bool {
		// Find call expressions
		if callExpr, ok := n.(*ast.CallExpr); ok {
			// Check if the call expression is a slog.Info call
			if fun, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := fun.X.(*ast.Ident); ok {
					if ident.Name == "slog" && slices.Contains(matches, fun.Sel.Name) {
						foundUniqueCode := false
						linePos := 0
						for _, arg := range callExpr.Args {
							linePos = fset.Position(arg.Pos()).Line
							val := code[arg.Pos()-1 : arg.End()]
							if strings.HasPrefix(val, "UC") || strings.HasPrefix(val, "common.UC") {
								for _, uc := range c.FindAllString(val, 1) {
									uniqueCodes[uc] = append(uniqueCodes[uc], uniqueCodeMatch{
										Path: path,
										Line: fset.Position(arg.Pos()).Line,
									})
								}
								foundUniqueCode = true
							}
						}
						if !foundUniqueCode {
							missingUnique = append(missingUnique, uniqueCodeMatch{
								Path: path,
								Line: linePos,
							})
						}
					}
				}
			}
		}
		return true
	})

	return uniqueCodes, missingUnique, nil
}

const (
	defaultPort = "4001"
	maxAttempts = 10
)

// builds a copy of the api locally and smoke tests it
func runSmokeTest() {
	port := getEnv("HTTP_SERVER_PORT", defaultPort)
	fmt.Println("Building application and starting API...")

	cmd := exec.Command("go", "build", "-o", "tendersearch-smoke", ".")
	cmd.Dir = "./cmd/tendersearch/"
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error building application: %v\n", err)
		os.Exit(1)
	}

	cmd = exec.Command("./tendersearch-smoke")
	cmd.Dir = "./cmd/tendersearch/"
	cmd.Env = append(os.Environ(), "HTTP_SERVER_PORT="+port)
	cmd.Stdout = nil
	cmd.Stderr = nil

	err = cmd.Start()
	if err != nil {
		fmt.Printf("Error starting application: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		err = cmd.Process.Kill()
		if err != nil {
			fmt.Println("Error killing process, run 'pkill api-web-smoke' manually.")
		}
	}()

	fmt.Println("Waiting for application to startup...")
	waitForServer(port)

	fmt.Println("Running local smoke tests...")
	runTests()
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func waitForServer(port string) {
	attempts := 0

	for {
		attempts++
		fmt.Printf("Attempt %d...\n", attempts)
		resp, err := http.Get("http://localhost:" + port + "/v1/healthcheck")
		if err == nil {
			fmt.Println("Server has finished startup process commencing smoke tests")
			_ = resp.Body.Close()
			break
		} else {
			time.Sleep(200 * time.Millisecond)
		}

		if attempts >= maxAttempts {
			fmt.Println("Server has not started, expect smoke tests to fail")
			break
		}
	}
}

func runTests() {
	cmd := exec.Command("go", "test", "-count=1", "-tags=smoke", "./assets/smoke-tests/...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error running smoke tests: %v\n", err)
	}
}
