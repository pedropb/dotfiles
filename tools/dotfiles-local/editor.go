package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// Row keys for the editor's main menu. Identity, credential-helper, and
// provider rows carry their map key after one of the prefixes below, e.g.
// "identity:work".
const (
	rowKeySystem          = "machine.system"
	rowKeyUsername        = "machine.username"
	rowKeyHomeDirectory   = "machine.home_directory"
	rowKeyAddIdentity     = "identity.add"
	rowKeyAddHelper       = "helper.add"
	rowKeyAddProvider     = "provider.add"
	rowKeyDefaultProvider = "cln.default_provider"
	rowKeySave            = "action.save"
	rowKeyQuit            = "action.quit"
	rowPrefixIdentity     = "identity:"
	rowPrefixHelper       = "helper:"
	rowPrefixProvider     = "provider:"
)

// menuRow is one line of the editor's section list: a stable key the switch
// in runEditor dispatches on, and the label shown to the user.
type menuRow struct {
	key   string
	label string
}

// buildRows lists every editable fact in cfg plus the "add" and "action"
// rows, in the fixed order they're shown. It has no huh dependency so it can
// be tested directly.
func buildRows(cfg Config) []menuRow {
	rows := []menuRow{
		{rowKeySystem, "Machine: system            " + valueOr(cfg.Machine.System)},
		{rowKeyUsername, "Machine: username          " + valueOr(cfg.Machine.Username)},
		{rowKeyHomeDirectory, "Machine: home directory    " + valueOr(cfg.Machine.HomeDirectory)},
	}
	for _, name := range sortedKeys(cfg.identities()) {
		id := cfg.identities()[name]
		rows = append(rows, menuRow{rowPrefixIdentity + name,
			fmt.Sprintf("Identity: %-14s %s <%s>", name, id.Name, id.Email)})
	}
	rows = append(rows, menuRow{rowKeyAddIdentity, "+ Add git identity"})

	for _, host := range sortedKeys(cfg.credentialHelpers()) {
		rows = append(rows, menuRow{rowPrefixHelper + host,
			fmt.Sprintf("Credential helper: %-20s %s", host, cfg.credentialHelpers()[host])})
	}
	rows = append(rows, menuRow{rowKeyAddHelper, "+ Add credential helper"})

	for _, alias := range sortedKeys(cfg.providers()) {
		p := cfg.providers()[alias]
		rows = append(rows, menuRow{rowPrefixProvider + alias,
			fmt.Sprintf("cln provider: %-14s %s at %s", alias, p.Type, p.Host)})
	}
	rows = append(rows, menuRow{rowKeyAddProvider, "+ Add cln provider"})

	def := cfg.defaultProvider()
	if def == "" {
		def = builtinProvider
	}
	rows = append(rows, menuRow{rowKeyDefaultProvider, "Default cln provider:      " + def})
	rows = append(rows, menuRow{rowKeySave, "Save"})
	rows = append(rows, menuRow{rowKeyQuit, "Quit"})
	return rows
}

func valueOr(s string) string {
	if s == "" {
		return "(unset)"
	}
	return s
}

// configsEqual compares configs by their canonical encoding rather than
// reflect.DeepEqual, since encodeConfig is already guaranteed deterministic
// and sorted — the same property that makes a save idempotent.
func configsEqual(a, b Config) bool {
	da, errA := encodeConfig(a)
	db, errB := encodeConfig(b)
	return errA == nil && errB == nil && bytes.Equal(da, db)
}

// parseGitdirs splits the comma-separated directory-prefixes input into the
// trimmed, non-empty list Identity.Gitdir expects.
func parseGitdirs(raw string) []string {
	var dirs []string
	for _, part := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			dirs = append(dirs, trimmed)
		}
	}
	return dirs
}

// keyCollision reports whether writing newKey would silently clobber a
// different existing entry — true both for "+ Add" typing a name already in
// use, and for a rename that lands on another entry. Editing an entry
// without renaming it (newKey == existingKey) is not a collision.
func keyCollision(existingKey, newKey string, newKeyExists bool) bool {
	return newKeyExists && newKey != existingKey
}

// newForm builds a huh.Form themed and wired to run full-screen against
// env's streams. Every prompt in the editor goes through this, so the
// experience — theme, alt-screen, help line — is identical everywhere.
func newForm(env *env, groups ...*huh.Group) *huh.Form {
	return huh.NewForm(groups...).
		WithTheme(huh.ThemeCharm()).
		WithProgramOptions(tea.WithAltScreen()).
		WithInput(env.in).
		WithOutput(env.out).
		WithShowHelp(true)
}

