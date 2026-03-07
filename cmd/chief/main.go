package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/minicodemonkey/chief/internal/cmd"
	"github.com/minicodemonkey/chief/internal/config"
	"github.com/minicodemonkey/chief/internal/git"
	"github.com/minicodemonkey/chief/internal/prd"
	"github.com/minicodemonkey/chief/internal/tui"
)

// Version is set at build time via ldflags
var Version = "dev"

// TUIOptions holds the parsed command-line options for the TUI
type TUIOptions struct {
	PRDPath       string
	MaxIterations int
	Verbose       bool
	Merge         bool
	Force         bool
	NoRetry       bool
	PromptsDir    string
}

// findSubcmd returns the first non-flag argument in os.Args, skipping known
// value-taking flags and their values. Returns "" if no subcommand is found.
func findSubcmd() string {
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		// Value-taking flags: skip both the flag and its value.
		if arg == "--prompts-dir" || arg == "--max-iterations" || arg == "-n" {
			i++ // skip value
			continue
		}
		// = form of value-taking flags: no extra value to skip.
		if strings.HasPrefix(arg, "--prompts-dir=") ||
			strings.HasPrefix(arg, "--max-iterations=") ||
			strings.HasPrefix(arg, "-n=") {
			continue
		}
		// Any other flag: skip (boolean flags).
		if strings.HasPrefix(arg, "-") {
			continue
		}
		// First non-flag argument is the subcommand or PRD name.
		return arg
	}
	return ""
}

// extractGlobalPromptsDir scans os.Args for --prompts-dir and returns the
// validated directory path. Exits with a clear message if the path is invalid.
// Returns "" if the flag is absent.
func extractGlobalPromptsDir() string {
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--prompts-dir" {
			if i+1 < len(os.Args) {
				dir := os.Args[i+1]
				info, err := os.Stat(dir)
				if err != nil || !info.IsDir() {
					fmt.Fprintf(os.Stderr, "prompts directory not found: %s\n", dir)
					os.Exit(1)
				}
				return dir
			}
			fmt.Fprintf(os.Stderr, "Error: --prompts-dir requires a value\n")
			os.Exit(1)
		}
		if strings.HasPrefix(arg, "--prompts-dir=") {
			dir := strings.TrimPrefix(arg, "--prompts-dir=")
			info, err := os.Stat(dir)
			if err != nil || !info.IsDir() {
				fmt.Fprintf(os.Stderr, "prompts directory not found: %s\n", dir)
				os.Exit(1)
			}
			return dir
		}
	}
	return ""
}

func main() {
	// Route to subcommands, skipping any leading global flags so that
	// `chief --prompts-dir /foo new` correctly reaches runNew().
	switch findSubcmd() {
	case "new":
		runNew()
		return
	case "edit":
		runEdit()
		return
	case "status":
		runStatus()
		return
	case "list":
		runList()
		return
	case "help":
		printHelp()
		return
	case "update":
		runUpdate()
		return
	case "config":
		runConfig()
		return
	case "init-prompts":
		runInitPrompts()
		return
	case "wiggum":
		printWiggum()
		return
	}

	// Non-blocking version check on startup (for interactive TUI sessions)
	cmd.CheckVersionOnStartup(Version)

	// Parse flags for TUI mode
	opts := parseTUIFlags()

	// Handle special flags that were parsed
	if opts == nil {
		// Already handled (--help or --version)
		return
	}

	// Run the TUI
	runTUIWithOptions(opts)
}

// findAvailablePRD looks for any available PRD in .chief/prds/
// Returns the path to the first PRD found, or empty string if none exist.
func findAvailablePRD() string {
	prdsDir := ".chief/prds"
	entries, err := os.ReadDir(prdsDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			prdPath := filepath.Join(prdsDir, entry.Name(), "prd.json")
			if _, err := os.Stat(prdPath); err == nil {
				return prdPath
			}
		}
	}
	return ""
}

