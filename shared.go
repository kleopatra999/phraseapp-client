package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/phrase/phraseapp-go/phraseapp"
)

var (
	Debug             bool
	separator         = string(os.PathSeparator)
	placeholderRegexp = regexp.MustCompile("<(locale_name|locale_code|tag)>")
	localePlaceholder = regexp.MustCompile("<(locale_name|locale_code)>")
	tagPlaceholder    = regexp.MustCompile("<(tag)>")
)

const (
	docsBaseUrl   = "https://phraseapp.com/docs"
	docsConfigUrl = docsBaseUrl + "/developers/cli/configuration"
)

func containsAnyPlaceholders(s string) bool {
	return (localePlaceholder.MatchString(s) || tagPlaceholder.MatchString(s))
}

func containsLocalePlaceholder(s string) bool {
	return localePlaceholder.MatchString(s)
}

func containsTagPlaceholder(s string) bool {
	return tagPlaceholder.MatchString(s)
}

func docsFormatsUrl(formatName string) string {
	return fmt.Sprintf("%s/guides/formats/%s", docsBaseUrl, formatName)
}

func ValidPath(file, formatName, formatExtension string) error {
	if strings.TrimSpace(file) == "" {
		return fmt.Errorf("File patterns may not be empty!\nFor more information see %s", docsConfigUrl)
	}

	fileExtension := strings.Trim(filepath.Ext(file), ".")

	if fileExtension == "<locale_code>" {
		return nil
	}

	if fileExtension == "" {
		return fmt.Errorf("%q has no file extension", file)
	}

	if formatExtension != "" && formatExtension != fileExtension {
		return fmt.Errorf(
			"File extension %q does not equal %q (format: %q) for file %q.\nFor more information see %s",
			fileExtension, formatExtension, formatName, file, docsFormatsUrl(formatName),
		)
	}

	return nil
}

type ProjectLocales interface {
	ProjectIds() []string
}

func LocalesForProjects(client *phraseapp.Client, projectLocales ProjectLocales) (map[string][]*phraseapp.Locale, error) {
	projectIdToLocales := map[string][]*phraseapp.Locale{}
	for _, pid := range projectLocales.ProjectIds() {
		if _, ok := projectIdToLocales[pid]; !ok {
			remoteLocales, err := RemoteLocales(client, pid)
			if err != nil {
				return nil, err
			}

			projectIdToLocales[pid] = remoteLocales
		}
	}
	return projectIdToLocales, nil
}

func RemoteLocales(client *phraseapp.Client, projectId string) ([]*phraseapp.Locale, error) {
	page := 1
	locales, err := client.LocalesList(projectId, page, 25)
	if err != nil {
		return nil, err
	}
	result := locales
	for len(locales) == 25 {
		page = page + 1
		locales, err = client.LocalesList(projectId, page, 25)
		if err != nil {
			return nil, err
		}
		result = append(result, locales...)
	}
	return result, nil
}

func GetFormats(client *phraseapp.Client) (map[string]*phraseapp.Format, error) {
	formats, err := client.FormatsList(1, 25)
	if err != nil {
		return nil, err
	}
	formatMap := map[string]*phraseapp.Format{}
	for _, format := range formats {
		formatMap[format.ApiName] = format
	}
	return formatMap, nil
}

func Exists(absPath string) error {
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("no such file or directory: %s", absPath)
	}
	return nil
}

func isDir(path string) bool {
	stat, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func createFile(path string) error {
	err := Exists(path)
	if err != nil {
		absDir := filepath.Dir(path)
		err := Exists(absDir)
		if err != nil {
			os.MkdirAll(absDir, 0700)
		}

		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	return nil
}