// runForm runs form and reports whether the user backed out (Esc/Ctrl+C)
// instead of completing it, which every caller treats as "make no change".
func runForm(form *huh.Form) (aborted bool, err error) {
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

// confirmDialog is a themed yes/no prompt. Aborting counts as "no".
func confirmDialog(env *env, question string, fallback bool) (bool, error) {
	value := fallback
	form := newForm(env, huh.NewGroup(
		huh.NewConfirm().Title(question).Affirmative("Yes").Negative("No").Value(&value),
	))
	aborted, err := runForm(form)
	if aborted || err != nil {
		return false, err
	}
	return value, nil
}

// noteDialog shows a message the user dismisses with Enter — used for
// validation errors and other results that need to be seen before returning
// to the section list.
func noteDialog(env *env, title, description string) error {
	form := newForm(env, huh.NewGroup(
		huh.NewNote().Title(title).Description(description).Next(true).NextLabel("Back"),
	))
	_, err := runForm(form)
	return err
}

// runEditor is the full-screen settings editor shared by `init` and `edit`:
// a live preview of the whole file plus a picker over every section, mirror-
// ing the section list + detail layout of the omp settings TUI. It returns
// once the user saves or quits.
func runEditor(env *env, cfg Config, path string, existed bool) error {
	t := newTheme(env.out)
	original := cfg

	for {
		rows := buildRows(cfg)
		options := make([]huh.Option[string], len(rows))
		for i, row := range rows {
			options[i] = huh.NewOption(row.label, row.key)
		}
		var choice string
		form := newForm(env, huh.NewGroup(
			huh.NewNote().Title(path).Description(t.renderPreview(cfg)),
			huh.NewSelect[string]().
				Title("Select a section to edit").
				Options(options...).
				Height(len(options)+1).
				Value(&choice),
		))
		if aborted, err := runForm(form); err != nil {
			return err
		} else if aborted {
			choice = rowKeyQuit
		}

		var err error
		switch {
		case choice == rowKeySystem:
			err = editMachineField(env, &cfg, "system")
		case choice == rowKeyUsername:
			err = editMachineField(env, &cfg, "username")
		case choice == rowKeyHomeDirectory:
			err = editMachineField(env, &cfg, "home")
		case choice == rowKeyAddIdentity:
			err = editIdentityForm(env, &cfg, "")
		case strings.HasPrefix(choice, rowPrefixIdentity):
			name := strings.TrimPrefix(choice, rowPrefixIdentity)
			id := cfg.identities()[name]
			err = manageEntry(env, fmt.Sprintf("Identity %s: %s <%s>", name, id.Name, id.Email),
				func() error { return editIdentityForm(env, &cfg, name) },
				func() error { return cfg.removeIdentity(name) },
			)
		case choice == rowKeyAddHelper:
			err = editHelperForm(env, &cfg, "")
		case strings.HasPrefix(choice, rowPrefixHelper):
			host := strings.TrimPrefix(choice, rowPrefixHelper)
			err = manageEntry(env, fmt.Sprintf("Credential helper %s: %s", host, cfg.credentialHelpers()[host]),
				func() error { return editHelperForm(env, &cfg, host) },
				func() error { return cfg.removeCredentialHelper(host) },
			)
		case choice == rowKeyAddProvider:
			err = editProviderForm(env, &cfg, "")
		case strings.HasPrefix(choice, rowPrefixProvider):
			alias := strings.TrimPrefix(choice, rowPrefixProvider)
			p := cfg.providers()[alias]
			err = manageEntry(env, fmt.Sprintf("cln provider %s: %s at %s", alias, p.Type, p.Host),
				func() error { return editProviderForm(env, &cfg, alias) },
				func() error { _, err := cfg.removeProvider(alias); return err },
			)
		case choice == rowKeyDefaultProvider:
			err = editDefaultProviderForm(env, &cfg)
		case choice == rowKeySave:
			var saved bool
			saved, err = saveWithConfirm(env, t, cfg, path, existed)
			if saved {
				return nil
			}
		case choice == rowKeyQuit:
			var quit bool
			quit, err = confirmQuit(env, !configsEqual(cfg, original))
			if quit {
				fmt.Fprintln(env.out, t.muted.Render("nothing written"))
				return nil
			}
		}
		if err != nil {
			return err
		}
	}
}

// manageEntry is the Edit/Remove/Back sub-menu shown for an existing
// identity, credential helper, or provider row.
func manageEntry(env *env, title string, edit func() error, remove func() error) error {
	var action string
	form := newForm(env, huh.NewGroup(
		huh.NewSelect[string]().
			Title(title).
			Options(
				huh.NewOption("Edit", "edit"),
				huh.NewOption("Remove", "remove"),
				huh.NewOption("Back", "back"),
			).
			Value(&action),
	))
	if aborted, err := runForm(form); aborted || err != nil {
		return err
	}
	switch action {
	case "edit":
		return edit()
	case "remove":
		confirmed, err := confirmDialog(env, "Remove "+title+"?", false)
		if err != nil || !confirmed {
			return err
		}
		return remove()
	}
	return nil
}

func editMachineField(env *env, cfg *Config, field string) error {
	switch field {
	case "system":
		choice := cfg.Machine.System
		opts := make([]huh.Option[string], len(knownSystems))
		for i, s := range knownSystems {
			opts[i] = huh.NewOption(s, s)
		}
		form := newForm(env, huh.NewGroup(
			huh.NewSelect[string]().Title("Machine system").Options(opts...).Value(&choice),
		))
		if aborted, err := runForm(form); aborted || err != nil {
			return err
		}
		cfg.Machine.System = choice
	case "username":
		value := cfg.Machine.Username
		form := newForm(env, huh.NewGroup(
			huh.NewInput().Title("Username").Value(&value).Validate(huh.ValidateNotEmpty()),
		))
		if aborted, err := runForm(form); aborted || err != nil {
			return err
		}
		cfg.Machine.Username = value
	case "home":
		value := cfg.Machine.HomeDirectory
		form := newForm(env, huh.NewGroup(
			huh.NewInput().Title("Home directory").Value(&value).Validate(huh.ValidateNotEmpty()),
		))
		if aborted, err := runForm(form); aborted || err != nil {
			return err
		}
		cfg.Machine.HomeDirectory = value
	}
	return nil
}

func editIdentityForm(env *env, cfg *Config, existingName string) error {
	id := cfg.identities()[existingName]
	name := existingName
	author := id.Name
	email := id.Email
	gitdirs := strings.Join(id.Gitdir, ", ")
	title := "Add git identity"
	if existingName != "" {
		title = "Edit git identity " + existingName
	}

	form := newForm(env, huh.NewGroup(
		huh.NewInput().Title("Short name").Description("e.g. work").Value(&name).Validate(huh.ValidateNotEmpty()),
		huh.NewInput().Title("Author name").Value(&author).Validate(huh.ValidateNotEmpty()),
		huh.NewInput().Title("Author email").Value(&email).Validate(huh.ValidateNotEmpty()),
		huh.NewInput().Title("Directory prefixes").Description("comma separated, e.g. ~/src/git.example.com/").
			Value(&gitdirs).Validate(huh.ValidateNotEmpty()),
	).Title(title))
	if aborted, err := runForm(form); aborted || err != nil {
		return err
	}

	dirs := parseGitdirs(gitdirs)
	if _, exists := cfg.identities()[name]; keyCollision(existingName, name, exists) {
		return noteDialog(env, "Cannot save identity",
			fmt.Sprintf("%q already exists. Choose a different short name, or edit %s directly instead.", name, name))
	}
	if existingName != "" && existingName != name {
		if err := cfg.removeIdentity(existingName); err != nil {
			return err
		}
	}
	cfg.setIdentity(name, Identity{Gitdir: dirs, Name: author, Email: email})
	return nil
}

func editHelperForm(env *env, cfg *Config, existingHost string) error {
	host := existingHost
	command := cfg.credentialHelpers()[existingHost]
	if existingHost == "" {
		command = "!glab auth git-credential"
	}
	title := "Add credential helper"
	if existingHost != "" {
		title = "Edit credential helper " + existingHost
	}

	form := newForm(env, huh.NewGroup(
		huh.NewInput().Title("Host").Description("exactly as the remote URL spells it").
			Value(&host).Validate(huh.ValidateNotEmpty()),
		huh.NewInput().Title("Helper command").Value(&command).Validate(huh.ValidateNotEmpty()),
	).Title(title))
	if aborted, err := runForm(form); aborted || err != nil {
		return err
	}

	if _, exists := cfg.credentialHelpers()[host]; keyCollision(existingHost, host, exists) {
		return noteDialog(env, "Cannot save credential helper",
			fmt.Sprintf("%q already has a helper configured. Choose a different host, or edit %s directly instead.", host, host))
	}
	if existingHost != "" && existingHost != host {
		if err := cfg.removeCredentialHelper(existingHost); err != nil {
			return err
		}
	}
	cfg.setCredentialHelper(host, command)
	return nil
}

func editProviderForm(env *env, cfg *Config, existingAlias string) error {
	p := cfg.providers()[existingAlias]
	alias := existingAlias
	kind := p.Type
	if kind == "" {
		kind = "gitlab"
	}
	host := p.Host
	namespace := p.DefaultNamespace
	title := "Add cln provider"
	if existingAlias != "" {
		title = "Edit cln provider " + existingAlias
	}

	form := newForm(env, huh.NewGroup(
		huh.NewInput().Title("Alias").Description("e.g. corp").Value(&alias).Validate(huh.ValidateNotEmpty()),
		huh.NewSelect[string]().Title("Type").Options(
			huh.NewOption("gitlab", "gitlab"),
			huh.NewOption("github", "github"),
		).Value(&kind),
		huh.NewInput().Title("Host").Value(&host).Validate(huh.ValidateNotEmpty()),
		huh.NewInput().Title("Default namespace").Description("optional").Value(&namespace),
	).Title(title))
	if aborted, err := runForm(form); aborted || err != nil {
		return err
	}

	if _, exists := cfg.providers()[alias]; keyCollision(existingAlias, alias, exists) {
		return noteDialog(env, "Cannot save cln provider",
			fmt.Sprintf("%q already exists. Choose a different alias, or edit %s directly instead.", alias, alias))
	}
	if existingAlias != "" && existingAlias != alias {
		if _, err := cfg.removeProvider(existingAlias); err != nil {
			return err
		}
	}
	cfg.setProvider(alias, Provider{Type: kind, Host: host, DefaultNamespace: namespace})
	return nil
}

func editDefaultProviderForm(env *env, cfg *Config) error {
	if len(cfg.providers()) == 0 {
		return noteDialog(env, "Default cln provider",
			"No providers configured yet; gh (github.com) is the only option.")
	}
	options := append([]string{builtinProvider}, sortedKeys(cfg.providers())...)
	choice := cfg.defaultProvider()
	if choice == "" {
		choice = builtinProvider
	}
	opts := make([]huh.Option[string], len(options))
	for i, o := range options {
		opts[i] = huh.NewOption(o, o)
	}
	form := newForm(env, huh.NewGroup(
		huh.NewSelect[string]().Title("Default cln provider").Options(opts...).Value(&choice),
	))
	if aborted, err := runForm(form); aborted || err != nil {
		return err
	}
	cfg.setDefaultProvider(choice)
	return nil
}

// saveWithConfirm validates cfg, confirms overwriting an existing file
// (saveConfig backs it up to path+".bak" first), and writes it. It reports
// saved=false — without error — whenever the caller should stay in the
// editor: an invalid config, or a declined confirmation.
func saveWithConfirm(env *env, t *theme, cfg Config, path string, existed bool) (saved bool, err error) {
	if err := cfg.validate(); err != nil {
		if dialogErr := noteDialog(env, "Cannot save", t.fail.Render(err.Error())); dialogErr != nil {
			return false, dialogErr
		}
		return false, nil
	}

	question := fmt.Sprintf("Write %s?", path)
	if existed {
		question = fmt.Sprintf("Overwrite %s?\nThe existing file is kept as %s.bak.", path, path)
	}
	ok, err := confirmDialog(env, question, true)
	if err != nil || !ok {
		return false, err
	}
	if err := saveConfig(cfg); err != nil {
		return false, err
	}
	fmt.Fprintln(env.out, t.ok.Render("✓ wrote "+path))
	fmt.Fprintln(env.out, t.muted.Render("run `just switch` to apply"))
	return true, nil
}

func confirmQuit(env *env, dirty bool) (bool, error) {
	if !dirty {
		return true, nil
	}
	return confirmDialog(env, "Discard unsaved changes?", false)
}
