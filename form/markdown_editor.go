package form

import (
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"
)

// MarkdownEditorRender renders a dedicated Markdown editor field.
//
// Unlike RichEditorRender (WYSIWYG), this renders a monospace textarea with a
// markdown-specific toolbar. The textarea carries data-editor="markdown" so that
// a JS markdown library (EasyMDE, SimpleMDE, CodeMirror) can be progressively
// enhanced client-side by targeting [data-editor="markdown"].
//
// Respects RowCount to control the initial height.
func MarkdownEditorRender(f *MarkdownEditorInput) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		name := templ.EscapeString(f.GetName())
		label := templ.EscapeString(f.GetLabel())
		value := templ.EscapeString(f.GetValueString())
		help := templ.EscapeString(f.GetHelp())

		minHeight := fmt.Sprintf("%dpx", f.RowCount*24)

		requiredAttr := ""
		if f.IsRequired() {
			requiredAttr = ` required`
		}
		disabledAttr := ""
		if f.IsDisabled() {
			disabledAttr = ` disabled`
		}

		html := `<div class="space-y-1">`

		// Label
		if label != "" {
			html += `<label for="` + name + `" class="block text-sm font-medium text-gray-700 dark:text-gray-300">` + label
			if f.IsRequired() {
				html += `<span class="text-red-500 ml-1">*</span>`
			}
			html += `</label>`
		}

		// Editor container
		html += `<div class="border border-gray-300 dark:border-gray-600 rounded-xl overflow-hidden">`

		// Toolbar
		html += `<div class="flex items-center gap-1 px-3 py-1.5 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-700">`
		html += `<span class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide mr-auto select-none">MD</span>`

		type toolbarBtn struct{ icon, title string }
		buttons := []toolbarBtn{
			{"format_bold", "Bold (**text**)"},
			{"format_italic", "Italic (*text*)"},
			{"format_quote", "Blockquote"},
			{"format_list_bulleted", "Unordered list"},
			{"format_list_numbered", "Ordered list"},
			{"link", "Link"},
			{"code", "Inline code"},
			{"image", "Image"},
		}
		for _, btn := range buttons {
			html += `<button type="button" title="` + btn.title + `" class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-600 dark:text-gray-400 transition-colors">`
			html += `<span class="material-icons-outlined text-sm">` + btn.icon + `</span>`
			html += `</button>`
		}
		html += `</div>` // end toolbar

		// Textarea (monospace, markdown-aware)
		html += `<textarea`
		html += ` id="` + name + `"`
		html += ` name="` + name + `"`
		html += ` data-editor="markdown"`
		html += requiredAttr + disabledAttr
		html += ` style="min-height:` + minHeight + `"`
		html += ` class="block w-full px-3 py-2 text-sm font-mono bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:outline-none resize-y"`
		html += `>` + value + `</textarea>`

		html += `</div>` // end editor container

		// Help text
		if help != "" {
			html += `<p class="text-xs text-gray-500 dark:text-gray-400">` + help + `</p>`
		}

		html += `</div>`

		_, err := io.WriteString(w, html)
		return err
	})
}