// listAvailablePRDs returns all PRD names in .chief/prds/
func listAvailablePRDs() []string {
	prdsDir := ".chief/prds"
	entries, err := os.ReadDir(prdsDir)
	if err != nil {
		return nil
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			prdPath := filepath.Join(prdsDir, entry.Name(), "prd.json")
			if _, err := os.Stat(prdPath); err == nil {
				names = append(names, entry.Name())
			}
		}
	}
	return names
}

// parseTUIFlags parses command-line flags for TUI mode
func parseTUIFlags() *TUIOptions {
	opts := &TUIOptions{
		PRDPath:       "", // Will be resolved later
		MaxIterations: 0,  // 0 signals dynamic calculation (remaining stories + 5)
		Verbose:       false,
		Merge:         false,
		Force:         false,
		NoRetry:       false,
		PromptsDir:    "",
	}

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		switch {
		case arg == "--help" || arg == "-h":
			printHelp()
			return nil
		case arg == "--version" || arg == "-v":
			fmt.Printf("chief version %s\n", Version)
			return nil
		case arg == "--verbose":
			opts.Verbose = true
		case arg == "--merge":
			opts.Merge = true
		case arg == "--force":
			opts.Force = true
		case arg == "--no-retry":
			opts.NoRetry = true
		case arg == "--prompts-dir":
			if i+1 < len(os.Args) {
				i++
				dir := os.Args[i]
				info, err := os.Stat(dir)
				if err != nil || !info.IsDir() {
					fmt.Fprintf(os.Stderr, "prompts directory not found: %s\n", dir)
					os.Exit(1)
				}
				opts.PromptsDir = dir
			} else {
				fmt.Fprintf(os.Stderr, "Error: --prompts-dir requires a value\n")
				os.Exit(1)
			}
		case strings.HasPrefix(arg, "--prompts-dir="):
			dir := strings.TrimPrefix(arg, "--prompts-dir=")
			info, err := os.Stat(dir)
			if err != nil || !info.IsDir() {
				fmt.Fprintf(os.Stderr, "prompts directory not found: %s\n", dir)
				os.Exit(1)
			}
			opts.PromptsDir = dir
		case arg == "--max-iterations" || arg == "-n":
			// Next argument should be the number
			if i+1 < len(os.Args) {
				i++
				n, err := strconv.Atoi(os.Args[i])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: invalid value for %s: %s\n", arg, os.Args[i])
					os.Exit(1)
				}
				if n < 1 {
					fmt.Fprintf(os.Stderr, "Error: --max-iterations must be at least 1\n")
					os.Exit(1)
				}
				opts.MaxIterations = n
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s requires a value\n", arg)
				os.Exit(1)
			}
		case strings.HasPrefix(arg, "--max-iterations="):
			val := strings.TrimPrefix(arg, "--max-iterations=")
			n, err := strconv.Atoi(val)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid value for --max-iterations: %s\n", val)
				os.Exit(1)
			}
			if n < 1 {
				fmt.Fprintf(os.Stderr, "Error: --max-iterations must be at least 1\n")
				os.Exit(1)
			}
			opts.MaxIterations = n
		case strings.HasPrefix(arg, "-n="):
			val := strings.TrimPrefix(arg, "-n=")
			n, err := strconv.Atoi(val)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid value for -n: %s\n", val)
				os.Exit(1)
			}
			if n < 1 {
				fmt.Fprintf(os.Stderr, "Error: -n must be at least 1\n")
				os.Exit(1)
			}
			opts.MaxIterations = n
		case strings.HasPrefix(arg, "-"):
			// Unknown flag
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
			fmt.Fprintf(os.Stderr, "Run 'chief --help' for usage.\n")
			os.Exit(1)
		default:
			// Positional argument: PRD name or path
			if strings.HasSuffix(arg, ".json") || strings.HasSuffix(arg, "/") {
				opts.PRDPath = arg
			} else {
				// Treat as PRD name
				opts.PRDPath = fmt.Sprintf(".chief/prds/%s/prd.json", arg)
			}
		}
	}

	return opts
}

