package main

import (
	"fmt"
	"path/filepath"
	"testing"
)

func getBaseSource() *Source {
	source := &Source{
		File:        "./tests/<locale_code>.yml",
		ProjectID:   "project-id",
		AccessToken: "access-token",
		FileFormat:  "yml",
		Extension:   "",
		Params: &PushParams{
			FileFormat:         "",
			LocaleID:           "",
			ConvertEmoji:       nil,
			SkipUnverification: nil,
			SkipUploadTags:     nil,
			Tags:               nil,
			UpdateTranslations: nil,
		},
		RemoteLocales: getBaseLocales(),
	}
	source.Extension = filepath.Ext(source.File)
	return source
}

func TestSourceFields(t *testing.T) {
	fmt.Println("Source#Fields test")
	source := getBaseSource()

	if source.File != "./tests/<locale_code>.yml" {
		t.Errorf("Expected File to be %s and not %s", "./tests/<locale_code>.yml", source.File)
	}

	if source.AccessToken != "access-token" {
		t.Errorf("Expected AccesToken to be %s and not %s", "access-token", source.AccessToken)
	}

	if source.ProjectID != "project-id" {
		t.Errorf("Expected ProjectID to be %s and not %s", "project-id", source.ProjectID)
	}

	if source.FileFormat != "yml" {
		t.Errorf("Expected FileFormat to be %s and not %s", "yml", source.FileFormat)
	}

}

func TestSourceLocaleFilesOne(t *testing.T) {
	fmt.Println("Source#LocaleFiles#1 test")
	source := getBaseSource()
	localeFiles, err := source.LocaleFiles()

	if err != nil {
		t.Errorf("Should not fail with: %s", err.Error())
	}

	absPath, _ := filepath.Abs("./tests/en.yml")
	expectedFiles := []*LocaleFile{
		&LocaleFile{
			Name: "",
			RFC:  "en",
			ID:   "",
			Path: absPath,
		},
	}

	if len(localeFiles) == len(expectedFiles) {
		if err = compareLocaleFiles(localeFiles, expectedFiles); err != nil {
			t.Errorf(err.Error())
		}
	} else {
		t.Errorf(".LocaleFiles should contain %s and not %s", expectedFiles, localeFiles)
	}
}

func TestSourceLocaleFilesTwo(t *testing.T) {
	fmt.Println("Source#LocaleFiles#2 test")
	source := getBaseSource()
	source.File = "./**/<locale_name>.yml"

	localeFiles, err := source.LocaleFiles()

	if err != nil {
		t.Errorf("Should not fail with: %s", err.Error())
	}

	absPath, _ := filepath.Abs("./tests/en.yml")
	expectedFiles := []*LocaleFile{
		&LocaleFile{
			Name: "en",
			RFC:  "",
			ID:   "",
			Path: absPath,
		},
	}

	if len(localeFiles) == len(expectedFiles) {
		if err = compareLocaleFiles(localeFiles, expectedFiles); err != nil {
			t.Errorf(err.Error())
		}
	} else {
		t.Errorf("LocaleFiles should contain %s and not %s", expectedFiles, localeFiles)
	}
}

type Pattern struct {
	File         string
	Ext          string
	TestPath     string
	ExpectedRFC  string
	ExpectedName string
	ExpectedTag  string
}

