package help

import "fmt"

func DisplayHelp() {
	helpText := `Envini CLI Help
Usage: envini [command] [options]

Commands:
  auth                                              Authenticate with GitHub
  repos                                            List your repositories		
  repos-with-versions                              List your repositories with all secret versions
  upload <file> [--tag=development]                Upload a secret file (auto-detects repo)
  upload <owner> <repo> <file> [--tag=development] Upload with explicit repo
  download <output> [--version=latest] [--tag=tag]  Download version (auto-detects repo)
  download <owner> <repo> [output] [--version=latest] [--tag=tag] Download with explicit repo
  delete [--version=latest] [--tag=tag]            Delete version (auto-detects repo)
  delete <owner> <repo> [--version=latest] [--tag=tag] Delete with explicit repo
  versions                                         List all versions (auto-detects repo)
  versions <owner> <repo>                          List versions with explicit repo

Options:
  --tag=value        Specify tag for upload/download/delete (default: development for latest operations)
  --version=value    Specify version number or 'latest' (default: latest)

Notes:
  • Auto-detection uses the current git repository's remote origin URL
  • --version=latest or --version omitted downloads/deletes the latest from specified tag (default: development)
  • --version=N (number) targets specific version number
  • --tag=tagname downloads/deletes the latest version from that specific tag
  • You can combine --version and --tag for precise targeting
  • Upload always creates new versions with specified tag
  • Different tags maintain separate version sequences

Examples:
  # Authentication and listing
  envini auth
  envini repos
  envini repos-with-versions
  
  # Auto-detect repository from git
  envini upload .env                              # Upload to development tag
  envini upload .env --tag=production             # Upload to production tag
  envini download .env.downloaded                 # Download latest from development tag
  envini download .env.downloaded --tag=production # Download latest from production tag
  envini download .env.downloaded --version=1     # Download specific version
  envini download .env.downloaded --version=2 --tag=production # Download version 2 from production tag
  envini delete                                   # Delete latest from development tag
  envini delete --tag=production                  # Delete latest from production tag
  envini delete --version=1                       # Delete specific version
  envini versions                                 # List all versions
  
  # Explicit repository specification
  envini upload kurs0n 8080-emulator .env
  envini upload kurs0n 8080-emulator .env --tag=production
  envini download kurs0n 8080-emulator .env.downloaded
  envini download kurs0n 8080-emulator .env.downloaded --tag=production
  envini download kurs0n 8080-emulator .env.downloaded --version=1
  envini delete kurs0n 8080-emulator
  envini delete kurs0n 8080-emulator --tag=production
  envini delete kurs0n 8080-emulator --version=1
  envini versions kurs0n 8080-emulator
`
	fmt.Println(helpText)
}