func runNew() {
	opts := cmd.NewOptions{
		PromptsDir: extractGlobalPromptsDir(),
	}

	// Find position of "new" in os.Args, then parse name and context from
	// the arguments that follow it, regardless of leading global flags.
	newIdx := -1
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "new" {
			newIdx = i
			break
		}
	}
	remaining := os.Args[0:0]
	if newIdx >= 0 {
		remaining = os.Args[newIdx+1:]
	}
	if len(remaining) > 0 && !strings.HasPrefix(remaining[0], "-") {
		opts.Name = remaining[0]
	}
	if len(remaining) > 1 {
		opts.Context = strings.Join(remaining[1:], " ")
	}

	if err := cmd.RunNew(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runEdit() {
	opts := cmd.EditOptions{
		PromptsDir: extractGlobalPromptsDir(),
	}

	// Find position of "edit" in os.Args, then parse name and flags from
	// the arguments that follow it, regardless of leading global flags.
	editIdx := -1
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "edit" {
			editIdx = i
			break
		}
	}
	remaining := os.Args[0:0]
	if editIdx >= 0 {
		remaining = os.Args[editIdx+1:]
	}
	for _, arg := range remaining {
		switch arg {
		case "--merge":
			opts.Merge = true
		case "--force":
			opts.Force = true
		default:
			if opts.Name == "" && !strings.HasPrefix(arg, "-") {
				opts.Name = arg
			}
		}
	}

	if err := cmd.RunEdit(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runStatus() {
	opts := cmd.StatusOptions{}

	// Parse arguments: chief status [name]
	if len(os.Args) > 2 && !strings.HasPrefix(os.Args[2], "-") {
		opts.Name = os.Args[2]
	}

	if err := cmd.RunStatus(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runUpdate() {
	if err := cmd.RunUpdate(cmd.UpdateOptions{
		Version: Version,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runList() {
	opts := cmd.ListOptions{}

	if err := cmd.RunList(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runInitPrompts() {
	opts := cmd.InitPromptsOptions{}

	// Find position of "init-prompts" in os.Args, then take the next positional
	// argument (if any) as the target directory path.
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "init-prompts" {
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "-") {
				opts.Path = os.Args[i+1]
			}
			break
		}
	}

	if err := cmd.RunInitPrompts(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runConfig() {
	// Find position of "config" in os.Args
	configIdx := -1
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "config" {
			configIdx = i
			break
		}
	}
	remaining := os.Args[0:0]
	if configIdx >= 0 {
		remaining = os.Args[configIdx+1:]
	}

	// Check for --help
	for _, arg := range remaining {
		if arg == "--help" || arg == "-h" {
			printConfigHelp()
			return
		}
	}

	// Get first positional sub-subcommand (if any)
	subCmd := ""
	for _, arg := range remaining {
		if !strings.HasPrefix(arg, "-") {
			subCmd = arg
			break
		}
	}

	switch subCmd {
	case "":
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		opts := cmd.ConfigOptions{BaseDir: cwd}
		if err := cmd.RunConfig(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "init":
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		force := false
		for _, arg := range remaining {
			if arg == "--force" {
				force = true
			}
		}
		opts := cmd.ConfigInitOptions{BaseDir: cwd, Force: force}
		if err := cmd.RunConfigInit(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown config subcommand: %s\n", subCmd)
		fmt.Fprintf(os.Stderr, "Run 'chief config --help' for usage.\n")
		os.Exit(1)
	}
}

func printConfigHelp() {
	fmt.Println(`chief config - View and manage configuration

Usage:
  chief config             Print current config as YAML
  chief config init        Create a default config file (.chief/config.yaml)
  chief config --help      Show this help message

The config file is stored at .chief/config.yaml in your project directory.
Run 'chief config' to view settings, 'chief config init' to create a config file.`)
}

func runTUIWithOptions(opts *TUIOptions) {
	prdPath := opts.PRDPath

	// If no PRD specified, try to find one
	if prdPath == "" {
		// Try "main" first
		mainPath := ".chief/prds/main/prd.json"
		if _, err := os.Stat(mainPath); err == nil {
			prdPath = mainPath
		} else {
			// Look for any available PRD
			prdPath = findAvailablePRD()
		}

		// If still no PRD found, run first-time setup
		if prdPath == "" {
			cwd, _ := os.Getwd()
			showGitignore := git.IsGitRepo(cwd) && !git.IsChiefIgnored(cwd)

			// Run the first-time setup TUI
			result, err := tui.RunFirstTimeSetup(cwd, showGitignore)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if result.Cancelled {
				return
			}

			// Save config from setup
			cfg := config.Default()
			cfg.OnComplete.Push = result.PushOnComplete
			cfg.OnComplete.CreatePR = result.CreatePROnComplete
			if err := config.Save(cwd, cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save config: %v\n", err)
			}

			// Create the PRD
			newOpts := cmd.NewOptions{
				Name:       result.PRDName,
				PromptsDir: opts.PromptsDir,
			}
			if err := cmd.RunNew(newOpts); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Restart TUI with the new PRD
			opts.PRDPath = fmt.Sprintf(".chief/prds/%s/prd.json", result.PRDName)
			runTUIWithOptions(opts)
			return
		}
	}

	prdDir := filepath.Dir(prdPath)

	// Check if prd.md is newer than prd.json and run conversion if needed
	needsConvert, err := prd.NeedsConversion(prdDir)
	if err != nil {
		fmt.Printf("Warning: failed to check conversion status: %v\n", err)
	} else if needsConvert {
		fmt.Println("prd.md is newer than prd.json, running conversion...")
		convertOpts := prd.ConvertOptions{
			PRDDir:     prdDir,
			Merge:      opts.Merge,
			Force:      opts.Force,
			PromptsDir: opts.PromptsDir,
		}
		if err := prd.Convert(convertOpts); err != nil {
			fmt.Printf("Error converting PRD: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Conversion complete.")
	}

	app, err := tui.NewAppWithOptions(prdPath, opts.MaxIterations)
	if err != nil {
		// Check if this is a missing PRD file error
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			fmt.Printf("PRD not found: %s\n", prdPath)
			fmt.Println()
			// Show available PRDs if any exist
			available := listAvailablePRDs()
			if len(available) > 0 {
				fmt.Println("Available PRDs:")
				for _, name := range available {
					fmt.Printf("  chief %s\n", name)
				}
				fmt.Println()
			}
			fmt.Println("Or create a new one:")
			fmt.Println("  chief new               # Create default PRD")
			fmt.Println("  chief new <name>        # Create named PRD")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		os.Exit(1)
	}

	// Thread PromptsDir into the manager so each iteration can load override prompts
	if opts.PromptsDir != "" {
		app.SetPromptsDir(opts.PromptsDir)
	}

	// Set verbose mode if requested
	if opts.Verbose {
		app.SetVerbose(true)
	}

	// Disable retry if requested
	if opts.NoRetry {
		app.DisableRetry()
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	model, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	// Check for post-exit actions
	if finalApp, ok := model.(tui.App); ok {
		switch finalApp.PostExitAction {
		case tui.PostExitInit:
			// Run new command then restart TUI
			newOpts := cmd.NewOptions{
				Name:       finalApp.PostExitPRD,
				PromptsDir: opts.PromptsDir,
			}
			if err := cmd.RunNew(newOpts); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			// Restart TUI with the new PRD
			opts.PRDPath = fmt.Sprintf(".chief/prds/%s/prd.json", finalApp.PostExitPRD)
			runTUIWithOptions(opts)

		case tui.PostExitEdit:
			// Run edit command then restart TUI
			editOpts := cmd.EditOptions{
				Name:       finalApp.PostExitPRD,
				Merge:      opts.Merge,
				Force:      opts.Force,
				PromptsDir: opts.PromptsDir,
			}
			if err := cmd.RunEdit(editOpts); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			// Restart TUI with the edited PRD
			opts.PRDPath = fmt.Sprintf(".chief/prds/%s/prd.json", finalApp.PostExitPRD)
			runTUIWithOptions(opts)
		}
	}
}

func printHelp() {
	fmt.Println(`Chief - Autonomous PRD Agent

Usage:
  chief [options] [<name>|<path/to/prd.json>]
  chief <command> [arguments]

Commands:
  new [name] [context]      Create a new PRD interactively
  edit [name] [options]     Edit an existing PRD interactively
  status [name]             Show progress for a PRD (default: main)
  list                      List all PRDs with progress
  update                    Update Chief to the latest version
  init-prompts [path]       Scaffold a prompts directory with default templates
  help                      Show this help message

Global Options:
  --prompts-dir <path>      Load prompt overrides from directory (hard-fails if path is invalid)
  --max-iterations N, -n N  Set maximum iterations (default: dynamic)
  --no-retry                Disable auto-retry on Claude crashes
  --verbose                 Show raw Claude output in log
  --merge                   Auto-merge progress on conversion conflicts
  --force                   Auto-overwrite on conversion conflicts
  --help, -h                Show this help message
  --version, -v             Show version number

Edit Options:
  --merge                   Auto-merge progress on conversion conflicts
  --force                   Auto-overwrite on conversion conflicts

Positional Arguments:
  <name>                    PRD name (loads .chief/prds/<name>/prd.json)
  <path/to/prd.json>        Direct path to a prd.json file

Examples:
  chief                     Launch TUI with default PRD (.chief/prds/main/)
  chief auth                Launch TUI with named PRD (.chief/prds/auth/)
  chief ./my-prd.json       Launch TUI with specific PRD file
  chief -n 20               Launch with 20 max iterations
  chief --max-iterations=5 auth
                            Launch auth PRD with 5 max iterations
  chief --verbose           Launch with raw Claude output visible
  chief --prompts-dir ~/chief-prompts
                            Launch using prompt overrides from ~/chief-prompts
  chief new                 Create PRD in .chief/prds/main/
  chief new auth            Create PRD in .chief/prds/auth/
  chief new auth "JWT authentication for REST API"
                            Create PRD with context hint
  chief edit                Edit PRD in .chief/prds/main/
  chief edit auth           Edit PRD in .chief/prds/auth/
  chief edit auth --merge   Edit and auto-merge progress
  chief status              Show progress for default PRD
  chief status auth         Show progress for auth PRD
  chief list                List all PRDs with progress
  chief init-prompts        Scaffold ~/chief-prompts/ with default templates
  chief --version           Show version number`)
}

func printWiggum() {
	// ANSI color codes
	blue := "\033[34m"
	yellow := "\033[33m"
	reset := "\033[0m"

	art := blue + `
                                                                 -=
                                      +%#-   :=#%#**%-
                                     ##+**************#%*-::::=*-
                                   :##***********************+***#
                                 :@#********%#%#******************#*
                                 :##*****%+-:::-%%%%%##************#:
                                   :#%###%%-:::+#*******##%%%*******#%*:
                                      -+%**#%%@@%%%%%%%%%#****#%##*##%%=
                                      -@@%%%%%%%%%%%%%%@*#%%#*##:::
                                    +%%%%%%%%%%%%%%@#+--=#--=#@+:
                                   -@@@@@%@@@@#%#=-=**--+*-----=#:
` + yellow + `                                       :*     *-   - :#-:*=-----=#:
                                       %::%@- *:  *@# +::=*--#=:-%:
                                       #- =+**##-    =*:::#*#-++:*:
                                        #+:-::+--%***-::::::::-*##
                                      :+#:+=:-==-*:::::::::::::::-%
                                     *=::::::::::::::-=*##*:::::::-+
                                     *-::::::::-=+**+-+%%%%+:::::--+
                                      :*%##**==++%%%######%:::::--%-
                                        :-=#--%####%%%%@@+:::::--%=
` + blue + `                     -#%%%%#-` + yellow + `          *:::+%%##%%#%%*:::::::-*#%-
                   :##++++=+++%:` + yellow + `        :@%*:::::::::::::::-=##*%%*%=
                  :%++++@%#+=++#` + yellow + `         %%%=--:::::---=+%%****%##@%#%%*:
                -%=-:-%%%*=+++##` + yellow + `      :*@%***@%%%###*********%%#%********%-
               *#+==**%++++++#*-` + yellow + `   :*%@*+*%*%%%%@*********%%**##****%=--#%*#
             *%#%-:+*++++*%#=#-` + yellow + `  :%#%#*+***#@%%%@%#%%%@%#*****%****%::::::##%-
            :*::::*-%@%@#=*%-` + yellow + `  :%*#%+*******%%%@#*************%****%-::::::**%=
             +==%*+-----+%` + yellow + `    %#*%#********#@%%@********%*%***#%**+*%-:::::*#*%:
              *=::----##**%:` + yellow + `+%#*@**********@%%%%*+***%-::::::#*%#****%#:::-%***%-
               #-:+@#***+*@%` + yellow + `**#%**********%%%#%%*****%::::::-#**%***************%
               =%*****+%%+**` + yellow + `@#%***********@%#%%#******%:::::%****@*********+****##
` + blue + `                %*#%@#*+++**#%` + yellow + `************%%%%%#********###*******@**************%:
                =#**++***+**@` + yellow + `************%%%%#%%*******************%*************##
                 %*++******@#` + yellow + `************@%%#%%@*******************#@*************@:
                  #***+***%#*` + yellow + `************@%%%%%@#*******************#%*************+
                   +#***##%**` + yellow + `************@%%%%%%%********************%************%
                     :######**` + yellow + `*+**********%%%%%%%%*********************%************%
                       :+%@#**` + yellow + `*******+*****#%@@%#******+***************#@*****+*****%:
` + blue + `                         @*********************************************##*+**+*****#+
                        =%%%%%@@@%%#**************************##%%@@@%%%@**********##
                        =%%#%%%%%%%%%%%%%----====%%%%%%%%%%%%%%%%#%%#%%%%%******#%#*%
                        :@@%%#%%%%%%%%%%#::::::::*%%%%%%%%%%%%%%%%%%#%%%@@#%%%##***#%
                          %*##%%@@@@%%%%%::::::::#%%%%%%%@@@@@@%%####****##****#%#==#
                          :%*********************************************#%#*+=-----*-
                           :%************************************+********@:::::----=+
                             ##**********+******************+************##::-::=--#-%
                              =%******************+*+*********************%:=-*:++:#-%
                               *#*****************************************@*#:*:*=:*+=
                                %*********#%#**************************+*%   -#+%**=:
                                **************#%%%%###*******************#
                                =#***************%      #****************#
                                :@***+**********##      *****************#
                                 %**************#=      =#+******+*******#
                                 =#*************%:      :@***************#
                                 :#****+********#        #***************#
                                 :#**************        =#**************#
                                 :%************%-        :%*************##
                                  #***********##          %*************%=
                                -%@@@%######%@@+          =%#***#*#%@@%#@:
                              :%%%%%%%%%%%%%%%%#         +@%%%%%%%%%%%%%%*
                             +@%%%%%%%%%%%%%%%%+       :%%%%%%%%%%%%%%##@+
                             #%%%%%%%%%%%@%@%@*       :@%%%%%%%%%%%%@%%@*
` + reset + `
                         "Bake 'em away, toys!"
                               - Chief Wiggum
`
	fmt.Print(art)
}
