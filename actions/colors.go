package actions

// Color is a strongly-typed semantic color name used for action and button styling.
// Using the Color type instead of raw strings prevents typos and enables IDE autocompletion.
//
// Example:
//
//	actions.New("archive").SetColor(actions.ColorWarning)
type Color = string

// Predefined semantic color constants.
// All panel components (actions, badges, buttons) accept these values.
const (
	ColorPrimary   Color = "primary"
	ColorSecondary Color = "secondary"
	ColorDanger    Color = "danger"
	ColorWarning   Color = "warning"
	ColorSuccess   Color = "success"
	ColorInfo      Color = "info"
	ColorGray      Color = "gray"
)
