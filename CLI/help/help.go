package help

import "fmt"

func DisplayHelp() {
	helpText := `Envini CLI Help
Usage: envini [command] [options]
Commands:
  auth                Authenticate with GitHub
  repos              List your repositories		
  repos-with-versions List your repositories with all secret versions
  upload <file> [--tag=development] Upload a secret file (auto-detects repo)
  upload <owner> <repo> <file> [--tag=development] Upload with explicit repo
  download <output> [--version=latest] Download latest version (auto-detects repo)
  download <owner> <repo> [output] [--version=latest] Download with explicit repo
  delete [--version=latest] Delete latest version (auto-detects repo)
  delete <owner> <repo> [--version=latest] Delete with explicit repo
  versions List all versions (auto-detects repo)
  versions <owner> <repo> List versions with explicit repo
Options:
  --help, -h         Show this help message
  --version, -v      Show version information
  --tag=value        Specify tag for upload (default: development)
  --version=value    Specify version number or 'latest' (default: latest)
Notes:
  • Version numbers are tag-specific. Uploading to different tags creates separate version sequences.
  • Example: v1(development), v2(development), v1(production), v2(production)
Examples:
  envini auth
  envini repos
  envini repos-with-versions
  
  # Auto-detect repository from git
  envini upload .env
  envini upload .env --tag=production
  envini download .env.downloaded
  envini download .env.downloaded --version=1
  envini delete
  envini delete --version=1
  envini versions
  
  # Explicit repository
  envini upload kurs0n 8080-emulator .env
  envini upload kurs0n 8080-emulator .env --tag=production
  envini download kurs0n 8080-emulator .env.downloaded
  envini download kurs0n 8080-emulator .env.downloaded --version=1
  envini delete kurs0n 8080-emulator
  envini delete kurs0n 8080-emulator --version=1
  envini versions kurs0n 8080-emulator
`
	fmt.Println(helpText)
}