func TestParserPatterns(t *testing.T) {
	fmt.Println("Parser pattern test")
	for _, pattern := range []*Pattern{
		&Pattern{
			File:        "./locales/<locale_code>.yml",
			Ext:         "yml",
			TestPath:    "locales/en.yml",
			ExpectedRFC: "en",
		},
		&Pattern{
			File:        "./config/<tag>/<locale_code>.yml",
			Ext:         "yml",
			TestPath:    "config/abc/en.yml",
			ExpectedRFC: "en",
			ExpectedTag: "abc",
		},
		&Pattern{
			File:         "./config/<locale_name>/<locale_code>.yml",
			Ext:          "yml",
			TestPath:     "config/german/de.yml",
			ExpectedRFC:  "de",
			ExpectedTag:  "",
			ExpectedName: "german",
		},
		&Pattern{
			File:        "./config/<locale_code>/*.yml",
			Ext:         "yml",
			TestPath:    "config/en/english.yml",
			ExpectedRFC: "en",
		},
		&Pattern{
			File:        "./<tag>/<locale_code>.lproj/Localizable.strings",
			Ext:         "strings",
			TestPath:    "no_tag/abc/en.lproj/Localizable.strings",
			ExpectedRFC: "en",
			ExpectedTag: "abc",
		},
		&Pattern{
			File:        "./<locale_code>.lproj/<tag>.strings",
			Ext:         "strings",
			TestPath:    "abc/en.lproj/MyStoryboard.strings",
			ExpectedRFC: "en",
			ExpectedTag: "MyStoryboard",
		},
		&Pattern{
			File:        "./<tag>/<locale_code>-values/Strings.xml",
			Ext:         "xml",
			TestPath:    "no_tag/abc/en-values/Strings.xml",
			ExpectedRFC: "en",
			ExpectedTag: "abc",
		},
		&Pattern{
			File:        "./<tag>/play.<locale_code>",
			Ext:         "<locale_code>",
			TestPath:    "no_tag/abc/play.en",
			ExpectedRFC: "en",
			ExpectedTag: "abc",
		},
		&Pattern{
			File:         "./<tag>/<locale_name>.<locale_code>",
			Ext:          "<locale_code>",
			TestPath:     "no_tag/abc/play.en",
			ExpectedRFC:  "en",
			ExpectedTag:  "abc",
			ExpectedName: "play",
		},
		&Pattern{
			File:         "./<locale_code>/<tag>.<locale_name>",
			Ext:          "strings",
			TestPath:     "abc/defg/hijk/en/MyStoryboard.english",
			ExpectedRFC:  "en",
			ExpectedTag:  "MyStoryboard",
			ExpectedName: "english",
		},
	} {
		parser := &Parser{
			SourceFile: pattern.File,
			Extension:  pattern.Ext,
		}
		parser.Initialize()
		parser.Search()

		if !parser.MatchesPath(pattern.TestPath) {
			printPattern(pattern, parser)
			t.Fail()
		}

		localeFile, err := parser.Eval(pattern.TestPath)
		if err != nil {
			fmt.Println(fmt.Sprintf("\tRFC:'%s', Name:'%s', Tag:'%s'", localeFile.RFC, localeFile.Name, localeFile.Tag))
			t.Errorf(err.Error())
			t.Fail()
		}

		if localeFile.RFC != pattern.ExpectedRFC {
			t.Errorf("Expected RFC to equal '%s' but was '%s'", pattern.ExpectedRFC, localeFile.RFC)
			t.Fail()
		}

		if localeFile.Tag != pattern.ExpectedTag {
			t.Errorf("Expected Tag to equal '%s' but was '%s'", pattern.ExpectedTag, localeFile.Tag)
			t.Fail()
		}

		if localeFile.Name != pattern.ExpectedName {
			t.Errorf("Expected LocaleName to equal '%s' but was '%s'", pattern.ExpectedName, localeFile.Name)
			t.Fail()
		}
	}
}

func printPattern(pattern *Pattern, parser *Parser) {
	fmt.Println("Pattern:", pattern.File, " -> ", parser.Matcher)
	fmt.Println("File:", pattern.TestPath)
}

func TestReplacePlaceholderInParams(t *testing.T) {
	fmt.Println("Source#ReplacePlaceholderInParams test")
	source := getBaseSource()
	source.Params.LocaleID = "<locale_code>"
	localeFile := &LocaleFile{
		Name: "en",
		RFC:  "en",
		ID:   "",
		Path: "",
	}
	s := source.replacePlaceholderInParams(localeFile)
	if s != "en" {
		t.Errorf("Expected LocaleId to equal '%s' but was '%s'", "en", s)
		t.Fail()
	}
}
func TestGetRemoteLocaleForLocaleFile(t *testing.T) {
	fmt.Println("Source#getRemoteLocaleForLocaleFile test")
	source := getBaseSource()
	localeFile := &LocaleFile{
		Name: "english",
		RFC:  "en",
		ID:   "",
		Path: "",
	}
	locale := source.getRemoteLocaleForLocaleFile(localeFile)
	if locale.Name != localeFile.Name {
		t.Errorf("Expected LocaleName to equal '%s' but was '%s'", "ennglish", localeFile.Name)
		t.Fail()
	}
	if locale.Code != localeFile.RFC {
		t.Errorf("Expected LocaleId to equal '%s' but was '%s'", "en", localeFile.RFC)
		t.Fail()
	}
}

func TestGenerateLocaleForFile(t *testing.T) {
	fmt.Println("Source#generateLocaleForFile test")
	source := getBaseSource()

	validPathOnFileSystem := "abc/defg"
	localeFile := &LocaleFile{
		Name: "english",
		RFC:  "en",
		ID:   "",
		Path: "",
	}
	newLocaleFile, err := source.generateLocaleForFile(localeFile, validPathOnFileSystem)
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}

	if newLocaleFile.ExistsRemote == false {
		t.Errorf("Expected that Locale exists remote but was '%b'", newLocaleFile.ExistsRemote)
		t.Fail()
	}

	if newLocaleFile.RFC != "en" {
		t.Errorf("Expected LocaleCode to equal '%s' but was '%s'", "en", newLocaleFile.RFC)
		t.Fail()
	}

	if newLocaleFile.Name != "english" {
		t.Errorf("Expected LocaleName to equal '%s' but was '%s'", "english", newLocaleFile.Name)
		t.Fail()
	}

	if newLocaleFile.ID != "en-locale-id" {
		t.Errorf("Expected LocaleId to equal '%s' but was '%s'", "en-locale-id", newLocaleFile.ID)
		t.Fail()
	}
}
