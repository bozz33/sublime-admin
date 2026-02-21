package commands

import (
"bufio"
"fmt"
"os"
"strings"

"github.com/bozz33/sublimeadmin/generator"
"github.com/spf13/cobra"
)

// make:widget
var makeWidgetCmd = &cobra.Command{
Use:     "widget [name]",
Aliases: []string{"w"},
Short:   "Generate a new dashboard widget",
Long: `Generate a widget file with a data provider.

Example: sublimego make:widget RevenueWidget`,
Args: cobra.MaximumNArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
name, err := promptName(args, "Widget name (ex: Revenue): ")
if err != nil {
return err
}
g, err := newGenerator()
if err != nil {
return err
}
return generator.GenerateWidget(g, name, ".")
},
}

// make:action
var makeActionCmd = &cobra.Command{
Use:     "action [name]",
Aliases: []string{"a"},
Short:   "Generate a new resource action",
Long: `Generate an action file with handler, confirmation, and notification.

Example: sublimego make:action PublishPost`,
Args: cobra.MaximumNArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
name, err := promptName(args, "Action name (ex: Publish): ")
if err != nil {
return err
}
g, err := newGenerator()
if err != nil {
return err
}
return generator.GenerateAction(g, name, ".")
},
}

// make:enum
var makeEnumCmd = &cobra.Command{
Use:     "enum [name]",
Aliases: []string{"e"},
Short:   "Generate a new typed enum",
Long: `Generate an enum with HasLabel, HasColor, HasIcon interfaces.

Example: sublimego make:enum Status`,
Args: cobra.MaximumNArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
name, err := promptName(args, "Enum name (ex: Status): ")
if err != nil {
return err
}
g, err := newGenerator()
if err != nil {
return err
}
return generator.GenerateEnum(g, name, ".")
},
}

// make:infolist
var makeInfolistCmd = &cobra.Command{
Use:     "infolist [name]",
Aliases: []string{"il"},
Short:   "Generate an infolist view for a resource",
Long: `Generate a read-only detail view (infolist) for an existing resource.

Example: sublimego make:infolist Product`,
Args: cobra.MaximumNArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
name, err := promptName(args, "Resource name (ex: Product): ")
if err != nil {
return err
}
g, err := newGenerator()
if err != nil {
return err
}
return generator.GenerateInfolist(g, name, ".")
},
}

// promptName reads a name from args or stdin interactively.
func promptName(args []string, prompt string) (string, error) {
if len(args) > 0 {
return strings.TrimSpace(args[0]), nil
}
fmt.Print(prompt)
reader := bufio.NewReader(os.Stdin)
input, _ := reader.ReadString('\n')
name := strings.TrimSpace(input)
if name == "" {
return "", fmt.Errorf("name is required")
}
return name, nil
}

// newGenerator creates a generator with current flags.
func newGenerator() (*generator.Generator, error) {
return generator.New(&generator.Options{
Force:   forceFlag,
DryRun:  dryRunFlag,
Skip:    skipFlag,
NoBackup: noBackupFlag,
Verbose: verboseFlag,
})
}

func init() {
makeCmd.AddCommand(makeWidgetCmd)
makeCmd.AddCommand(makeActionCmd)
makeCmd.AddCommand(makeEnumCmd)
makeCmd.AddCommand(makeInfolistCmd)
}
