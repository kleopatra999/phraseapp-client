package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"bytes"

	"github.com/mgutz/ansi"
	"github.com/phrase/phraseapp-go/phraseapp"
)

// PhrasePath replacement for ugly slicing string logic on paths
type PhrasePath struct {
	UserPath    string
	Separator   string
	Components  []string
	GlobPattern string
	IsDir       bool
}

func (p *PhrasePath) RelPath() string {
	return path.Join(p.Separator, p.UserPath, p.Separator)
}

func (p *PhrasePath) AbsPath() (string, error) {
	absBase, err := filepath.Abs("")
	if err != nil {
		return "", err
	}
	return path.Join(absBase, p.Separator, p.RelPath()), nil
}

func (p *PhrasePath) isLocalePatternUsed() bool {
	return p.isLocaleNameTagInPath() || p.isLocaleCodeTagInPath()
}

func (p *PhrasePath) isLocaleNameTagInPath() bool {
	return componentContains(p.Components, "<locale_name>")
}

func (p *PhrasePath) isLocaleCodeTagInPath() bool {
	return componentContains(p.Components, "<locale_code>")
}

func (p *PhrasePath) isFormatTagInPath() bool {
	return componentContains(p.Components, "<format_name>")
}

func (p *PhrasePath) isTagNameInPath() bool {
	return componentContains(p.Components, "<tag_name>")
}

func (p *PhrasePath) isValidLocalePath(locale *phraseapp.Locale) bool {
	return locale != nil && (p.isLocaleCodeTagInPath() && locale.Code != "") || (p.isLocaleNameTagInPath() && locale.Name != "")
}

func PathComponents(userPath string) (*PhrasePath, error) {
	p := &PhrasePath{Separator: string(os.PathSeparator)}

	p.GlobPattern = extractGlobPattern(userPath)
	p.UserPath = cleanUserPath(userPath, p.GlobPattern)
	p.Components = splitToParts(userPath, p.Separator)
	isDir, err := isDir(userPath)
	if err != nil {
		return nil, err
	}

	p.IsDir = isDir

	return p, nil
}

func isDir(path string) (bool, error) {
	if strings.Contains(path, "<") {
		return false, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return false, err
	}
	switch mode := stat.Mode(); {
	case mode.IsDir():
		return true, nil
	}
	return false, nil
}

func extractGlobPattern(userPath string) string {
	if strings.HasSuffix(userPath, path.Join("**", "*")) {
		return "**/*"
	} else if strings.HasSuffix(userPath, "*") {
		return "*"
	} else {
		return ""
	}
}

func cleanUserPath(userPath, mode string) string {
	pathWithoutMode := trimSuffix(userPath, mode)
	return strings.TrimSpace(pathWithoutMode)
}

func splitToParts(userPath, separator string) []string {
	split := strings.Split(userPath, separator)
	components := []string{}
	for _, part := range split {
		if part != separator {
			components = append(components, part)
		}
	}
	return components
}

func componentContains(components []string, pattern string) bool {
	for _, part := range components {
		if strings.Contains(part, pattern) {
			return true
		}
	}
	return false
}

// Locale to Path mapping
func CopyLocalePath(relPath string, l *LocalePath) *LocalePath {
	return &LocalePath{Path: relPath, LocaleId: l.LocaleId, LocaleName: l.LocaleName, LocaleCode: l.LocaleCode}
}

func InitLocalePathWithLocale(relPath string, l *phraseapp.Locale) *LocalePath {
	return &LocalePath{Path: relPath, LocaleId: l.Id, LocaleName: l.Name, LocaleCode: l.Code}
}

func InitLocalePathWithLocaleId(relPath string, localeId string) *LocalePath {
	return &LocalePath{Path: relPath, LocaleId: localeId}
}

type LocalePaths []*LocalePath
type LocalePath struct {
	Path       string
	LocaleId   string
	LocaleName string
	LocaleCode string
}

// Locale placeholder logic <locale_name>
func ExpandPathsWithLocale(p *PhrasePath, localeId string, locales []*phraseapp.Locale) (LocalePaths, error) {
	switch {
	case p.isLocalePatternUsed() && localeId != "":
		return singlePathWithLocale(p, localeId, locales)

	case p.isLocalePatternUsed() && localeId == "":
		return multiplePathsWithLocales(p, locales)

	case !p.isLocalePatternUsed():
		return singlePathWithoutLocale(p, localeId)

	default:
		if localeId == "" {
			return nil, fmt.Errorf("no target locale id specified")
		}
		return defaultPathWithLocale(p, localeId, locales)
	}
}

func singlePathWithLocale(p *PhrasePath, localeId string, locales []*phraseapp.Locale) (LocalePaths, error) {
	newPaths := []*LocalePath{}
	locale := localeForLocaleId(localeId, locales)
	if !p.isValidLocalePath(locale) {
		return nil, fmt.Errorf("Could not find remote locale with Id:", localeId)
	}
	absPath, err := newLocaleFile(p, locale.Name, locale.Code)
	if err != nil {
		return nil, err
	}
	newPaths = append(newPaths, InitLocalePathWithLocale(absPath, locale))
	return newPaths, nil
}

