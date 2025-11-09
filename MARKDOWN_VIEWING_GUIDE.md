# Markdown Viewing Guide

## Quick Setup for Better Markdown Viewing

### 1. Install Recommended Extensions

Open Command Palette (`Cmd+Shift+P` on Mac, `Ctrl+Shift+P` on Windows/Linux) and type:
```
Extensions: Show Recommended Extensions
```

Or install manually:
- **Markdown All in One** (`yzhang.markdown-all-in-one`)
  - Keyboard shortcuts, table formatting, auto-completion
- **Markdown Preview Enhanced** (`shd101wyy.markdown-preview-enhanced`)
  - Beautiful preview with themes, diagrams, math support
- **Markdown Preview Mermaid Support** (`bierner.markdown-mermaid`)
  - For rendering Mermaid diagrams
- **markdownlint** (`davidanson.vscode-markdownlint`)
  - Linting and formatting help

### 2. View Markdown Preview

**Keyboard Shortcuts:**
- **Mac:** `Cmd+Shift+V` - Open preview
- **Windows/Linux:** `Ctrl+Shift+V` - Open preview
- **Side-by-side:** `Cmd+K V` (Mac) or `Ctrl+K V` (Windows/Linux)

**Or use Command Palette:**
1. Press `Cmd+Shift+P` (Mac) or `Ctrl+Shift+P` (Windows/Linux)
2. Type "Markdown: Open Preview"
3. Select the command

### 3. Recommended Settings

The `.vscode/settings.json` file has been created with optimal settings:
- ✅ Word wrap enabled
- ✅ Line numbers enabled
- ✅ Better font size and line height
- ✅ Auto-formatting on save
- ✅ Preview scrolls with editor

### 4. Tips for Better Viewing

#### View Side-by-Side
1. Open your markdown file
2. Press `Cmd+K V` (Mac) or `Ctrl+K V` (Windows/Linux)
3. Editor and preview will be side-by-side

#### Use Zen Mode for Reading
- Press `Cmd+K Z` (Mac) or `Ctrl+K Z` (Windows/Linux)
- Full-screen distraction-free reading

#### Navigate Headings
- Press `Cmd+Shift+O` (Mac) or `Ctrl+Shift+O` (Windows/Linux)
- See outline of all headings
- Click to jump to section

#### Toggle Word Wrap
- Press `Alt+Z` to toggle word wrap on/off

### 5. Markdown Preview Enhanced Features

If you install **Markdown Preview Enhanced**, you get:
- Multiple themes (GitHub, Vue, OneDark, etc.)
- Export to PDF, HTML, PNG
- Math equation support
- Mermaid diagram support
- Code block execution
- Presentation mode

**To use:**
1. Right-click in markdown file
2. Select "Markdown Preview Enhanced: Open Preview to the Side"
3. Or use `Cmd+Shift+M` (Mac) / `Ctrl+Shift+M` (Windows/Linux)

### 6. Custom CSS (Optional)

Create `.vscode/markdown.css` for custom styling:

```css
/* Custom Markdown Preview Styles */
.markdown-preview {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  line-height: 1.8;
}

.markdown-preview h1 {
  border-bottom: 3px solid #667eea;
  padding-bottom: 10px;
}

.markdown-preview h2 {
  border-bottom: 2px solid #e0e0e0;
  padding-bottom: 8px;
}

.markdown-preview code {
  background: #f4f4f4;
  padding: 2px 6px;
  border-radius: 3px;
}

.markdown-preview pre {
  background: #2d2d2d;
  color: #f8f8f2;
  padding: 15px;
  border-radius: 5px;
  overflow-x: auto;
}
```

Then add to `settings.json`:
```json
"markdown.styles": [".vscode/markdown.css"]
```

### 7. Keyboard Shortcuts Reference

| Action | Mac | Windows/Linux |
|--------|-----|---------------|
| Open Preview | `Cmd+Shift+V` | `Ctrl+Shift+V` |
| Open Preview Side | `Cmd+K V` | `Ctrl+K V` |
| Toggle Word Wrap | `Alt+Z` | `Alt+Z` |
| Show Outline | `Cmd+Shift+O` | `Ctrl+Shift+O` |
| Zen Mode | `Cmd+K Z` | `Ctrl+K Z` |
| Format Document | `Shift+Alt+F` | `Shift+Alt+F` |

### 8. Viewing Large Documents

For large documents like `PRODUCTION_READINESS_ANALYSIS.md`:

1. **Use Outline View:**
   - Press `Cmd+Shift+O` to see all headings
   - Click any heading to jump to that section

2. **Use Breadcrumbs:**
   - Enable breadcrumbs in View menu
   - See your current location in the document

3. **Use Find:**
   - Press `Cmd+F` to search within document
   - Press `Cmd+Shift+F` to search across all files

4. **Fold Sections:**
   - Click the fold icon next to headings
   - Or use `Cmd+K Cmd+0` to fold all
   - Use `Cmd+K Cmd+J` to unfold all

### 9. Print-Friendly View

To print or export:
1. Open preview (`Cmd+Shift+V`)
2. Right-click in preview
3. Select "Print" or use browser print (`Cmd+P`)
4. Or use Markdown Preview Enhanced to export as PDF

### 10. Troubleshooting

**Preview not showing:**
- Make sure file has `.md` extension
- Check if markdown extension is installed
- Try reloading window (`Cmd+R` or `Ctrl+R`)

**Preview looks different:**
- Check settings in `.vscode/settings.json`
- Try different preview themes
- Install Markdown Preview Enhanced for more options

**Text too small/large:**
- Adjust `markdown.preview.fontSize` in settings
- Or use `Cmd++` / `Cmd+-` to zoom in preview

---

## Quick Start Checklist

- [ ] Install "Markdown All in One" extension
- [ ] Install "Markdown Preview Enhanced" extension (optional but recommended)
- [ ] Open `PRODUCTION_READINESS_ANALYSIS.md`
- [ ] Press `Cmd+Shift+V` (or `Ctrl+Shift+V`) to open preview
- [ ] Press `Cmd+K V` for side-by-side view
- [ ] Press `Cmd+Shift+O` to see document outline
- [ ] Enjoy reading! 📖

---

**Pro Tip:** For the best experience with large technical documents, use:
1. **Side-by-side view** (`Cmd+K V`)
2. **Outline panel** (`Cmd+Shift+O`) 
3. **Zen mode** (`Cmd+K Z`) for focused reading

Happy reading! 🎉

