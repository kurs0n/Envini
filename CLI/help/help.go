package help

import "fmt"

func DisplayHelp() {
	helpText := `Envini CLI Help
Usage: envini [command] [options]
Commands:
  auth                Authenticate with GitHub
  list repos         List your repositories		
  list environments  List environments in a repository
  upload [file]       Upload a file to a repository
Options:
  --help, -h         Show this help message
  --version, -v      Show version information
Examples:
  envini auth
  envini list repos
  envini list environments
  envini upload .env
`
	fmt.Println(helpText)
}