func multiplePathsWithLocales(p *PhrasePath, locales []*phraseapp.Locale) (LocalePaths, error) {
	files := []*LocalePath{}
	for _, locale := range locales {
		if !p.isValidLocalePath(locale) {
			continue
		}
		absPath, err := newLocaleFile(p, locale.Name, locale.Code)
		if err != nil {
			return nil, err
		}
		localePath := InitLocalePathWithLocale(absPath, locale)
		files = append(files, localePath)
	}
	return files, nil
}

func singlePathWithoutLocale(p *PhrasePath, localeId string) (LocalePaths, error) {
	absPath, err := filepath.Abs(p.UserPath)
	if err != nil {
		return nil, err
	}

	localePath := []*LocalePath{InitLocalePathWithLocaleId(absPath, localeId)}
	return localePath, nil
}

func defaultPathWithLocale(p *PhrasePath, localeId string, locales []*phraseapp.Locale) (LocalePaths, error) {
	absPath, err := filepath.Abs(p.UserPath)
	if err != nil {
		return nil, err
	}

	matchedLocale := localeIdForPath(localeId, locales)
	if matchedLocale == "" {
		return nil, fmt.Errorf("locale specified in your path did not match any remote locales")
	}

	localePath := []*LocalePath{InitLocalePathWithLocaleId(absPath, matchedLocale)}
	return localePath, nil
}

// Locale logic
func localeForLocaleId(localeId string, locales []*phraseapp.Locale) *phraseapp.Locale {
	for _, locale := range locales {
		if locale.Id == localeId {
			return locale
		}
	}
	return nil
}

func localeIdForPath(localeId string, locales []*phraseapp.Locale) string {
	for _, locale := range locales {
		if localeId == locale.Id {
			return locale.Id
		}
	}
	return ""
}

func newLocaleFile(p *PhrasePath, localeName, localeCode string) (string, error) {
	absPath, err := p.AbsPath()
	if err != nil {
		return "", err
	}

	realPath := strings.Replace(absPath, "<locale_name>", localeName, -1)
	realPath = strings.Replace(realPath, "<locale_code>", localeCode, -1)

	return realPath, nil
}

func isLocaleFile(file, extension string) bool {
	fileExtension := fmt.Sprintf(".%s", extension)
	return strings.HasSuffix(file, fileExtension)
}

func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

// File creation
func createFile(realPath string) error {
	err := fileExists(realPath)
	if err != nil {
		absDir := filepath.Dir(realPath)
		err := fileExists(absDir)
		if err != nil {
			os.MkdirAll(absDir, 0700)
		}

		f, err := os.Create(realPath)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	return nil
}

func fileExists(absPath string) error {
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("no such file or directory:", absPath)
	}
	return nil
}

func Authenticate() error {
	defaultCredentials, err := ConfigDefaultCredentials()
	if err != nil {
		return err
	}
	phraseapp.RegisterAuthCredentials(defaultCredentials, defaultCredentials)
	return nil
}

func printErr(err error, msg string) {
	red := ansi.ColorCode("red+b:black")
	reset := ansi.ColorCode("reset")
	fmt.Fprintf(os.Stderr, "%sERROR: %s %s%s\n", red, err, msg, reset)
}

func sharedMessage(method string, localePath *LocalePath) {
	yellow := ansi.ColorCode("yellow+b:black")
	reset := ansi.ColorCode("reset")

	localPath := localePath.Path
	callerPath, err := os.Getwd()
	if err == nil {
		localPath = "." + strings.Replace(localePath.Path, callerPath, "", 1)
	}
	localeName := localePath.LocaleName
	localeCode := localePath.LocaleCode

	fmt1 := []string{}

	fmt1 = append(fmt1, yellow)
	if localeName != "" {
		fmt1 = append(fmt1, localeName)
	}
	if localeCode != "" {
		fmt1 = append(fmt1, fmt.Sprintf(" (%s)", localeCode))
	}
	fmt1 = append(fmt1, reset)

	if len(fmt1) <= 2 {
		fmt1 = []string{yellow, "?", reset}
	}

	fmt2 := []string{}
	fmt2 = append(fmt2, yellow)
	fmt2 = append(fmt2, localPath)
	fmt2 = append(fmt2, reset)

	remote := strings.Join(fmt1, "")
	local := strings.Join(fmt2, "")

	from, to := "", ""
	if method == "pull" {
		from, to = remote, local
	} else {
		from, to = local, remote
	}

	var buffer bytes.Buffer
	buffer.WriteString("From: ")
	buffer.WriteString(from)
	buffer.WriteString(" To: ")
	buffer.WriteString(to)
	fmt.Println(buffer.String())
}

func wasAlreadySeen(alreadySeen []string, maybeSeen string) bool {
	for _, seen := range alreadySeen {
		if maybeSeen == seen {
			return true
		}
	}
	return false
}
